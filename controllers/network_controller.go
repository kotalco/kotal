/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereumv1alpha1 "github.com/mfarghaly/kotal/api/v1alpha1"
)

// NetworkReconciler reconciles a Network object
type NetworkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets;services,verbs=get;create;update

// Reconcile reconciles ethereum networks
func (r *NetworkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("network", req.NamespacedName)

	var network ethereumv1alpha1.Network

	// Get desired ethereum network
	if err := r.Client.Get(ctx, req.NamespacedName, &network); err != nil {
		log.Error(err, "Unable to fetch Ethereum Network")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	bootnodes := []string{}

	for i, node := range network.Spec.Nodes {
		// one node is a bootnode for every three nodes
		isBootnode := i%3 == 0

		bootnode, err := r.reconcileNode(ctx, &node, &network, isBootnode, bootnodes)
		if err != nil {
			return ctrl.Result{}, err
		}

		if isBootnode {
			bootnodes = append(bootnodes, bootnode)
		}

	}

	if err := r.deleteRedundantNodes(ctx, network.Spec.Nodes, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createNodekey creates private key for node to be used for enodeURL
func (r *NetworkReconciler) createNodekey(hex string) (privateKeyHex, publicKeyHex string, err error) {
	// private key
	var privateKey *ecdsa.PrivateKey

	if hex != "" {
		privateKey, err = crypto.HexToECDSA(hex)
		if err != nil {
			return
		}
		privateKeyHex = hex
	} else {
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			return
		}
		privateKeyBytes := crypto.FromECDSA(privateKey)
		privateKeyHex = hexutil.Encode(privateKeyBytes)[2:]
	}

	// public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err = errors.New("publicKey is not of type *ecdsa.PublicKey")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex = hexutil.Encode(publicKeyBytes)[4:]

	return

}

// createSecretForNode creates a secret for bootnode
func (r *NetworkReconciler) createSecretForNode(name, ns string) *corev1.Secret {

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

// createServiceForNode creates a service that directs traffic to the node
func (r *NetworkReconciler) createServiceForNode(name, ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "discovery",
					Port:       30303,
					TargetPort: intstr.FromInt(30303),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "p2p",
					Port:       30303,
					TargetPort: intstr.FromInt(30303),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app":      "node",
				"instance": name,
			},
		},
	}
}

// deleteRedundantNode deletes all nodes that has been removed from spec
func (r *NetworkReconciler) deleteRedundantNodes(ctx context.Context, nodes []ethereumv1alpha1.Node, ns string) error {
	log := r.Log.WithName("delete redunudant nodes")

	var deps appsv1.DeploymentList
	names := map[string]bool{}

	// all node names in the spec
	for _, node := range nodes {
		names[node.Name] = true
	}

	// all nodes deployments that's currently running
	if err := r.Client.List(ctx, &deps, client.MatchingLabels{"app": "node"}); err != nil {
		log.Error(err, "unable to list all node deployments")
		return err
	}

	for _, dep := range deps.Items {
		name := dep.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("node (%s) deployment doesn't exist anymore in the spec", name))
			log.Info(fmt.Sprintf("deleting node (%s) deployment", name))

			if err := r.Client.Delete(ctx, &dep); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) deployment", name))
				return err
			}
		}
	}

	return nil
}

// reconcileNode create a new node deployment if it doesn't exist
// updates existing deployments if node spec changed
func (r *NetworkReconciler) reconcileNode(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, isBootnode bool, bootnodes []string) (enodeURL string, err error) {
	log := r.Log.WithValues("node", node.Name)

	pvc := r.createPersistentVolumeClaimForNode(node, network.GetNamespace())
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(network, pvc, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference")
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create/update pvc")
		return
	}

	dep := r.createDeploymentForNode(node, network.GetNamespace())

	// TODO: take into account resource and its owner being deleted
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, dep, func() error {
		args := r.createArgsForClient(node, network.Spec.Join, bootnodes)
		if err := ctrl.SetControllerReference(network, dep, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference")
			return err
		}
		volumes := []corev1.Volume{}
		volumeMounts := []corev1.VolumeMount{}

		if isBootnode {
			// add node key arg to the client
			args = append(args, "--node-private-key-file", "/mnt/bootnode/nodekey")
			// create volume from nodekey secret
			nodekeyVolume := corev1.Volume{
				Name: "nodekey",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: node.Name,
					},
				},
			}
			volumes = append(volumes, nodekeyVolume)
			// create volume mount
			nodekeyMount := corev1.VolumeMount{
				Name:      "nodekey",
				MountPath: "/mnt/bootnode/",
				ReadOnly:  true,
			}
			volumeMounts = append(volumeMounts, nodekeyMount)
		}
		// create volume from nodekey secret
		dataVolume := corev1.Volume{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: node.Name,
				},
			},
		}
		volumes = append(volumes, dataVolume)
		// create volume mount
		dataMount := corev1.VolumeMount{
			Name:      "data",
			MountPath: "/mnt/data/",
		}
		volumeMounts = append(volumeMounts, dataMount)
		args = append(args, "--data-path", "/mnt/data/")

		spec := &dep.Spec.Template.Spec
		// attach the volumes to the deployment
		spec.Volumes = volumes
		// mount the volumes
		spec.Containers[0].VolumeMounts = volumeMounts
		dep.Spec.Template.Spec.Containers[0].Args = args
		return nil
	})

	if err != nil {
		return
	}

	if !isBootnode {
		return
	}

	var privateKey, publicKey string

	// create node key
	privateKey, publicKey, err = r.createNodekey("")
	if err != nil {
		return
	}

	// create node key secret
	secret := r.createSecretForNode(node.Name, network.GetNamespace())
	if err = ctrl.SetControllerReference(network, secret, r.Scheme); err != nil {
		log.Error(err, "Unable to set controller reference")
		return
	}
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.CreationTimestamp.IsZero() {
			secret.StringData = map[string]string{
				"nodekey": privateKey,
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	// get node key from deployed secret
	nodekey := string(secret.Data["nodekey"])
	// if deployed nodekey and new generated key differ
	// return old deployed one
	if nodekey != privateKey {
		privateKey, publicKey, err = r.createNodekey(nodekey)
		if err != nil {
			return
		}
	}

	// create service for the node
	svc := r.createServiceForNode(node.Name, network.GetNamespace())
	if err = ctrl.SetControllerReference(network, svc, r.Scheme); err != nil {
		log.Error(err, "Unable to set controller reference")
		return
	}
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		return nil
	})
	if err != nil {
		return
	}

	// TODO: use p2pPort instead of hardcoded 30303 after defaulting
	enodeURL = fmt.Sprintf("enode://%s@%s:30303", publicKey, svc.Spec.ClusterIP)

	return
}

// createPersistentVolumeClaimForNode creates a new pvc for node
func (r *NetworkReconciler) createPersistentVolumeClaimForNode(node *ethereumv1alpha1.Node, ns string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: ns,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}
}

// createArgsForClient create arguments to be passed to the node client from node specs
func (r *NetworkReconciler) createArgsForClient(node *ethereumv1alpha1.Node, join string, bootnodes []string) []string {
	args := []string{"--nat-method", "KUBERNETES"}
	// TODO: update after admissionmutating webhook
	// because it will default all args

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	if join != "" {
		appendArg("--network", join)
	}

	if node.P2PPort != 0 {
		appendArg("--p2p-port", fmt.Sprintf("%d", node.P2PPort))
	}

	if len(bootnodes) != 0 {
		commaSeperatedBootnodes := strings.Join(bootnodes, ",")
		appendArg("--bootnodes", commaSeperatedBootnodes)
	}

	// TODO: create per client type(besu, geth ... etc)
	if node.SyncMode != "" {
		appendArg("--sync-mode", node.SyncMode.String())
	}

	if node.Miner {
		appendArg("--miner-enabled")
	}

	if node.MinerAccount != "" {
		appendArg("--miner-coinbase", node.MinerAccount)
	}

	if node.RPC {
		appendArg("--rpc-http-enabled")
	}

	if node.RPCPort != 0 {
		appendArg("--rpc-http-port", fmt.Sprintf("%d", node.RPCPort))
	}

	if node.RPCHost != "" {
		appendArg("--rpc-http-host", node.RPCHost)
	}

	if len(node.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.RPCAPI {
			apis = append(apis, api.String())
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg("--rpc-http-api", commaSeperatedAPIs)
	}

	if node.WS {
		appendArg("--rpc-ws-enabled")
	}

	if node.WSPort != 0 {
		appendArg("--rpc-ws-port", fmt.Sprintf("%d", node.WSPort))
	}

	if node.WSHost != "" {
		appendArg("--rpc-ws-host", node.WSHost)
	}

	if len(node.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.WSAPI {
			apis = append(apis, api.String())
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg("--rpc-ws-api", commaSeperatedAPIs)
	}

	if node.GraphQL {
		appendArg("--graphql-http-enabled")
	}

	if node.GraphQLPort != 0 {
		appendArg("--graphql-http-port", fmt.Sprintf("%d", node.GraphQLPort))
	}

	if node.GraphQLHost != "" {
		appendArg("--graphql-http-host", node.GraphQLHost)
	}

	if len(node.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Hosts, ",")
		appendArg("--host-whitelist", commaSeperatedHosts)
	}

	if len(node.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.CORSDomains, ",")
		if node.RPC {
			appendArg("--rpc-http-cors-origins", commaSeperatedDomains)
		}
		if node.GraphQL {
			appendArg("--graphql-http-cors-origins", commaSeperatedDomains)
		}
	}

	return args
}

// createDeploymentForNode creates a new deployment for node
func (r *NetworkReconciler) createDeploymentForNode(node *ethereumv1alpha1.Node, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: ns,
			Labels: map[string]string{
				"app":      "node",
				"instance": node.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":      "node",
					"instance": node.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":      "node",
						"instance": node.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "node",
							Image: "hyperledger/besu:1.4.3",
							Command: []string{
								"besu",
							},
						},
					},
				},
			},
		},
	}

}

// SetupWithManager adds reconciler to the manager
func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Network{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
