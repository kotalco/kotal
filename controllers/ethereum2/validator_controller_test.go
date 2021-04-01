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

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 validator client", func() {

	Context("Teku validator client", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "validator-client",
			},
		}

		key := types.NamespacedName{
			Name:      "teku-validator",
			Namespace: ns.Name,
		}

		spec := ethereum2v1alpha1.ValidatorSpec{
			Network:        "mainnet",
			Client:         ethereum2v1alpha1.TekuClient,
			BeaconEndpoint: "http://10.96.130.88:9999",
			Graffiti:       "testing Kotal validator controller",
			Keystores: []ethereum2v1alpha1.Keystore{
				{
					SecretName: "my-validator",
				},
			},
		}

		toCreate := &ethereum2v1alpha1.Validator{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}

		t := true

		validatorOwnerReference := metav1.OwnerReference{
			APIVersion:         "ethereum2.kotal.io/v1alpha1",
			Kind:               "Validator",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		client, _ := NewValidatorClient(ethereum2v1alpha1.TekuClient)

		It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.TODO(), ns))
		})

		It("Should create validator client", func() {
			if os.Getenv("USE_EXISTING_CLUSTER") != "true" {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
		})

		It("should get validator client", func() {
			fetched := &ethereum2v1alpha1.Validator{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			validatorOwnerReference.UID = fetched.GetUID()
			time.Sleep(5 * time.Second)
		})

		It("Should create statefulset with correct arguments", func() {
			validatorSts := &appsv1.StatefulSet{}
			secretsDir := PathSecrets(client.HomeDir())

			Expect(k8sClient.Get(context.Background(), key, validatorSts)).To(Succeed())
			Expect(validatorSts.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Image).To(Equal(client.Image()))
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				TekuVC,
				TekuDataPath,
				PathBlockchainData(client.HomeDir()),
				TekuNetwork,
				"mainnet",
				TekuValidatorsKeystoreLockingEnabled,
				"false",
				TekuBeaconNodeEndpoint,
				"http://10.96.130.88:9999",
				TekuGraffiti,
				"testing Kotal validator controller",
				TekuValidatorKeys,
				fmt.Sprintf("%s/validator-keys/my-validator/keystore-0.json:%s/validator-keys/my-validator/password.txt", secretsDir, secretsDir),
			}))
		})

		It("Should allocate correct resources to validator statefulset", func() {
			validatorSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), key, validatorSts)).To(Succeed())
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create data persistent volume with correct resources", func() {
			validatorPVC := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereum2v1alpha1.DefaultStorage),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, validatorPVC)).To(Succeed())
			Expect(validatorPVC.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(validatorPVC.Spec.Resources).To(Equal(expectedResources))
		})

	})
})
