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
	ethereum2Clients "github.com/kotalco/kotal/clients/ethereum2"
	"github.com/kotalco/kotal/controllers/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("Ethereum 2.0 validator client", func() {

	Context("Teku validator client", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "teku",
			},
		}

		key := types.NamespacedName{
			Name:      "teku-validator",
			Namespace: ns.Name,
		}

		testImage := "kotalco/teku:test"

		spec := ethereum2v1alpha1.ValidatorSpec{
			Image:           testImage,
			Network:         "mainnet",
			Client:          ethereum2v1alpha1.TekuClient,
			BeaconEndpoints: []string{"http://10.96.130.88:9999"},
			Graffiti:        "testing Kotal validator controller",
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

		It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.TODO(), ns))
		})

		It("Should create validator client", func() {
			if os.Getenv(shared.EnvUseExistingCluster) != "true" {
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

		It("Should create statefulset", func() {
			validatorSts := &appsv1.StatefulSet{}

			Expect(k8sClient.Get(context.Background(), key, validatorSts)).To(Succeed())
			Expect(validatorSts.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(*validatorSts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
			// container volume mounts
			Expect(validatorSts.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.TekuHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.TekuHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.TekuHomeDir), "my-validator"),
				},
			))
			// container volume
			mode := corev1.ConfigMapVolumeSourceDefaultMode
			Expect(validatorSts.Spec.Template.Spec.Volumes).To(ContainElements(
				corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: validatorSts.Name,
						},
					},
				},
				corev1.Volume{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: validatorSts.Name},
							DefaultMode:          &mode,
						},
					},
				},
				corev1.Volume{
					Name: "my-validator",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "my-validator",
							Items: []corev1.KeyToPath{
								{
									Key:  "keystore",
									Path: "keystore-0.json",
								},
								{
									Key:  "password",
									Path: "password.txt",
								},
							},
							DefaultMode: &mode,
						},
					},
				},
			))
			// teku doesn't require init containers
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

		It("Should create validator configmap", func() {
			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, configmap)).To(Succeed())
			Expect(configmap.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
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

		It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
		})

	})

	Context("Prysm validator client", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "prysm",
			},
		}

		key := types.NamespacedName{
			Name:      "prysm-validator",
			Namespace: ns.Name,
		}

		testImage := "kotalco/prysm:test"

		spec := ethereum2v1alpha1.ValidatorSpec{
			Image:                testImage,
			Network:              "mainnet",
			Client:               ethereum2v1alpha1.PrysmClient,
			BeaconEndpoints:      []string{"http://10.96.130.88:9999"},
			Graffiti:             "testing Kotal validator controller",
			WalletPasswordSecret: "my-wallet-password",
			Keystores: []ethereum2v1alpha1.Keystore{
				{
					SecretName: "my-validator",
				},
			},
			CertSecretName: "my-cert",
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

		It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.TODO(), ns))
		})

		It("Should create validator client", func() {
			if os.Getenv(shared.EnvUseExistingCluster) != "true" {
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

		It("Should create statefulset", func() {
			validatorSts := &appsv1.StatefulSet{}

			Expect(k8sClient.Get(context.Background(), key, validatorSts)).To(Succeed())
			Expect(validatorSts.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(*validatorSts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
			// container volume mounts
			Expect(validatorSts.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.PrysmHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.PrysmHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.PrysmHomeDir), "my-validator"),
				},
				corev1.VolumeMount{
					Name:      "my-wallet-password",
					ReadOnly:  true,
					MountPath: fmt.Sprintf("%s/prysm-wallet", shared.PathSecrets(ethereum2Clients.PrysmHomeDir)),
				},
				corev1.VolumeMount{
					Name:      "cert",
					ReadOnly:  true,
					MountPath: fmt.Sprintf("%s/cert", shared.PathSecrets(ethereum2Clients.PrysmHomeDir)),
				},
			))
			// container volume
			mode := corev1.ConfigMapVolumeSourceDefaultMode
			Expect(validatorSts.Spec.Template.Spec.Volumes).To(ContainElements(
				corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: validatorSts.Name,
						},
					},
				},
				corev1.Volume{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: validatorSts.Name},
							DefaultMode:          &mode,
						},
					},
				},
				corev1.Volume{
					Name: "my-validator",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "my-validator",
							Items: []corev1.KeyToPath{
								{
									Key:  "keystore",
									Path: "keystore-0.json",
								},
								{
									Key:  "password",
									Path: "password.txt",
								},
							},
							DefaultMode: &mode,
						},
					},
				},
				corev1.Volume{
					Name: "my-wallet-password",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "my-wallet-password",
							Items: []corev1.KeyToPath{
								{
									Key:  "password",
									Path: "prysm-wallet-password.txt",
								},
							},
							DefaultMode: &mode,
						},
					},
				},
				corev1.Volume{
					Name: "cert",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  toCreate.Spec.CertSecretName,
							DefaultMode: &mode,
						},
					},
				},
			))
			// init containers
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(testImage))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Env).To(ContainElements(
				corev1.EnvVar{
					Name:  envNetwork,
					Value: "mainnet",
				},
				corev1.EnvVar{
					Name:  shared.EnvDataPath,
					Value: shared.PathData(ethereum2Clients.PrysmHomeDir),
				},
				corev1.EnvVar{
					Name:  envKeyDir,
					Value: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.PrysmHomeDir), "my-validator"),
				},
				corev1.EnvVar{
					Name:  envKeystoreIndex,
					Value: "0",
				},
				corev1.EnvVar{
					Name:  shared.EnvSecretsPath,
					Value: shared.PathSecrets(ethereum2Clients.PrysmHomeDir),
				},
			))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Command).To(ConsistOf("/bin/sh"))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Args).To(ConsistOf(
				fmt.Sprintf("%s/prysm_import_keystore.sh", shared.PathConfig(ethereum2Clients.PrysmHomeDir))),
			)
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.PrysmHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.PrysmHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.PrysmHomeDir), "my-validator"),
				},
				corev1.VolumeMount{
					Name:      "my-wallet-password",
					ReadOnly:  true,
					MountPath: fmt.Sprintf("%s/prysm-wallet", shared.PathSecrets(ethereum2Clients.PrysmHomeDir)),
				},
			))

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

		It("Should create validator configmap", func() {
			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, configmap)).To(Succeed())
			Expect(configmap.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(configmap.Data).To(HaveKey("prysm_import_keystore.sh"))
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

		It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
		})

	})

	Context("Lighthouse validator client", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lighthouse",
			},
		}

		key := types.NamespacedName{
			Name:      "lighthouse-validator",
			Namespace: ns.Name,
		}

		testImage := "kotalco/lighthouse:test"

		spec := ethereum2v1alpha1.ValidatorSpec{
			Image:                testImage,
			Network:              "mainnet",
			Client:               ethereum2v1alpha1.LighthouseClient,
			BeaconEndpoints:      []string{"http://10.96.130.88:9999"},
			Graffiti:             "testing Kotal validator controller",
			WalletPasswordSecret: "my-wallet-password",
			Keystores: []ethereum2v1alpha1.Keystore{
				{
					SecretName: "my-validator",
					PublicKey:  "0x83dbb18e088cb16a07fca598db2ac24da3e8549601eedd75eb28d8a9d4be405f49f7dbdcad5c9d7df54a8a40a143e852",
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

		It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.TODO(), ns))
		})

		It("Should create validator client", func() {
			if os.Getenv(shared.EnvUseExistingCluster) != "true" {
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

		It("Should create statefulset", func() {
			validatorSts := &appsv1.StatefulSet{}

			Expect(k8sClient.Get(context.Background(), key, validatorSts)).To(Succeed())
			Expect(validatorSts.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(*validatorSts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
			// container volume mounts
			Expect(validatorSts.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.LighthouseHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.LighthouseHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.LighthouseHomeDir), "my-validator"),
				},
			))
			// container volume
			mode := corev1.ConfigMapVolumeSourceDefaultMode
			Expect(validatorSts.Spec.Template.Spec.Volumes).To(ContainElements(
				corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: validatorSts.Name,
						},
					},
				},
				corev1.Volume{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: validatorSts.Name},
							DefaultMode:          &mode,
						},
					},
				},
				corev1.Volume{
					Name: "my-validator",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "my-validator",
							Items: []corev1.KeyToPath{
								{
									Key:  "keystore",
									Path: "keystore-0.json",
								},
								{
									Key:  "password",
									Path: "password.txt",
								},
							},
							DefaultMode: &mode,
						},
					},
				},
			))
			// init containers
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(testImage))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Env).To(ContainElements(
				corev1.EnvVar{
					Name:  envNetwork,
					Value: "mainnet",
				},
				corev1.EnvVar{
					Name:  shared.EnvDataPath,
					Value: shared.PathData(ethereum2Clients.LighthouseHomeDir),
				},
				corev1.EnvVar{
					Name:  envKeyDir,
					Value: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.LighthouseHomeDir), "my-validator"),
				},
				corev1.EnvVar{
					Name:  envKeystoreIndex,
					Value: "0",
				},
			))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Command).To(ConsistOf("/bin/sh"))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Args).To(ConsistOf(
				fmt.Sprintf("%s/lighthouse_import_keystore.sh", shared.PathConfig(ethereum2Clients.LighthouseHomeDir))),
			)
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.LighthouseHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.LighthouseHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.LighthouseHomeDir), "my-validator"),
				},
			))

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

		It("Should create validator configmap", func() {
			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, configmap)).To(Succeed())
			Expect(configmap.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(configmap.Data).To(HaveKey("lighthouse_import_keystore.sh"))
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

		It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
		})

	})

	Context("Nimbus validator client", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nimbus",
			},
		}

		key := types.NamespacedName{
			Name:      "nimbus-validator",
			Namespace: ns.Name,
		}

		testImage := "kotalco/nimbus:test"

		spec := ethereum2v1alpha1.ValidatorSpec{
			Image:                testImage,
			Network:              "mainnet",
			Client:               ethereum2v1alpha1.NimbusClient,
			BeaconEndpoints:      []string{"http://10.96.130.88:9999"},
			Graffiti:             "testing Kotal validator controller",
			WalletPasswordSecret: "my-wallet-password",
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

		It(fmt.Sprintf("Should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.TODO(), ns))
		})

		It("Should create validator client", func() {
			if os.Getenv(shared.EnvUseExistingCluster) != "true" {
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

		It("Should create statefulset", func() {
			validatorSts := &appsv1.StatefulSet{}

			Expect(k8sClient.Get(context.Background(), key, validatorSts)).To(Succeed())
			Expect(validatorSts.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(*validatorSts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(validatorSts.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))
			// container volume mounts
			Expect(validatorSts.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.NimbusHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.NimbusHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.NimbusHomeDir), "my-validator"),
				},
				corev1.VolumeMount{
					Name:      "validator-secrets",
					MountPath: fmt.Sprintf("%s/validator-secrets", shared.PathSecrets(ethereum2Clients.NimbusHomeDir)),
				},
			))
			// container volume
			mode := corev1.ConfigMapVolumeSourceDefaultMode
			fmt.Sprintln(validatorSts.Spec.Template.Spec.Volumes)
			Expect(validatorSts.Spec.Template.Spec.Volumes).To(ContainElements(
				corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: validatorSts.Name,
						},
					},
				},
				corev1.Volume{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: validatorSts.Name},
							DefaultMode:          &mode,
						},
					},
				},
				corev1.Volume{
					Name: "my-validator",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "my-validator",
							Items: []corev1.KeyToPath{
								{
									Key:  "keystore",
									Path: "keystore.json",
								},
							},
							DefaultMode: &mode,
						},
					},
				},
				corev1.Volume{
					Name: "validator-secrets",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "my-validator",
										},
										Items: []corev1.KeyToPath{
											{
												Key:  "password",
												Path: "my-validator",
											},
										},
									},
								},
							},
							DefaultMode: &mode,
						},
					},
				},
			))
			// init containers
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(testImage))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Env).To(ContainElements(
				corev1.EnvVar{
					Name:  shared.EnvSecretsPath,
					Value: shared.PathSecrets(ethereum2Clients.NimbusHomeDir),
				},
				corev1.EnvVar{
					Name:  envValidatorsPath,
					Value: fmt.Sprintf("%s/kotal-validators", shared.PathData(ethereum2Clients.NimbusHomeDir)),
				},
			))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Command).To(ConsistOf("/bin/sh"))
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].Args).To(ConsistOf(
				fmt.Sprintf("%s/nimbus_copy_validators.sh", shared.PathConfig(ethereum2Clients.NimbusHomeDir))),
			)
			Expect(validatorSts.Spec.Template.Spec.InitContainers[0].VolumeMounts).To(ContainElements(
				corev1.VolumeMount{
					Name:      "data",
					MountPath: shared.PathData(ethereum2Clients.NimbusHomeDir),
				},
				corev1.VolumeMount{
					Name:      "config",
					MountPath: shared.PathConfig(ethereum2Clients.NimbusHomeDir),
				},
				corev1.VolumeMount{
					Name:      "my-validator",
					MountPath: fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(ethereum2Clients.NimbusHomeDir), "my-validator"),
				},
			))

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

		It("Should create validator configmap", func() {
			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, configmap)).To(Succeed())
			Expect(configmap.GetOwnerReferences()).To(ContainElement(validatorOwnerReference))
			Expect(configmap.Data).To(HaveKey("nimbus_copy_validators.sh"))
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

		It(fmt.Sprintf("Should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).To(Succeed())
		})

	})

})
