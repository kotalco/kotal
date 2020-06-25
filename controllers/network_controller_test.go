package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	ethereumv1alpha1 "github.com/mfarghaly/kotal/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Ethereum network controller", func() {

	const (
		sleepTime  = 5 * time.Second
		interval   = 2 * time.Second
		timeout    = 60 * time.Second
		privatekey = ethereumv1alpha1.PrivateKey("0x608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e")
	)

	var (
		useExistingCluster = os.Getenv("USE_EXISTING_CLUSTER") == "true"
	)

	Context("Joining Rinkeby", func() {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rinkeby",
			},
		}
		key := types.NamespacedName{
			Name:      "my-network",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NetworkSpec{
			Join: "rinkeby",
			Nodes: []ethereumv1alpha1.Node{
				{
					Name:     "node-1",
					Bootnode: true,
					Nodekey:  privatekey,
				},
			},
		}

		toCreate := &ethereumv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		ownerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Network",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}
		bootnodeKey := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-1"),
			Namespace: key.Namespace,
		}
		node2Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-2"),
			Namespace: key.Namespace,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create the network", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should Get the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			Expect(fetched.Status.NodesCount).To(Equal(len(toCreate.Spec.Nodes)))
			ownerReference.UID = fetched.GetUID()
		})

		It("Should not create genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-genesis", key.Name),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).ToNot(Succeed())
		})

		It("Should create bootnode privatekey secret with correct data", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).To(Succeed())
			Expect(nodeSecret.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(string(nodeSecret.Data["nodekey"])).To(Equal(string(privatekey)[2:]))
		})

		It("Should create bootnode service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should create bootnode deployment with correct arguments and resources", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(DefaultPublicNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(DefaultPublicNetworkNodeMemoryRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgNetwork,
				"rinkeby",
				ArgDataPath,
				ArgNodePrivateKey,
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))

		})

		It("Should create bootnode data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				RPCPort: 8547,
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		It("Should create node-2 deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgNetwork,
				"rinkeby",
				ArgDataPath,
				ArgBootnodes,
				ArgRPCHTTPEnabled,
				"8547",
			}))
		})

		It("Should create node-2 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should not create privatekey secret for node-2 (without nodekey)", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSecret)).ToNot(Succeed())
		})

		It("Should not create bootnode service (not a bootnode)", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
		})

		It("Should update the network by removing node-2", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			fetched.Spec.Nodes = fetched.Spec.Nodes[:1]
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		if useExistingCluster {
			It("Should delete node-2 deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})
		}

		It("Should delete network", func() {
			toDelete := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get network after deletion", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete bootnode deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).ToNot(Succeed())
			})

			It("Should delete bootnode data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), bootnodeKey, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete bootnode privatekey secret", func() {
				nodeSecret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).ToNot(Succeed())
			})

			It("Should delete bootnode service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).ToNot(Succeed())
			})
		}

		It(fmt.Sprintf("should delete %s namespace", ns.Name), func() {
			Expect(k8sClient.Delete(context.Background(), ns)).Should(Succeed())
		})
	})

	Context("private PoA network", func() {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "poa",
			},
		}
		key := types.NamespacedName{
			Name:      "my-poa-network",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NetworkSpec{
			Consensus: ethereumv1alpha1.ProofOfAuthority,
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID: 55555,
				Clique: &ethereumv1alpha1.Clique{
					Signers: []ethereumv1alpha1.EthereumAddress{
						ethereumv1alpha1.EthereumAddress("0xd2c21213027cbf4d46c16b55fa98e5252b048706"),
					},
				},
			},
			Nodes: []ethereumv1alpha1.Node{
				{
					Name:     "node-1",
					Bootnode: true,
					Nodekey:  privatekey,
				},
			},
		}

		toCreate := &ethereumv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		ownerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Network",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}
		bootnodeKey := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-1"),
			Namespace: key.Namespace,
		}
		node2Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-2"),
			Namespace: key.Namespace,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create the network", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should Get the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			Expect(fetched.Status.NodesCount).To(Equal(len(toCreate.Spec.Nodes)))
			ownerReference.UID = fetched.GetUID()
		})

		It("Should create genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-genesis", key.Name),
				Namespace: key.Namespace,
			}
			expectedExtraData := "0x0000000000000000000000000000000000000000000000000000000000000000d2c21213027cbf4d46c16b55fa98e5252b0487060000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
		})

		It("Should create bootnode privatekey secret with correct data", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).To(Succeed())
			Expect(nodeSecret.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(string(nodeSecret.Data["nodekey"])).To(Equal(string(privatekey)[2:]))
		})

		It("Should create bootnode service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should create bootnode deployment with correct arguments and resources", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(DefaultPrivateNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(DefaultPrivateNetworkNodeMemoryRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgDataPath,
				ArgNodePrivateKey,
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				RPCPort: 8547,
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		It("Should create node-2 deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgDataPath,
				ArgBootnodes,
				ArgRPCHTTPEnabled,
				"8547",
			}))
		})

		It("Should create node-2 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should not create privatekey secret for node-2 (without nodekey)", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSecret)).ToNot(Succeed())
		})

		It("Should not create bootnode service (not a bootnode)", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
		})

		It("Should update the network by removing node-2", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			fetched.Spec.Nodes = fetched.Spec.Nodes[:1]
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		if useExistingCluster {
			It("Should delete node-2 deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})
		}

		It("Should delete network", func() {
			toDelete := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get network after deletion", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete bootnode deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).ToNot(Succeed())
			})

			It("Should delete bootnode data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), bootnodeKey, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete bootnode privatekey secret", func() {
				nodeSecret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).ToNot(Succeed())
			})

			It("Should delete bootnode service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-genesis", key.Name),
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
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pow",
			},
		}
		key := types.NamespacedName{
			Name:      "my-pow-network",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NetworkSpec{
			Consensus: ethereumv1alpha1.ProofOfWork,
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID: 55555,
				Ethash: &ethereumv1alpha1.Ethash{
					FixedDifficulty: 1500,
				},
			},
			Nodes: []ethereumv1alpha1.Node{
				{
					Name:     "node-1",
					Bootnode: true,
					Nodekey:  privatekey,
				},
			},
		}

		toCreate := &ethereumv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		ownerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Network",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}
		bootnodeKey := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-1"),
			Namespace: key.Namespace,
		}
		node2Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-2"),
			Namespace: key.Namespace,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create the network", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should Get the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			Expect(fetched.Status.NodesCount).To(Equal(len(toCreate.Spec.Nodes)))
			ownerReference.UID = fetched.GetUID()
		})

		It("Should create genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-genesis", key.Name),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
		})

		It("Should create bootnode privatekey secret with correct data", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).To(Succeed())
			Expect(nodeSecret.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(string(nodeSecret.Data["nodekey"])).To(Equal(string(privatekey)[2:]))
		})

		It("Should create bootnode service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should create bootnode deployment with correct arguments and resources", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(DefaultPrivateNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(DefaultPrivateNetworkNodeMemoryRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgDataPath,
				ArgNodePrivateKey,
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				RPCPort: 8547,
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		It("Should create node-2 deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgDataPath,
				ArgBootnodes,
				ArgRPCHTTPEnabled,
				"8547",
			}))
		})

		It("Should create node-2 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should not create privatekey secret for node-2 (without nodekey)", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSecret)).ToNot(Succeed())
		})

		It("Should not create bootnode service (not a bootnode)", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
		})

		It("Should update the network by removing node-2", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			fetched.Spec.Nodes = fetched.Spec.Nodes[:1]
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		if useExistingCluster {
			It("Should delete node-2 deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})
		}

		It("Should delete network", func() {
			toDelete := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get network after deletion", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete bootnode deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).ToNot(Succeed())
			})

			It("Should delete bootnode data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), bootnodeKey, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete bootnode privatekey secret", func() {
				nodeSecret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).ToNot(Succeed())
			})

			It("Should delete bootnode service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-genesis", key.Name),
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
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ibft2",
			},
		}
		key := types.NamespacedName{
			Name:      "my-ibft2-network",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NetworkSpec{
			Consensus: ethereumv1alpha1.IstanbulBFT,
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID: 55555,
				IBFT2: &ethereumv1alpha1.IBFT2{
					Validators: []ethereumv1alpha1.EthereumAddress{
						"0x427e2c7cecd72bc4cdd4f7ebb8bb6e49789c8044",
						"0xd2c21213027cbf4d46c16b55fa98e5252b048706",
						"0x8e1f6c7c76a1d7f74eda342d330ca9749f31cc2b",
					},
				},
			},
			Nodes: []ethereumv1alpha1.Node{
				{
					Name:     "node-1",
					Bootnode: true,
					Nodekey:  privatekey,
				},
			},
		}

		toCreate := &ethereumv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}
		t := true
		ownerReference := metav1.OwnerReference{
			// TODO: update version
			APIVersion:         "ethereum.kotal.io/v1alpha1",
			Kind:               "Network",
			Name:               toCreate.Name,
			Controller:         &t,
			BlockOwnerDeletion: &t,
		}
		bootnodeKey := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-1"),
			Namespace: key.Namespace,
		}
		node2Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-2"),
			Namespace: key.Namespace,
		}

		It(fmt.Sprintf("should create %s namespace", ns.Name), func() {
			Expect(k8sClient.Create(context.Background(), ns)).Should(Succeed())
		})

		It("Should create the network", func() {
			if !useExistingCluster {
				toCreate.Default()
			}
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should Get the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched.Spec).To(Equal(toCreate.Spec))
			Expect(fetched.Status.NodesCount).To(Equal(len(toCreate.Spec.Nodes)))
			ownerReference.UID = fetched.GetUID()
		})

		It("Should create genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-genesis", key.Name),
				Namespace: key.Namespace,
			}
			expectedExtraData := "0xf869a00000000000000000000000000000000000000000000000000000000000000000f83f94427e2c7cecd72bc4cdd4f7ebb8bb6e49789c804494d2c21213027cbf4d46c16b55fa98e5252b048706948e1f6c7c76a1d7f74eda342d330ca9749f31cc2b808400000000c0"
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
		})

		It("Should create bootnode privatekey secret with correct data", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).To(Succeed())
			Expect(nodeSecret.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(string(nodeSecret.Data["nodekey"])).To(Equal(string(privatekey)[2:]))
		})

		It("Should create bootnode service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should create bootnode deployment with correct arguments and resources", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(DefaultPrivateNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(DefaultPrivateNetworkNodeMemoryRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgDataPath,
				ArgNodePrivateKey,
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				RPCPort: 8547,
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		It("Should create node-2 deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ArgDataPath,
				ArgBootnodes,
				ArgRPCHTTPEnabled,
				"8547",
			}))
		})

		It("Should create node-2 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
		})

		It("Should not create privatekey secret for node-2 (without nodekey)", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSecret)).ToNot(Succeed())
		})

		It("Should not create bootnode service (not a bootnode)", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
		})

		It("Should update the network by removing node-2", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			fetched.Spec.Nodes = fetched.Spec.Nodes[:1]
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		if useExistingCluster {
			It("Should delete node-2 deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})
		}

		It("Should delete network", func() {
			toDelete := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, toDelete)).To(Succeed())
			Expect(k8sClient.Delete(context.Background(), toDelete)).To(Succeed())
			time.Sleep(sleepTime)
		})

		It("Should not get network after deletion", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).ToNot(Succeed())
		})

		if useExistingCluster {
			It("Should delete bootnode deployment", func() {
				nodeDep := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).ToNot(Succeed())
			})

			It("Should delete bootnode data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), bootnodeKey, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete bootnode privatekey secret", func() {
				nodeSecret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSecret)).ToNot(Succeed())
			})

			It("Should delete bootnode service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-genesis", key.Name),
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
