package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
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

var _ = Describe("aptos node controller", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "aptos",
		},
	}

	key := types.NamespacedName{
		Name:      "aptos-node",
		Namespace: ns.Name,
	}

	testImage := "kotalco/aptos:controller-test"

	spec := aptosv1alpha1.NodeSpec{
		Image:       testImage,
		Network:     "testnet",
		API:         true,
		APIPort:     4321,
		P2PPort:     5432,
		MetricsPort: 9099,
	}

	toCreate := &aptosv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: spec,
	}

	t := true

	nodeOwnerReference := metav1.OwnerReference{
		APIVersion:         "aptos.kotal.io/v1alpha1",
		Kind:               "Node",
		Name:               toCreate.Name,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}

	It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	It("should create aptos node", func() {
		if os.Getenv(shared.EnvUseExistingCluster) != "true" {
			toCreate.Default()
		}
		Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	})

	It("Should get aptos node", func() {
		fetched := &aptosv1alpha1.Node{}
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
		Expect(fetched.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
	})

	It("Should create allocate correct resources to node statefulset", func() {
		fetched := &appsv1.StatefulSet{}
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(aptosv1alpha1.DefaultNodeCPURequest),
				corev1.ResourceMemory: resource.MustParse(aptosv1alpha1.DefaultNodeMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(aptosv1alpha1.DefaultNodeCPULimit),
				corev1.ResourceMemory: resource.MustParse(aptosv1alpha1.DefaultNodeMemoryLimit),
			},
		}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
	})

	It("Should create node configmap", func() {
		fetched := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(fetched.Data).To(HaveKey("config.yaml"))

	})

	It("Should create node data persistent volume with correct resources", func() {
		fetched := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(aptosv1alpha1.DefaultNodeStorageRequest),
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
					Port:       int32(spec.P2PPort),
					TargetPort: intstr.FromInt(int(spec.P2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "metrics",
					Port:       int32(spec.MetricsPort),
					TargetPort: intstr.FromInt(int(spec.MetricsPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "api",
					Port:       int32(spec.APIPort),
					TargetPort: intstr.FromInt(int(spec.APIPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		))
	})

	It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
		Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
	})

})
