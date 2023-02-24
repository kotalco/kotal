package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	stacksClients "github.com/kotalco/kotal/clients/stacks"
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

var _ = Describe("Stacks node controller", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "stacks",
		},
	}

	key := types.NamespacedName{
		Name:      "stacks-node",
		Namespace: ns.Name,
	}

	testImage := "kotalco/stacks:controller-test"

	spec := stacksv1alpha1.NodeSpec{
		Image:   testImage,
		Network: stacksv1alpha1.Mainnet,
		BitcoinNode: stacksv1alpha1.BitcoinNode{
			Endpoint:              "bitcoin.blockstack.com",
			P2pPort:               8332,
			RpcPort:               8333,
			RpcUsername:           "blockstack",
			RpcPasswordSecretName: "bitcoin-node-rpc-password",
		},
	}

	toCreate := &stacksv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: spec,
	}

	client := stacksClients.NewClient(toCreate)

	t := true

	nodeOwnerReference := metav1.OwnerReference{
		APIVersion:         "stacks.kotal.io/v1alpha1",
		Kind:               "Node",
		Name:               toCreate.Name,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}

	It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	It("Should create Bitcoin node rpc password secret", func() {
		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bitcoin-node-rpc-password",
				Namespace: ns.Name,
			},
			StringData: map[string]string{
				"password": "blockstacksystem",
			},
		}
		Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())
	})

	It("should create Stacks node", func() {
		if os.Getenv(shared.EnvUseExistingCluster) != "true" {
			toCreate.Default()
		}
		Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	})

	It("Should get Stacks node", func() {
		fetched := &stacksv1alpha1.Node{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec).To(Equal(toCreate.Spec))
		nodeOwnerReference.UID = fetched.UID
		time.Sleep(5 * time.Second)
	})

	It("Should create peer configmap", func() {
		fetched := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(fetched.Data).To(HaveKey("config.toml"))
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
		Expect(fetched.Spec.Template.Spec.Containers[0].Name).To(Equal("node"))
		Expect(fetched.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
		Expect(fetched.Spec.Template.Spec.Containers[0].Env).To(Equal(client.Env()))
		Expect(fetched.Spec.Template.Spec.Containers[0].Command).To(Equal(client.Command()))
		Expect(fetched.Spec.Template.Spec.Containers[0].Args).To(Equal(client.Args()))
		Expect(fetched.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: shared.PathData(client.HomeDir()),
			},
			corev1.VolumeMount{
				Name:      "config",
				ReadOnly:  true,
				MountPath: shared.PathConfig(client.HomeDir()),
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
			},
		))
	})

	It("Should create allocate correct resources to node statefulset", func() {
		fetched := &appsv1.StatefulSet{}
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(stacksv1alpha1.DefaultNodeCPURequest),
				corev1.ResourceMemory: resource.MustParse(stacksv1alpha1.DefaultNodeMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(stacksv1alpha1.DefaultNodeCPULimit),
				corev1.ResourceMemory: resource.MustParse(stacksv1alpha1.DefaultNodeMemoryLimit),
			},
		}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
	})

	It("Should create node data persistent volume with correct resources", func() {
		fetched := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(stacksv1alpha1.DefaultNodeStorageRequest),
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
					Port:       int32(stacksv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(stacksv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "rpc",
					Port:       int32(stacksv1alpha1.DefaultRPCPort),
					TargetPort: intstr.FromInt(int(stacksv1alpha1.DefaultRPCPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		))
	})

	It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
		Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
	})

})
