package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	ipfsClients "github.com/kotalco/kotal/clients/ipfs"
	"github.com/kotalco/kotal/controllers/shared"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("IPFS cluster peer controller", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ipfs-cluster-peer",
		},
	}

	key := types.NamespacedName{
		Name:      "my-cluster-peer",
		Namespace: ns.Name,
	}

	image := "kotalco/ipfs-cluster:test"

	spec := ipfsv1alpha1.ClusterPeerSpec{
		Image:                image,
		ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
		PrivateKeySecretName: "cluster-privatekey",
		ClusterSecretName:    "cluster-secret",
		Consensus:            ipfsv1alpha1.CRDT,
		TrustedPeers: []string{
			"12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
			"12D3KooWQ9yZnqowEDme3gSgS45KY9ZoEmAiGYRxusdEaqtFa9pr",
		},
		PeerEndpoint: "/dns4/ipfs-peer/tcp/5001",
	}

	toCreate := &ipfsv1alpha1.ClusterPeer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: spec,
	}

	t := true

	peerOwnerReference := metav1.OwnerReference{
		APIVersion:         "ipfs.kotal.io/v1alpha1",
		Kind:               "ClusterPeer",
		Name:               toCreate.Name,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}

	It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	It("should create cluster privatekey secret", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-privatekey",
				Namespace: ns.Name,
			},
			StringData: map[string]string{
				"key": "CAESQOH/DvUJmeJ9z6m3wAStpkrlBwJQxIyNSK0YGf0EI5ZRGpwsWxl4wmgReqmHl8LQjTC2iPM0QbYAjeY3Z63AFnI=",
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
	})

	It("should create cluster secret", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-secret",
				Namespace: ns.Name,
			},
			StringData: map[string]string{
				"secret": "clu$ter$3cr3t",
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
	})

	It("should create ipfs cluster peer", func() {
		if os.Getenv(shared.EnvUseExistingCluster) != "true" {
			toCreate.Default()
		}
		Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	})

	It("should get ipfs cluster peer", func() {
		fetched := &ipfsv1alpha1.ClusterPeer{}
		Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
		Expect(fetched.Spec).To(Equal(toCreate.Spec))
		peerOwnerReference.UID = fetched.UID
		time.Sleep(5 * time.Second)
	})

	It("Should create ipfs cluster peer statefulset with correct image and arguments", func() {
		fetched := &appsv1.StatefulSet{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(peerOwnerReference))
		Expect(*fetched.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
			"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
			"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
			"RunAsNonRoot": gstruct.PointTo(Equal(true)),
		}))
		container := fetched.Spec.Template.Spec.Containers[0]
		Expect(container.Image).To(Equal(image))

	})

	It("Should pass correct environment variables to ipfs cluster peer statefulset containers", func() {
		fetched := &appsv1.StatefulSet{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())

		container := fetched.Spec.Template.Spec.Containers[0]
		Expect(container.Env).To(ContainElements([]corev1.EnvVar{
			{
				Name:  ipfsClients.EnvIPFSClusterPath,
				Value: shared.PathData(ipfsClients.GoIPFSClusterHomeDir),
			},
			{
				Name:  ipfsClients.EnvIPFSClusterPeerName,
				Value: toCreate.Name,
			},
		}))

		initContainer := fetched.Spec.Template.Spec.InitContainers[0]
		Expect(initContainer.Env).To(ContainElements([]corev1.EnvVar{
			{
				Name:  ipfsClients.EnvIPFSClusterPath,
				Value: shared.PathData(ipfsClients.GoIPFSClusterHomeDir),
			},
			{
				Name:  ipfsClients.EnvIPFSClusterConsensus,
				Value: string(toCreate.Spec.Consensus),
			},
			{
				Name:  ipfsClients.EnvIPFSClusterPeerEndpoint,
				Value: toCreate.Spec.PeerEndpoint,
			},
			{
				Name: ipfsClients.EnvIPFSClusterSecret,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: toCreate.Spec.ClusterSecretName,
						},
						Key: "secret",
					},
				},
			},
			{
				Name:  ipfsClients.EnvIPFSClusterTrustedPeers,
				Value: strings.Join(toCreate.Spec.TrustedPeers, ","),
			},
			{
				Name:  ipfsClients.EnvIPFSClusterId,
				Value: toCreate.Spec.ID,
			},
			{
				Name: ipfsClients.EnvIPFSClusterPrivateKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: toCreate.Spec.PrivateKeySecretName,
						},
						Key: "key",
					},
				},
			},
		}))
	})

	It("Should create allocate correct resources to cluster peer statefulset", func() {
		fetched := &appsv1.StatefulSet{}
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(ipfsv1alpha1.DefaultNodeCPURequest),
				corev1.ResourceMemory: resource.MustParse(ipfsv1alpha1.DefaultNodeMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(ipfsv1alpha1.DefaultNodeCPULimit),
				corev1.ResourceMemory: resource.MustParse(ipfsv1alpha1.DefaultNodeMemoryLimit),
			},
		}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
	})

	It("Should create peer configmap", func() {
		fetched := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(peerOwnerReference))
		Expect(fetched.Data).To(HaveKey("init_ipfs_cluster_config.sh"))
	})

	It("Should create peer data persistent volume with correct resources", func() {
		fetched := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(peerOwnerReference))

		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(ipfsv1alpha1.DefaultNodeStorageRequest),
			},
		}
		Expect(fetched.Spec.Resources).To(Equal(expectedResources))
	})

	It("Should create peer service", func() {
		fetched := &corev1.Service{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(peerOwnerReference))
		Expect(fetched.Spec.Ports).To(ContainElements(
			[]corev1.ServicePort{
				{
					Name:       "swarm",
					Port:       9096,
					TargetPort: intstr.FromInt(9096),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "swarm-udp",
					Port:       9096,
					TargetPort: intstr.FromInt(9096),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					// Pinning service API
					// https://ipfscluster.io/documentation/reference/pinsvc_api/
					Name:       "api",
					Port:       5001,
					TargetPort: intstr.FromInt(int(5001)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					// Proxy API
					// https://ipfscluster.io/documentation/reference/proxy/
					Name:       "proxy-api",
					Port:       9095,
					TargetPort: intstr.FromInt(int(9095)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					// REST API
					//https://ipfscluster.io/documentation/reference/api/
					Name:       "rest",
					Port:       9094,
					TargetPort: intstr.FromInt(int(9094)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "metrics",
					Port:       8888,
					TargetPort: intstr.FromInt(int(8888)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "tracing",
					Port:       6831,
					TargetPort: intstr.FromInt(int(6831)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		))
	})

	It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
		Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
	})

})
