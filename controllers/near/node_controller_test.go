package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	nearClients "github.com/kotalco/kotal/clients/near"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("NEAR node controller", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "near",
		},
	}

	key := types.NamespacedName{
		Name:      "near-node",
		Namespace: ns.Name,
	}

	testImage := "kotalco/nearcore:controller-test"

	spec := nearv1alpha1.NodeSpec{
		Image:                    testImage,
		Network:                  "mainnet",
		RPC:                      true,
		Archive:                  true, // test volume storage size
		NodePrivateKeySecretName: "my-node-key",
		ValidatorSecretName:      "validator-key",
	}

	toCreate := &nearv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: spec,
	}

	client := nearClients.NewClient(toCreate)

	t := true

	nodeOwnerReference := metav1.OwnerReference{
		APIVersion:         "near.kotal.io/v1alpha1",
		Kind:               "Node",
		Name:               toCreate.Name,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}

	It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	It("should create NEAR node", func() {
		if os.Getenv(shared.EnvUseExistingCluster) != "true" {
			toCreate.Default()
		}
		Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	})

	It("Should get NEAR node", func() {
		fetched := &nearv1alpha1.Node{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec).To(Equal(toCreate.Spec))
		nodeOwnerReference.UID = fetched.UID
		time.Sleep(5 * time.Second)
	})

	It("Should create node statefulset", func() {
		fetched := &appsv1.StatefulSet{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(*fetched.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
			"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
			"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
			"RunAsNonRoot": gstruct.PointTo(Equal(true)),
		}))
		// init near node
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Name).To(Equal("init-near-node"))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Image).To(Equal(testImage))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Command).To(Equal([]string{"/bin/sh"}))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Args).To(ContainElements(
			fmt.Sprintf("%s/init_near_node.sh", shared.PathConfig(client.HomeDir())),
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Env).To(ContainElements(
			corev1.EnvVar{
				Name:  shared.EnvDataPath,
				Value: shared.PathData(client.HomeDir()),
			},
			corev1.EnvVar{
				Name:  envNetwork,
				Value: toCreate.Spec.Network,
			},
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: shared.PathData(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "config",
				MountPath: shared.PathConfig(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "secrets",
				MountPath: shared.PathSecrets(client.HomeDir()),
			},
		))
		// copy node key
		Expect(fetched.Spec.Template.Spec.InitContainers[1].Name).To(Equal("copy-node-key"))
		Expect(fetched.Spec.Template.Spec.InitContainers[1].Image).To(Equal(shared.BusyboxImage))
		Expect(fetched.Spec.Template.Spec.InitContainers[1].Command).To(Equal([]string{"/bin/sh"}))
		Expect(fetched.Spec.Template.Spec.InitContainers[1].Args).To(ContainElements(
			fmt.Sprintf("%s/copy_node_key.sh", shared.PathConfig(client.HomeDir())),
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[1].Env).To(ContainElements(
			corev1.EnvVar{
				Name:  shared.EnvDataPath,
				Value: shared.PathData(client.HomeDir()),
			},
			corev1.EnvVar{
				Name:  shared.EnvSecretsPath,
				Value: shared.PathSecrets(client.HomeDir()),
			},
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[1].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: shared.PathData(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "config",
				MountPath: shared.PathConfig(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "secrets",
				MountPath: shared.PathSecrets(client.HomeDir()),
			},
		))
		// copy validator key
		Expect(fetched.Spec.Template.Spec.InitContainers[2].Name).To(Equal("copy-validator-key"))
		Expect(fetched.Spec.Template.Spec.InitContainers[2].Image).To(Equal(shared.BusyboxImage))
		Expect(fetched.Spec.Template.Spec.InitContainers[2].Command).To(Equal([]string{"/bin/sh"}))
		Expect(fetched.Spec.Template.Spec.InitContainers[2].Args).To(ContainElements(
			fmt.Sprintf("%s/copy_validator_key.sh", shared.PathConfig(client.HomeDir())),
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[2].Env).To(ContainElements(
			corev1.EnvVar{
				Name:  shared.EnvDataPath,
				Value: shared.PathData(client.HomeDir()),
			},
			corev1.EnvVar{
				Name:  shared.EnvSecretsPath,
				Value: shared.PathSecrets(client.HomeDir()),
			},
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[2].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: shared.PathData(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "config",
				MountPath: shared.PathConfig(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "secrets",
				MountPath: shared.PathSecrets(client.HomeDir()),
			},
		))
		// node
		Expect(fetched.Spec.Template.Spec.Containers[0].Name).To(Equal("node"))
		Expect(fetched.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
		Expect(fetched.Spec.Template.Spec.Containers[0].Command).To(Equal(client.Command()))
		Expect(fetched.Spec.Template.Spec.Containers[0].Args).To(Equal(client.Args()))
		Expect(fetched.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: shared.PathData(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "config",
				MountPath: shared.PathConfig(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "secrets",
				MountPath: shared.PathSecrets(client.HomeDir()),
			},
		))
		// volumes
		mode := corev1.ConfigMapVolumeSourceDefaultMode
		Expect(fetched.Spec.Template.Spec.Volumes).To(ContainElements(
			[]corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: toCreate.Name,
						},
					},
				},
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: toCreate.Name,
							},
							DefaultMode: &mode,
						},
					},
				},
				{
					Name: "secrets",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: toCreate.Spec.NodePrivateKeySecretName,
										},
										Items: []corev1.KeyToPath{
											{
												Key:  "key",
												Path: "node_key.json",
											},
										},
									},
								},
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: toCreate.Spec.ValidatorSecretName,
										},
										Items: []corev1.KeyToPath{
											{
												Key:  "key",
												Path: "validator_key.json",
											},
										},
									},
								},
							},
							DefaultMode: &mode,
						},
					},
				},
			},
		))
	})

	It("Should create allocate correct resources to peer statefulset", func() {
		fetched := &appsv1.StatefulSet{}
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(nearv1alpha1.DefaultNodeCPURequest),
				corev1.ResourceMemory: resource.MustParse(nearv1alpha1.DefaultNodeMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(nearv1alpha1.DefaultNodeCPULimit),
				corev1.ResourceMemory: resource.MustParse(nearv1alpha1.DefaultNodeMemoryLimit),
			},
		}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
	})

	It("Should create node configmap", func() {
		fetched := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(fetched.Data).To(HaveKey("init_near_node.sh"))
		Expect(fetched.Data).To(HaveKey("copy_node_key.sh"))

	})

	It("Should create node data persistent volume with correct resources", func() {
		fetched := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(nearv1alpha1.DefaultArchivalNodeStorageRequest),
			},
		}
		Expect(fetched.Spec.Resources).To(Equal(expectedResources))
	})

	It("Should create node service", func() {
		fetched := &corev1.Service{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(fetched.Spec.Ports).To(ContainElements(
			[]corev1.ServicePort{
				{
					Name:       "p2p",
					Port:       int32(nearv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(nearv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "discovery",
					Port:       int32(nearv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(nearv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "prometheus",
					Port:       int32(nearv1alpha1.DefaultPrometheusPort),
					TargetPort: intstr.FromInt(int(nearv1alpha1.DefaultPrometheusPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "rpc",
					Port:       int32(nearv1alpha1.DefaultRPCPort),
					TargetPort: intstr.FromInt(int(nearv1alpha1.DefaultRPCPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		))
	})

	It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
		Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
	})

})
