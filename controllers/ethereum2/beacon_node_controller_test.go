package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 beacon node", func() {

	Context("Joining Mainnet", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mainnet",
			},
		}

		key := types.NamespacedName{
			Name:      "my-node",
			Namespace: ns.Name,
		}

		spec := ethereum2v1alpha1.BeaconNodeSpec{
			Join: "mainnet",
		}

		toCreate := &ethereum2v1alpha1.BeaconNode{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}

		t := true

		nodeOwnerReference := metav1.OwnerReference{
			APIVersion:         "ethereum2.kotal.io/v1alpha1",
			Kind:               "BeaconNode",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.TODO(), ns))
		})

		It("Should create beacon node", func() {
			if os.Getenv("USE_EXISTING_CLUSTER") != "true" {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
		})

		It("should get beacon node", func() {
			fetched := &ethereum2v1alpha1.BeaconNode{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			nodeOwnerReference.UID = fetched.GetUID()
			time.Sleep(5 * time.Second)
		})

		It("Should create statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			client, err := NewBeaconNodeClient(toCreate.Spec.Client)
			Expect(err).To(BeNil())
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(client.Image()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
			}))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereum2v1alpha1.DefaultCPURequest),
					corev1.ResourceMemory: resource.MustParse(ethereum2v1alpha1.DefaultMemoryRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereum2v1alpha1.DefaultCPULimit),
					corev1.ResourceMemory: resource.MustParse(ethereum2v1alpha1.DefaultMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create data persistent volume with correct resources", func() {
			nodePVC := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereum2v1alpha1.DefaultStorage),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should create node service", func() {
			nodeSVC := &corev1.Service{}
			Expect(k8sClient.Get(context.Background(), key, nodeSVC)).To(Succeed())
			Expect(nodeSVC.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodeSVC.Spec.Ports).To(ContainElements([]corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       int32(ethereum2v1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereum2v1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       int32(ethereum2v1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereum2v1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

	})
})
