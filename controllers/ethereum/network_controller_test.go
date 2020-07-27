package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ethereumv1alpha1 "github.com/mfarghaly/kotal/apis/ethereum/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Ethereum network controller", func() {

	const (
		sleepTime  = 5 * time.Second
		interval   = 2 * time.Second
		timeout    = 2 * time.Minute
		networkID  = 7777
		privatekey = ethereumv1alpha1.PrivateKey("0x608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e")
		// imported account
		accountKey      = ethereumv1alpha1.PrivateKey("0x5df5eff7ef9e4e82739b68a34c6b23608d79ee8daf3b598a01ffb0dd7aa3a2fd")
		accountAddress  = ethereumv1alpha1.EthereumAddress("0x2b3430337f12Ce89EaBC7b0d865F4253c7744c0d")
		accountPassword = "secret"
	)

	var (
		useExistingCluster = os.Getenv("USE_EXISTING_CLUSTER") == "true"
	)

	if useExistingCluster {
		It("Should label all nodes with topology key", func() {
			nodes := &v1.NodeList{}
			Expect(k8sClient.List(context.Background(), nodes)).To(Succeed())
			for i, node := range nodes.Items {
				node.Labels[ethereumv1alpha1.DefaultTopologyKey] = fmt.Sprintf("zone-%d", i)
				Expect(k8sClient.Update(context.Background(), &node)).To(Succeed())
			}
		})
	}

	Context("Joining Mainnet", func() {
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mainnet",
			},
		}
		key := types.NamespacedName{
			Name:      "my-network",
			Namespace: ns.Name,
		}

		spec := ethereumv1alpha1.NetworkSpec{
			Join:            "mainnet",
			HighlyAvailable: true,
			Nodes: []ethereumv1alpha1.Node{
				{
					Name:     "node-1",
					Bootnode: true,
					Nodekey:  privatekey,
					SyncMode: ethereumv1alpha1.FullSynchronization,
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
		cpu := "250m"
		cpuLimit := "500m"
		memory := "1Gi"
		memoryLimit := "2Gi"

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

		It("Should create bootnode deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuNetwork,
				"mainnet",
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
			}))
		})

		It("Should allocate correct resources to bootnode deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryRequest),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPULimit),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultMainNetworkFullNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				Client:  ethereumv1alpha1.GethClient,
				RPCPort: 8547,
				Resources: &ethereumv1alpha1.NodeResources{
					CPU:         cpu,
					CPULimit:    cpuLimit,
					Memory:      memory,
					MemoryLimit: memoryLimit,
				},
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			if !useExistingCluster {
				fetched.Default()
			}
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		if useExistingCluster {
			It("Should schedule node-1 and node-2 on different nodes", func() {
				pods := &v1.PodList{}
				matchingLabels := client.MatchingLabels{"name": "node"}
				inNamespace := client.InNamespace(ns.Name)
				Expect(k8sClient.List(context.Background(), pods, matchingLabels, inNamespace)).To(Succeed())
				Expect(pods.Items[0].Spec.NodeName).NotTo(Equal(pods.Items[1].Spec.NodeName))
			})
		}

		It("Should create node-2 deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				GethDataDir,
				GethBootnodes,
				GethRPCHTTPEnabled,
				GethRPCHTTPPort,
				"8547",
				GethSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).ToNot(ContainElements([]string{
				ethereumv1alpha1.MainNetwork,
			}))
		})

		It("Should allocate correct resources to node-2 deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpu),
					v1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpuLimit),
					v1.ResourceMemory: resource.MustParse(memoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultMainNetworkFastNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
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
			Join:            "rinkeby",
			HighlyAvailable: true,
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
		cpu := "1"
		cpuLimit := "1500m"
		memory := "500Mi"
		memoryLimit := "1500Mi"

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

		It("Should create bootnode deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuNetwork,
				"rinkeby",
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
			}))
		})

		It("Should allocate correct resources to bootnode deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryRequest),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeCPULimit),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPublicNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))

		})

		It("Should create bootnode data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultTestNetworkStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:     "node-2",
				Client:   ethereumv1alpha1.GethClient,
				Miner:    true,
				Coinbase: accountAddress,
				SyncMode: ethereumv1alpha1.FullSynchronization,
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKey: accountKey,
					Password:   accountPassword,
				},
				Resources: &ethereumv1alpha1.NodeResources{
					CPU:         cpu,
					CPULimit:    cpuLimit,
					Memory:      memory,
					MemoryLimit: memoryLimit,
				},
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			if !useExistingCluster {
				fetched.Default()
			}
			Expect(k8sClient.Update(context.Background(), fetched)).To(Succeed())
			fetchedUpdated := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetchedUpdated)).To(Succeed())
			Expect(fetchedUpdated.Spec).To(Equal(fetched.Spec))
			time.Sleep(sleepTime)
		})

		if useExistingCluster {
			It("Should schedule node-1 and node-2 on different nodes", func() {
				pods := &v1.PodList{}
				matchingLabels := client.MatchingLabels{"name": "node"}
				inNamespace := client.InNamespace(ns.Name)
				Expect(k8sClient.List(context.Background(), pods, matchingLabels, inNamespace)).To(Succeed())
				Expect(pods.Items[0].Spec.NodeName).NotTo(Equal(pods.Items[1].Spec.NodeName))
			})
		}

		It("Should create node-2 deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage))
			Expect(nodeDep.Spec.Template.Spec.InitContainers[0].Image).To(Equal(GethImage))
			Expect(strings.Split(nodeDep.Spec.Template.Spec.InitContainers[0].Args[1], " ")).To(ContainElements([]string{
				"account",
				"import",
				GethDataDir,
				PathBlockchainData,
				GethPassword,
				fmt.Sprintf("%s/account.password", PathImportedAccount),
				fmt.Sprintf("%s/account.key", PathImportedAccount),
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				"--rinkeby",
				GethDataDir,
				GethBootnodes,
				GethMinerEnabled,
				GethMinerCoinbase,
				GethUnlock,
				GethPassword,
				GethSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
			}))
		})

		It("Should allocate correct resources to node-2 deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpu),
					v1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpuLimit),
					v1.ResourceMemory: resource.MustParse(memoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 imported account secret", func() {
			secret := &v1.Secret{}
			secretKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-%s-imported-account", toCreate.Name, "node-2"),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), secretKey, secret)).To(Succeed())
			Expect(string(secret.Data["account.key"])).To(Equal(string(accountKey)[2:]))
			Expect(string(secret.Data["account.password"])).To(Equal(accountPassword))
		})

		It("Should create node-2 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultTestNetworkStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
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
			ID:        networkID,
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

		cpu := "500m"
		cpuLimit := "750m"
		memory := "1500Mi"
		memoryLimit := "2500Mi"
		storage := "1234Mi"

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

		It("Should create bootnode deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
			}))
		})

		It("Should allocate correct resources to bootnode deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryRequest),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:     "node-2",
				Client:   ethereumv1alpha1.GethClient,
				Miner:    true,
				Coinbase: accountAddress,
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKey: accountKey,
					Password:   accountPassword,
				},
				SyncMode: ethereumv1alpha1.FastSynchronization,
				Resources: &ethereumv1alpha1.NodeResources{
					CPU:         cpu,
					CPULimit:    cpuLimit,
					Memory:      memory,
					MemoryLimit: memoryLimit,
					Storage:     storage,
				},
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			if !useExistingCluster {
				fetched.Default()
			}
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
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage))
			Expect(nodeDep.Spec.Template.Spec.InitContainers[0].Image).To(Equal(GethImage))
			Expect(strings.Split(nodeDep.Spec.Template.Spec.InitContainers[0].Args[1], " ")).To(ContainElements([]string{
				GethDataDir,
			}))
			Expect(nodeDep.Spec.Template.Spec.InitContainers[1].Image).To(Equal(GethImage))
			Expect(strings.Split(nodeDep.Spec.Template.Spec.InitContainers[1].Args[1], " ")).To(ContainElements([]string{
				"account",
				"import",
				GethDataDir,
				PathBlockchainData,
				GethPassword,
				fmt.Sprintf("%s/account.password", PathImportedAccount),
				fmt.Sprintf("%s/account.key", PathImportedAccount),
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				GethDataDir,
				GethBootnodes,
				GethSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
			}))
		})

		It("Should allocate correct resources to node-2 deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpu),
					v1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpuLimit),
					v1.ResourceMemory: resource.MustParse(memoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 imported account secret", func() {
			secret := &v1.Secret{}
			secretKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-%s-imported-account", toCreate.Name, "node-2"),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), secretKey, secret)).To(Succeed())
			Expect(string(secret.Data["account.key"])).To(Equal(string(accountKey)[2:]))
			Expect(string(secret.Data["account.password"])).To(Equal(accountPassword))
		})

		It("Should create node-2 data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(storage),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
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
			ID:        networkID,
			Consensus: ethereumv1alpha1.ProofOfWork,
			Genesis: &ethereumv1alpha1.Genesis{
				ChainID: 55555,
				Ethash:  &ethereumv1alpha1.Ethash{},
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
		cpu := "1"
		cpuLimit := "1500m"
		memory := "1500Mi"
		memoryLimit := "2500Mi"

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

		It("Should create bootnode deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
			}))
		})

		It("Should allocate correct resources to bootnode deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryRequest),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:     "node-2",
				Client:   ethereumv1alpha1.GethClient,
				Miner:    true,
				Coinbase: accountAddress,
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKey: accountKey,
					Password:   accountPassword,
				},
				SyncMode: ethereumv1alpha1.FastSynchronization,
				Resources: &ethereumv1alpha1.NodeResources{
					CPU:         cpu,
					CPULimit:    cpuLimit,
					Memory:      memory,
					MemoryLimit: memoryLimit,
				},
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			if !useExistingCluster {
				fetched.Default()
			}
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
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage))
			Expect(nodeDep.Spec.Template.Spec.InitContainers[0].Image).To(Equal(GethImage))
			Expect(strings.Split(nodeDep.Spec.Template.Spec.InitContainers[0].Args[1], " ")).To(ContainElements([]string{
				GethDataDir,
			}))
			Expect(nodeDep.Spec.Template.Spec.InitContainers[1].Image).To(Equal(GethImage))
			Expect(strings.Split(nodeDep.Spec.Template.Spec.InitContainers[1].Args[1], " ")).To(ContainElements([]string{
				"account",
				"import",
				GethDataDir,
				PathBlockchainData,
				GethPassword,
				fmt.Sprintf("%s/account.password", PathImportedAccount),
				fmt.Sprintf("%s/account.key", PathImportedAccount),
			}))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				GethDataDir,
				GethBootnodes,
				GethSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
			}))
		})

		It("Should allocate correct resources to node-2 deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpu),
					v1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpuLimit),
					v1.ResourceMemory: resource.MustParse(memoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 imported account secret", func() {
			secret := &v1.Secret{}
			secretKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-%s-imported-account", toCreate.Name, "node-2"),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), secretKey, secret)).To(Succeed())
			Expect(string(secret.Data["account.key"])).To(Equal(string(accountKey)[2:]))
			Expect(string(secret.Data["account.password"])).To(Equal(accountPassword))
		})

		It("Should create node-2 data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
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
			ID:        networkID,
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

		cpu := "500m"
		memory := "500Mi"

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

		It("Should create bootnode deployment with correct arguments", func() {
			nodeDep := &appsv1.Deployment{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
			}))
		})

		It("Should allocate correct resources to bootnode deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPURequest),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryRequest),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create bootnode data persistent volume with correct resouces", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should update the network", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				RPCPort: 8547,
				Resources: &ethereumv1alpha1.NodeResources{
					CPU:    cpu,
					Memory: memory,
				},
				SyncMode: ethereumv1alpha1.FastSynchronization,
			}
			fetched.Spec.Nodes = append(fetched.Spec.Nodes, newNode)
			if !useExistingCluster {
				fetched.Default()
			}
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
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage))
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuBootnodes,
				BesuRPCHTTPEnabled,
				"8547",
				BesuSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
			}))
		})

		It("Should allocate correct resources to node-2 deployment", func() {
			nodeDep := &appsv1.Deployment{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(cpu),
					v1.ResourceMemory: resource.MustParse(memory),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeCPULimit),
					v1.ResourceMemory: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeMemoryLimit),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeDep)).To(Succeed())
			Expect(nodeDep.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node2Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
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
