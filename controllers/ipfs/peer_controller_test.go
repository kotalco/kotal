package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("IPFS peer controller", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ipfs-peer",
		},
	}

	key := types.NamespacedName{
		Name:      "my-peer",
		Namespace: ns.Name,
	}

	image := "kotalco/kubo:test"

	spec := ipfsv1alpha1.PeerSpec{
		Image:              image,
		API:                true,
		APIPort:            3333,
		Gateway:            true,
		GatewayPort:        4444,
		Routing:            ipfsv1alpha1.DHTClientRouting,
		SwarmKeySecretName: "my-swarm",
	}

	toCreate := &ipfsv1alpha1.Peer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: spec,
	}

	t := true

	peerOwnerReference := metav1.OwnerReference{
		APIVersion:         "ipfs.kotal.io/v1alpha1",
		Kind:               "Peer",
		Name:               toCreate.Name,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}

	It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	It("should create ipfs peer", func() {
		if os.Getenv(shared.EnvUseExistingCluster) != "true" {
			toCreate.Default()
		}
		Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	})

	It("Should get ipfs peer", func() {
		fetched := &ipfsv1alpha1.Peer{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec).To(Equal(toCreate.Spec))
		peerOwnerReference.UID = fetched.UID
		time.Sleep(5 * time.Second)
	})

	It("Should create peer statefulset with correct arguments", func() {
		fetched := &appsv1.StatefulSet{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(peerOwnerReference))
		Expect(*fetched.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
			"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
			"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
			"RunAsNonRoot": gstruct.PointTo(Equal(true)),
		}))
		Expect(fetched.Spec.Template.Spec.Containers[0].Image).To(Equal(image))
	})

	It("Should create allocate correct resources to peer statefulset", func() {
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
		Expect(fetched.Data).To(HaveKey("init_ipfs_config.sh"))
		Expect(fetched.Data).To(HaveKey("config_ipfs.sh"))
		Expect(fetched.Data).To(HaveKey("copy_swarm_key.sh"))
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
					Port:       4001,
					TargetPort: intstr.FromInt(4001),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "swarm-udp",
					Port:       4001,
					TargetPort: intstr.FromInt(4001),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "api",
					Port:       int32(3333),
					TargetPort: intstr.FromInt(int(3333)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "gateway",
					Port:       int32(4444),
					TargetPort: intstr.FromInt(int(4444)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		))
	})

	It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
		Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
	})

})
