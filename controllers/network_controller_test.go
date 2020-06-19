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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Ethereum network", func() {

	const (
		sleepTime  = 5 * time.Second
		interval   = 2 * time.Second
		timeout    = 30 * time.Second
		privatekey = ethereumv1alpha1.PrivateKey("0x608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e")
	)

	var (
		useExistingCluster = os.Getenv("USE_EXISTING_CLUSTER") == "true"
	)

	Context("Joining Rinkeby", func() {

		key := types.NamespacedName{
			Name:      "my-network",
			Namespace: "default",
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
			Name:      "node-1",
			Namespace: key.Namespace,
		}
		node2Key := types.NamespacedName{
			Name:      "node-2",
			Namespace: key.Namespace,
		}

		It("Should create the network", func() {
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
				ArgNetwork,
				"rinkeby",
				ArgDataPath,
				ArgNodePrivateKey,
			}))
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

	})

})
