package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	chainlinkClients "github.com/kotalco/kotal/clients/chainlink"
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

var _ = Describe("Chainlink node controller", func() {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "chainlink",
		},
	}

	key := types.NamespacedName{
		Name:      "chainlink-node",
		Namespace: ns.Name,
	}

	testImage := "kotalco/chainlink:controller-test"

	spec := chainlinkv1alpha1.NodeSpec{
		Image:                      testImage,
		EthereumChainId:            1,
		EthereumWSEndpoint:         "wss://my-eth-node:8546",
		LinkContractAddress:        "0x01BE23585060835E02B77ef475b0Cc51aA1e0709",
		DatabaseURL:                "postgresql://postgres:password@postgres:5432/postgres",
		KeystorePasswordSecretName: "keystore-password",
		API:                        true,
		APICredentials: chainlinkv1alpha1.APICredentials{
			Email:              "mostafa@kotal.co",
			PasswordSecretName: "api-password",
		},
	}

	toCreate := &chainlinkv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: spec,
	}

	client := chainlinkClients.NewClient(toCreate)

	t := true

	nodeOwnerReference := metav1.OwnerReference{
		APIVersion:         "chainlink.kotal.io/v1alpha1",
		Kind:               "Node",
		Name:               toCreate.Name,
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}

	It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	It("should create chainlink node", func() {
		if os.Getenv(shared.EnvUseExistingCluster) != "true" {
			toCreate.Default()
		}
		Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	})

	It("Should get chainlink node", func() {
		fetched := &chainlinkv1alpha1.Node{}
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
		// init container
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Image).To(Equal(shared.BusyboxImage))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Command).To(ConsistOf("/bin/sh"))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Args).To(ConsistOf(
			fmt.Sprintf("%s/copy_api_credentials.sh", shared.PathConfig(client.HomeDir())),
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].Env).To(ContainElements(
			corev1.EnvVar{
				Name:  shared.EnvDataPath,
				Value: shared.PathData(client.HomeDir()),
			},
			corev1.EnvVar{
				Name:  envApiEmail,
				Value: toCreate.Spec.APICredentials.Email,
			},
			corev1.EnvVar{
				Name:  shared.EnvSecretsPath,
				Value: shared.PathSecrets(client.HomeDir()),
			},
		))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: client.HomeDir(),
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
		// node container
		Expect(fetched.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
		Expect(fetched.Spec.Template.Spec.Containers[0].Command).To(Equal(client.Command()))
		Expect(fetched.Spec.Template.Spec.Containers[0].Args).To(Equal(client.Args()))
		Expect(fetched.Spec.Template.Spec.Containers[0].Env).To(Equal(client.Env()))
		Expect(fetched.Spec.Template.Spec.InitContainers[0].VolumeMounts).To(ContainElements(
			corev1.VolumeMount{
				Name:      "data",
				MountPath: client.HomeDir(),
			},
			corev1.VolumeMount{
				Name:      "secrets",
				MountPath: shared.PathSecrets(client.HomeDir()),
			},
		))
	})

	It("Should create allocate correct resources to peer statefulset", func() {
		fetched := &appsv1.StatefulSet{}
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(chainlinkv1alpha1.DefaultNodeCPURequest),
				corev1.ResourceMemory: resource.MustParse(chainlinkv1alpha1.DefaultNodeMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(chainlinkv1alpha1.DefaultNodeCPULimit),
				corev1.ResourceMemory: resource.MustParse(chainlinkv1alpha1.DefaultNodeMemoryLimit),
			},
		}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
	})

	It("Should create node configmap", func() {
		fetched := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(fetched.Data).To(HaveKey("copy_api_credentials.sh"))

	})

	It("Should create node service", func() {
		fetched := &corev1.Service{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		Expect(fetched.Spec.Ports).To(ContainElements(
			[]corev1.ServicePort{
				{
					Name:       "p2p",
					Port:       int32(toCreate.Spec.P2PPort),
					TargetPort: intstr.FromInt(int(toCreate.Spec.P2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "api",
					Port:       int32(toCreate.Spec.APIPort),
					TargetPort: intstr.FromInt(int(toCreate.Spec.APIPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		))

	})

	It("Should create node data persistent volume with correct resources", func() {
		fetched := &corev1.PersistentVolumeClaim{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.OwnerReferences).To(ContainElements(nodeOwnerReference))
		expectedResources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(chainlinkv1alpha1.DefaultNodeStorageRequest),
			},
		}
		Expect(fetched.Spec.Resources).To(Equal(expectedResources))
	})

	It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
		Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
	})

})
