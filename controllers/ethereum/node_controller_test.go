package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
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

var _ = Describe("Ethereum network controller", func() {

	const (
		sleepTime = 5 * time.Second
		interval  = 2 * time.Second
		timeout   = 2 * time.Minute
		networkID = 7777
		// node private key
		privatekey = "608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e"
		// imported account
		accountKey      = "5df5eff7ef9e4e82739b68a34c6b23608d79ee8daf3b598a01ffb0dd7aa3a2fd"
		accountAddress  = sharedAPI.EthereumAddress("0x2b3430337f12Ce89EaBC7b0d865F4253c7744c0d")
		accountPassword = "secret"
	)

	var (
		useExistingCluster = os.Getenv(shared.EnvUseExistingCluster) == "true"
	)

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

		spec := ethereumv1alpha1.NodeSpec{
			Client:                   ethereumv1alpha1.BesuClient,
			Network:                  "mainnet",
			NodePrivateKeySecretName: "nodekey",
			SyncMode:                 ethereumv1alpha1.FullSynchronization,
			Logging:                  sharedAPI.NoLogs,
		}

		toCreate := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true

		nodeOwnerReference := metav1.OwnerReference{
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Node",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create nodekey secret", func() {
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": privatekey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())
		})

		It("Should create the node", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should get the node", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			// TODO: test status
			nodeOwnerReference.UID = fetched.GetUID()
		})

		It("Should create node configmap", func() {
			config := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, config)).Should(Succeed())
			Expect(config.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
		})

		It("Should create node service", func() {
			svc := &corev1.Service{}
			Expect(k8sClient.Get(context.Background(), key, svc)).To(Succeed())
			Expect(svc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(svc.Spec.Ports).To(ContainElements([]corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

		It("Should create node statefulset with correct arguments", func() {
			sts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), key, sts)).To(Succeed())
			Expect(sts.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(*sts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(sts.Spec.Template.Spec.Containers[0].Image).To(Equal(toCreate.Spec.Image))
		})

		It("Should allocate correct resources to node statefulset", func() {
			sts := &appsv1.StatefulSet{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPURequest),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPULimit),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, sts)).To(Succeed())
			Expect(sts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node data persistent volume with correct resources", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultMainNetworkFullNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, pvc)).To(Succeed())
			Expect(pvc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(pvc.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should delete the node", func() {
			toDelete := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get the node after deletion", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete node statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &corev1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node service", func() {
				nodeSvc := &corev1.Service{}
				Expect(k8sClient.Get(context.Background(), key, nodeSvc)).ToNot(Succeed())
			})
		}

		It(fmt.Sprintf("should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).Should(Succeed())
		})
	})

	Context("Joining Goerli", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ethereumv1alpha1.GoerliNetwork,
			},
		}
		key := types.NamespacedName{
			Name:      "my-node",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NodeSpec{
			Client:                   ethereumv1alpha1.BesuClient,
			Network:                  ethereumv1alpha1.GoerliNetwork,
			NodePrivateKeySecretName: "nodekey",
			Logging:                  sharedAPI.FatalLogs,
		}

		toCreate := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		nodeOwnerReference := metav1.OwnerReference{
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Node",
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create nodekey secret", func() {
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": privatekey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())
		})

		It("Should create account private key and password secrets", func() {
			accountPrivateKeySecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-account-privatekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": accountKey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &accountPrivateKeySecret)).To(Succeed())

			accountPasswordSecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-account-password",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"password": accountPassword,
				},
			}
			Expect(k8sClient.Create(context.Background(), &accountPasswordSecret)).To(Succeed())
		})

		It("Should create the node", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should get the node", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			nodeOwnerReference.UID = fetched.GetUID()
			nodeOwnerReference.Name = key.Name
		})

		It("Should create node config", func() {
			config := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, config)).Should(Succeed())
		})

		It("Should create node service", func() {
			svc := &corev1.Service{}
			Expect(k8sClient.Get(context.Background(), key, svc)).To(Succeed())
			Expect(svc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(svc.Spec.Ports).To(ContainElements([]corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

		It("Should create node statefulset with correct arguments", func() {
			sts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), key, sts)).To(Succeed())
			Expect(sts.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(*sts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(sts.Spec.Template.Spec.Containers[0].Image).To(Equal(toCreate.Spec.Image))
		})

		It("Should allocate correct resources to node statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPURequest),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPULimit),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))

		})

		It("Should create bootnode data persistent volume with correct resources", func() {
			nodePVC := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultTestNetworkStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should delete node", func() {
			toDelete := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get node after deletion", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete node statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &corev1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			// TODO: remove this test
			It("Should delete node privatekey secret", func() {
				nodeSecret := &corev1.Secret{}
				Expect(k8sClient.Get(context.Background(), key, nodeSecret)).ToNot(Succeed())
			})

			It("Should delete node service", func() {
				nodeSvc := &corev1.Service{}
				Expect(k8sClient.Get(context.Background(), key, nodeSvc)).ToNot(Succeed())
			})
		}

		It(fmt.Sprintf("should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).Should(Succeed())
		})
	})

	Context("private PoA network", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "poa",
			},
		}
		key := types.NamespacedName{
			Name:      "my-poa-node",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NodeSpec{
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID:   55555,
				NetworkID: networkID,
				Clique: &ethereumv1alpha1.Clique{
					Signers: []sharedAPI.EthereumAddress{
						sharedAPI.EthereumAddress("0xd2c21213027cbf4d46c16b55fa98e5252b048706"),
					},
				},
			},
			Client:                   ethereumv1alpha1.BesuClient,
			NodePrivateKeySecretName: "nodekey",
		}

		toCreate := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		nodeOwnerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Node",
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create nodekey secret", func() {
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": privatekey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())
		})

		It("Should create account private key and password secrets", func() {
			accountPrivateKeySecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-account-privatekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": accountKey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &accountPrivateKeySecret)).To(Succeed())

			accountPasswordSecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-account-password",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"password": accountPassword,
				},
			}
			Expect(k8sClient.Create(context.Background(), &accountPasswordSecret)).To(Succeed())
		})

		It("Should create the node", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should get the node", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			nodeOwnerReference.UID = fetched.GetUID()
			nodeOwnerReference.Name = key.Name
		})

		It("Should create node service", func() {
			svc := &corev1.Service{}
			Expect(k8sClient.Get(context.Background(), key, svc)).To(Succeed())
			Expect(svc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(svc.Spec.Ports).To(ContainElements([]corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

		It("Should create node statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(*nodeSts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(toCreate.Spec.Image))
		})

		It("Should create node genesis block config", func() {
			genesisConfig := &corev1.ConfigMap{}
			expectedExtraData := "0x0000000000000000000000000000000000000000000000000000000000000000d2c21213027cbf4d46c16b55fa98e5252b0487060000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			Expect(k8sClient.Get(context.Background(), key, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
		})

		It("Should allocate correct resources to node statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPURequest),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume with correct resources", func() {
			nodePVC := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should delete node", func() {
			toDelete := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get node after deletion", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete node statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &corev1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node service", func() {
				nodeSvc := &corev1.Service{}
				Expect(k8sClient.Get(context.Background(), key, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete besu genesis block configmap", func() {
				genesisConfig := &corev1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-besu", key.Name),
					Namespace: key.Namespace,
				}
				Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).ToNot(Succeed())
			})
		}

		It(fmt.Sprintf("should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).Should(Succeed())
		})
	})

	Context("private PoW network", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pow",
			},
		}
		key := types.NamespacedName{
			Name:      "my-pow-node",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NodeSpec{
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID:   55555,
				NetworkID: networkID,
				Ethash:    &ethereumv1alpha1.Ethash{},
			},
			Client:                   ethereumv1alpha1.BesuClient,
			NodePrivateKeySecretName: "nodekey",
			Logging:                  sharedAPI.TraceLogs,
		}

		toCreate := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		nodeOwnerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Node",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create nodekey secret", func() {
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": privatekey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())
		})

		It("Should create account private key and password secrets", func() {
			accountPrivateKeySecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-account-privatekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": accountKey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &accountPrivateKeySecret)).To(Succeed())

			accountPasswordSecret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-account-password",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"password": accountPassword,
				},
			}
			Expect(k8sClient.Create(context.Background(), &accountPasswordSecret)).To(Succeed())
		})

		It("Should create the network", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should create the node", func() {
			node := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, node)).To(Succeed())
			Expect(node.Spec).To(Equal(toCreate.Spec))
			nodeOwnerReference.UID = node.GetUID()
			nodeOwnerReference.Name = key.Name
		})

		It("Should create node genesis block configmap", func() {
			config := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), key, config)).To(Succeed())
		})

		It("Should create node service", func() {
			svc := &corev1.Service{}
			Expect(k8sClient.Get(context.Background(), key, svc)).To(Succeed())
			Expect(svc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(svc.Spec.Ports).To(ContainElements([]corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

		It("Should create node statefulset with correct arguments", func() {
			sts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), key, sts)).To(Succeed())
			Expect(sts.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(*sts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(sts.Spec.Template.Spec.Containers[0].Image).To(Equal(toCreate.Spec.Image))
		})

		It("Should allocate correct resources to node statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPURequest),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node data persistent volume with correct resources", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, pvc)).To(Succeed())
			Expect(pvc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(pvc.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should delete node", func() {
			toDelete := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get node after deletion", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete node statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &corev1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node service", func() {
				nodeSvc := &corev1.Service{}
				Expect(k8sClient.Get(context.Background(), key, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete besu genesis block configmap", func() {
				genesisConfig := &corev1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-besu", key.Name),
					Namespace: key.Namespace,
				}
				Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).ToNot(Succeed())
			})
		}

		It(fmt.Sprintf("should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).Should(Succeed())
		})
	})

	Context("private ibft2 network", func() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ibft2",
			},
		}
		key := types.NamespacedName{
			Name:      "my-ibft2-node",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NodeSpec{
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID:   55555,
				NetworkID: networkID,
				IBFT2: &ethereumv1alpha1.IBFT2{
					Validators: []sharedAPI.EthereumAddress{
						"0x427e2c7cecd72bc4cdd4f7ebb8bb6e49789c8044",
						"0xd2c21213027cbf4d46c16b55fa98e5252b048706",
						"0x8e1f6c7c76a1d7f74eda342d330ca9749f31cc2b",
					},
				},
			},
			Client:                   ethereumv1alpha1.BesuClient,
			NodePrivateKeySecretName: "nodekey",
			Logging:                  sharedAPI.WarnLogs,
		}

		toCreate := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true

		nodeOwnerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Node",
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create nodekey secret", func() {
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodekey",
					Namespace: ns.Name,
				},
				StringData: map[string]string{
					"key": privatekey,
				},
			}
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())
		})

		It("Should create the node", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should get the node", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			nodeOwnerReference.UID = fetched.GetUID()
			nodeOwnerReference.Name = key.Name
		})

		It("Should create node genesis block configmap", func() {
			genesisConfig := &corev1.ConfigMap{}
			expectedExtraData := "0xf869a00000000000000000000000000000000000000000000000000000000000000000f83f94427e2c7cecd72bc4cdd4f7ebb8bb6e49789c804494d2c21213027cbf4d46c16b55fa98e5252b048706948e1f6c7c76a1d7f74eda342d330ca9749f31cc2b808400000000c0"
			Expect(k8sClient.Get(context.Background(), key, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
		})

		It("Should create node service", func() {
			nodeSvc := &corev1.Service{}
			Expect(k8sClient.Get(context.Background(), key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodeSvc.Spec.Ports).To(ContainElements([]corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       int32(ethereumv1alpha1.DefaultP2PPort),
					TargetPort: intstr.FromInt(int(ethereumv1alpha1.DefaultP2PPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

		It("Should create node statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(*nodeSts.Spec.Template.Spec.SecurityContext).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"RunAsUser":    gstruct.PointTo(Equal(int64(1000))),
				"RunAsGroup":   gstruct.PointTo(Equal(int64(3000))),
				"FSGroup":      gstruct.PointTo(Equal(int64(2000))),
				"RunAsNonRoot": gstruct.PointTo(Equal(true)),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(toCreate.Spec.Image))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPURequest),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					corev1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume with correct resouces", func() {
			nodePVC := &corev1.PersistentVolumeClaim{}
			expectedResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(nodeOwnerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should delete node", func() {
			toDelete := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get network after deletion", func() {
			fetched := &ethereumv1alpha1.Node{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete node statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &corev1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node service", func() {
				nodeSvc := &corev1.Service{}
				Expect(k8sClient.Get(context.Background(), key, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete genesis block configmap", func() {
				genesisConfig := &corev1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-besu", key.Name),
					Namespace: key.Namespace,
				}
				Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).ToNot(Succeed())
			})
		}

		It(fmt.Sprintf("should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).Should(Succeed())
		})
	})

})
