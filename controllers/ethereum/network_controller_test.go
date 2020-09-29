package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
		useExistingCluster  = os.Getenv("USE_EXISTING_CLUSTER") == "true"
		besuClient, _       = NewEthereumClient(ethereumv1alpha1.BesuClient)
		gethClient, _       = NewEthereumClient(ethereumv1alpha1.GethClient)
		parityClient, _     = NewEthereumClient(ethereumv1alpha1.ParityClient)
		initGenesis         string
		importGethAccount   string
		importParityAccount string
	)

	It("Should generate init container scripts", func() {
		var err error
		initGenesis, err = generateInitGenesisScript()
		Expect(err).To(BeNil())
		importGethAccount, err = generateImportAccountScript(ethereumv1alpha1.GethClient)
		Expect(err).To(BeNil())
		importParityAccount, err = generateImportAccountScript(ethereumv1alpha1.ParityClient)
		Expect(err).To(BeNil())
	})

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
					Logging:  ethereumv1alpha1.NoLogs,
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
		node3Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-3"),
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

		It("Should create configs (genesis, init scripts, static nodes ...) configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-besu", key.Name),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).Should(Succeed())
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

		It("Should create bootnode statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuNetwork,
				"mainnet",
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.NoLogs),
			}))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
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

		It("Should update the network by adding node-2", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-2",
				RPC:     true,
				Client:  ethereumv1alpha1.GethClient,
				RPCPort: 8547,
				Logging: ethereumv1alpha1.ErrorLogs,
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

		It("Should create node-2 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				GethDataDir,
				GethRPCHTTPEnabled,
				GethRPCHTTPPort,
				"8547",
				GethSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.ErrorLogs),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).ToNot(ContainElements([]string{
				ethereumv1alpha1.MainNetwork,
			}))
		})

		It("Should allocate correct resources to node-2 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
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

		It("Should not create privatekey secret for node-2 (without nodekey and not importing account)", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSecret)).ToNot(Succeed())
		})

		It("Should create node-2 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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
				{
					Name:       "json-rpc",
					Port:       int32(8547),
					TargetPort: intstr.FromInt(int(8547)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
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
			It("Should delete node-2 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-2 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
			})
		}

		It("Should update the network by adding node-3", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:    "node-3",
				RPC:     true,
				Client:  ethereumv1alpha1.ParityClient,
				RPCPort: 8547,
				Logging: ethereumv1alpha1.ErrorLogs,
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
			It("Should schedule node-1 and node-3 on different nodes", func() {
				pods := &v1.PodList{}
				matchingLabels := client.MatchingLabels{"name": "node"}
				inNamespace := client.InNamespace(ns.Name)
				Expect(k8sClient.List(context.Background(), pods, matchingLabels, inNamespace)).To(Succeed())
				Expect(pods.Items[0].Spec.NodeName).NotTo(Equal(pods.Items[1].Spec.NodeName))
			})
		}

		It("Should create node-3 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(ParityImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ParityDataDir,
				ParityRPCHTTPPort,
				"8547",
				ParitySyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.ErrorLogs),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).ToNot(ContainElements([]string{
				ethereumv1alpha1.MainNetwork,
			}))
		})

		It("Should allocate correct resources to node-3 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-3 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultMainNetworkFastNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node3Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should not create privatekey secret for node-3 (without nodekey and not importing account)", func() {
			nodeSecret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSecret)).ToNot(Succeed())
		})

		It("Should create node-3 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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
				{
					Name:       "json-rpc",
					Port:       int32(8547),
					TargetPort: intstr.FromInt(int(8547)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
		})

		It("Should update the network by removing node-3", func() {
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
			It("Should delete node-3 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node3Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-3 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).ToNot(Succeed())
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
			It("Should delete bootnode statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).ToNot(Succeed())
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
					Logging:  ethereumv1alpha1.FatalLogs,
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
		node3Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-3"),
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

		It("Should create configs (genesis, init scripts, static nodes ...) configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-besu", key.Name),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).Should(Succeed())
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

		It("Should create bootnode statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuNetwork,
				"rinkeby",
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.FatalLogs),
			}))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))

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

		It("Should update the network by adding node-2", func() {
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
				Logging: ethereumv1alpha1.WarnLogs,
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

		It("Should create node-2 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/import-account.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				"--rinkeby",
				GethDataDir,
				GethMinerEnabled,
				GethMinerCoinbase,
				GethUnlock,
				GethPassword,
				GethSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
			}))
		})

		It("Should allocate correct resources to node-2 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 imported account secret", func() {
			secret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, secret)).To(Succeed())
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

		It("Should create node-2 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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
			It("Should delete node-2 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-2 imported account secret", func() {
				secret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), node2Key, secret)).ToNot(Succeed())
			})

			It("Should delete node-2 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})
		}

		It("Should update the network by adding node-3", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:     "node-3",
				Client:   ethereumv1alpha1.ParityClient,
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
				Logging: ethereumv1alpha1.WarnLogs,
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
			It("Should schedule node-1 and node-3 on different nodes", func() {
				pods := &v1.PodList{}
				matchingLabels := client.MatchingLabels{"name": "node"}
				inNamespace := client.InNamespace(ns.Name)
				Expect(k8sClient.List(context.Background(), pods, matchingLabels, inNamespace)).To(Succeed())
				Expect(pods.Items[0].Spec.NodeName).NotTo(Equal(pods.Items[1].Spec.NodeName))
			})
		}

		It("Should create node-3 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(ParityImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(ParityImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/import-account.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				"rinkeby",
				ParityDataDir,
				ParityMinerCoinbase,
				ParityUnlock,
				ParityPassword,
				ParitySyncMode,
				"archive",
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
			}))
		})

		It("Should allocate correct resources to node-3 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-3 imported account secret", func() {
			secret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node3Key, secret)).To(Succeed())
			Expect(string(secret.Data["account.key"])).To(Equal(string(accountKey)[2:]))
			Expect(string(secret.Data["account.password"])).To(Equal(accountPassword))
		})

		It("Should create node-3 data persistent volume", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultTestNetworkStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node3Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should create node-3 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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

		It("Should update the network by removing node-3", func() {
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
			It("Should delete node-3 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-3 imported account secret", func() {
				secret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), node3Key, secret)).ToNot(Succeed())
			})

			It("Should delete node-3 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node3Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-3 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).ToNot(Succeed())
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
			It("Should delete bootnode statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).ToNot(Succeed())
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
		node3Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-3"),
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

		It("Should create bootnode statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.DefaultLogging),
				BesuDiscoveryEnabled,
				"false",
			}))
		})

		It("Should create bootnode genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-besu", key.Name),
				Namespace: key.Namespace,
			}
			expectedExtraData := "0x0000000000000000000000000000000000000000000000000000000000000000d2c21213027cbf4d46c16b55fa98e5252b0487060000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
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

		It("Should update the network by adding node-2", func() {
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
				Logging: ethereumv1alpha1.DebugLogs,
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

		It("Should create node-2 genesis block and scripts configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-geth", key.Name),
				Namespace: key.Namespace,
			}
			expectedExtraData := "0x0000000000000000000000000000000000000000000000000000000000000000d2c21213027cbf4d46c16b55fa98e5252b0487060000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
			Expect(genesisConfig.Data["init-genesis.sh"]).To(Equal(initGenesis))
			Expect(genesisConfig.Data["import-account.sh"]).To(Equal(importGethAccount))
		})

		It("Should create node-2 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/init-genesis.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[1].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[1].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/import-account.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				GethDataDir,
				GethSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
				GethNoDiscovery,
			}))
		})

		It("Should allocate correct resources to node-2 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 imported account secret", func() {
			secret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, secret)).To(Succeed())
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

		It("Should create node-2 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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
			It("Should delete node-2 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-2 imported account secret", func() {
				secret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), node2Key, secret)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-2 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
			})
		}

		It("Should update the network by adding node-3", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:     "node-3",
				Client:   ethereumv1alpha1.ParityClient,
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
				Logging: ethereumv1alpha1.DebugLogs,
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

		It("Should create node-3 genesis block and scripts configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-parity", key.Name),
				Namespace: key.Namespace,
			}
			expectedExtraData := "0x0000000000000000000000000000000000000000000000000000000000000000d2c21213027cbf4d46c16b55fa98e5252b0487060000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["genesis.json"]).To(ContainSubstring(expectedExtraData))
			Expect(genesisConfig.Data["import-account.sh"]).To(Equal(importParityAccount))
		})

		It("Should create node-3 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(ParityImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(ParityImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/import-account.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ParityDataDir,
				ParitySyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				ParityLogging,
				ParityPassword,
				ParityUnlock,
				ParityEngineSigner,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
				ParityNoDiscovery,
			}))
		})

		It("Should allocate correct resources to node-3 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-3 imported account secret", func() {
			secret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node3Key, secret)).To(Succeed())
			Expect(string(secret.Data["account.key"])).To(Equal(string(accountKey)[2:]))
			Expect(string(secret.Data["account.password"])).To(Equal(accountPassword))
		})

		It("Should create node-3 data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(storage),
				},
			}
			Expect(k8sClient.Get(context.Background(), node3Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should create node-3 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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

		It("Should update the network by removing node-3", func() {
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
			It("Should delete node-3 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-3 imported account secret", func() {
				secret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), node3Key, secret)).ToNot(Succeed())
			})

			It("Should delete node-3 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node3Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-3 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).ToNot(Succeed())
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
			It("Should delete bootnode statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).ToNot(Succeed())
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

			It("Should delete besu genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-besu", key.Name),
					Namespace: key.Namespace,
				}
				Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).ToNot(Succeed())
			})

			It("Should delete geth genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-geth", key.Name),
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
					Logging:  ethereumv1alpha1.TraceLogs,
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
		node3Key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s", toCreate.Name, "node-3"),
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

		It("Should create bootnode genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-besu", key.Name),
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

		It("Should create bootnode statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.TraceLogs),
				BesuDiscoveryEnabled,
				"false",
			}))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
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

		It("Should update the network by adding node-2", func() {
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
				Logging: ethereumv1alpha1.AllLogs,
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

		It("Should create node-2 genesis and scripts block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-geth", key.Name),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
			Expect(genesisConfig.Data["init-genesis.sh"]).To(Equal(initGenesis))
			Expect(genesisConfig.Data["import-account.sh"]).To(Equal(importGethAccount))
		})

		It("Should create node-2 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[0].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/init-genesis.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[1].Image).To(Equal(GethImage()))
			Expect(nodeSts.Spec.Template.Spec.InitContainers[1].Args).To(ContainElements([]string{
				fmt.Sprintf("%s/import-account.sh", PathConfig),
			}))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				GethDataDir,
				GethSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.AllLogs),
				GethNoDiscovery,
			}))
		})

		It("Should allocate correct resources to node-2 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should create node-2 imported account secret", func() {
			secret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node2Key, secret)).To(Succeed())
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

		It("Should create node-2 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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
			It("Should delete node-2 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-2 imported account secret", func() {
				secret := &v1.Secret{}
				Expect(k8sClient.Get(context.Background(), node2Key, secret)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-2 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
			})
		}

		It("Should update the network by adding node-3", func() {
			fetched := &ethereumv1alpha1.Network{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			newNode := ethereumv1alpha1.Node{
				Name:     "node-3",
				Client:   ethereumv1alpha1.ParityClient,
				SyncMode: ethereumv1alpha1.FastSynchronization,
				Resources: &ethereumv1alpha1.NodeResources{
					CPU:         cpu,
					CPULimit:    cpuLimit,
					Memory:      memory,
					MemoryLimit: memoryLimit,
				},
				Logging: ethereumv1alpha1.TraceLogs,
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

		It("Should create node-3 genesis and scripts block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-parity", key.Name),
				Namespace: key.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).To(Succeed())
		})

		It("Should create node-3 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(ParityImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				ParityDataDir,
				ParitySyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.TraceLogs),
				ParityNoDiscovery,
			}))
		})

		It("Should allocate correct resources to node-3 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
		})

		It("Should not create node-2 secret (neither imported account nor node key)", func() {
			secret := &v1.Secret{}
			Expect(k8sClient.Get(context.Background(), node3Key, secret)).ToNot(Succeed())
		})

		It("Should create node-2 data persistent volume with correct resources", func() {
			nodePVC := &v1.PersistentVolumeClaim{}
			expectedResources := v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(ethereumv1alpha1.DefaultPrivateNetworkNodeStorageRequest),
				},
			}
			Expect(k8sClient.Get(context.Background(), node3Key, nodePVC)).To(Succeed())
			Expect(nodePVC.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodePVC.Spec.Resources).To(Equal(expectedResources))
		})

		It("Should create node-3 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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

		It("Should update the network by removing node-3", func() {
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
			It("Should delete node-3 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-3 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node3Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-3 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node3Key, nodeSvc)).ToNot(Succeed())
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
			It("Should delete bootnode statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).ToNot(Succeed())
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

			It("Should delete besu genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-besu", key.Name),
					Namespace: key.Namespace,
				}
				Expect(k8sClient.Get(context.Background(), genesisKey, genesisConfig)).ToNot(Succeed())
			})

			It("Should delete geth genesis block configmap", func() {
				genesisConfig := &v1.ConfigMap{}
				genesisKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-geth", key.Name),
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
					Logging:  ethereumv1alpha1.WarnLogs,
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

		It("Should create bootnode genesis block configmap", func() {
			genesisConfig := &v1.ConfigMap{}
			genesisKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-besu", key.Name),
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

		It("Should create bootnode statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuNodePrivateKey,
				BesuSyncMode,
				string(ethereumv1alpha1.FullSynchronization),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
				BesuDiscoveryEnabled,
				"false",
			}))
		})

		It("Should allocate correct resources to bootnode statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
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
				Logging:  ethereumv1alpha1.DebugLogs,
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

		It("Should create node-2 statefulset with correct arguments", func() {
			nodeSts := &appsv1.StatefulSet{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.GetOwnerReferences()).To(ContainElement(ownerReference))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Image).To(Equal(BesuImage()))
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Args).To(ContainElements([]string{
				BesuDataPath,
				BesuRPCHTTPEnabled,
				"8547",
				BesuSyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
			}))
		})

		It("Should allocate correct resources to node-2 statefulset", func() {
			nodeSts := &appsv1.StatefulSet{}
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
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).To(Succeed())
			Expect(nodeSts.Spec.Template.Spec.Containers[0].Resources).To(Equal(expectedResources))
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

		It("Should create node-2 service", func() {
			nodeSvc := &v1.Service{}
			Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).To(Succeed())
			Expect(nodeSvc.GetOwnerReferences()).To(ContainElement(ownerReference))
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
				{
					Name:       "json-rpc",
					Port:       int32(8547),
					TargetPort: intstr.FromInt(int(8547)),
					Protocol:   corev1.ProtocolTCP,
				},
			}))
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
			It("Should delete node-2 statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSts)).ToNot(Succeed())
			})

			It("Should delete node-2 data persistent volume", func() {
				Eventually(func() error {
					nodePVC := &v1.PersistentVolumeClaim{}
					return k8sClient.Get(context.Background(), node2Key, nodePVC)
				}, timeout, interval).ShouldNot(Succeed())
			})

			It("Should delete node-2 service", func() {
				nodeSvc := &v1.Service{}
				Expect(k8sClient.Get(context.Background(), node2Key, nodeSvc)).ToNot(Succeed())
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
			It("Should delete bootnode statefulset", func() {
				nodeSts := &appsv1.StatefulSet{}
				Expect(k8sClient.Get(context.Background(), bootnodeKey, nodeSts)).ToNot(Succeed())
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
