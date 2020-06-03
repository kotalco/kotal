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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

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
	"github.com/mfarghaly/kotal/helpers"
)

const (
	nodekeyPath        = "/mnt/bootnode"
	genesisFilePath    = "/mnt/config"
	blockchainDataPath = "/mnt/data"
)

// NetworkReconciler reconciles a Network object
type NetworkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets;services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list

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

	network.Status.NodesCount = len(network.Spec.Nodes)
	if err := r.Status().Update(ctx, &network); err != nil {
		log.Error(err, "unable to update network status")
		return ctrl.Result{}, err
	}

	// network is not using existing network genesis block
	// TODO:validaiton: genesis is required if there's no network to join
	if network.Spec.Join == "" {
		err := r.reconcileGenesis(ctx, &network)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	bootnodes := []string{}

	for _, node := range network.Spec.Nodes {

		bootnode, err := r.reconcileNode(ctx, &node, &network, bootnodes)
		if err != nil {
			return ctrl.Result{}, err
		}

		if node.IsBootnode() {
			bootnodes = append(bootnodes, bootnode)
		}

	}

	if err := r.deleteRedundantNodes(ctx, network.Spec.Nodes, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NetworkReconciler) reconcileGenesis(ctx context.Context, network *ethereumv1alpha1.Network) error {
	log := r.Log.WithValues("genesis block", network.Name)

	configmap := r.createConfigmapForGenesis(network.Name, network.Namespace)
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(network, configmap, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference")
			return err
		}
		configmap.Data = make(map[string]string)
		b, err := r.createGenesisFile(network)
		if err != nil {
			return err
		}

		configmap.Data["genesis.json"] = string(b)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NetworkReconciler) createConfigmapForGenesis(name, ns string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-genesis", name),
			Namespace: ns,
		},
	}
}

// createExtraDataFromSigners creates extraDta genesis field value from initial signers
func (r *NetworkReconciler) createExtraDataFromSigners(signers []ethereumv1alpha1.EthereumAddress) string {
	extraData := "0x"
	// vanity data
	extraData += strings.Repeat("00", 32)
	// signers
	for _, signer := range signers {
		// append address without the 0x
		extraData += string(signer)[2:]
	}
	// proposer signature
	extraData += strings.Repeat("00", 65)
	return extraData
}

// createExtraDataFromValidators creates extraDta genesis field value from initial validators
func (r *NetworkReconciler) createExtraDataFromValidators(validators []ethereumv1alpha1.EthereumAddress) (string, error) {
	data := []interface{}{}
	extraData := "0x"

	// empty vanity bytes
	vanity := bytes.Repeat([]byte{0x00}, 32)

	// validator addresses bytes
	decodedValidators := []interface{}{}
	for _, validator := range validators {
		validatorBytes, err := hex.DecodeString(string(validator)[2:])
		if err != nil {
			return extraData, err
		}
		decodedValidators = append(decodedValidators, validatorBytes)
	}

	// no vote
	var vote []byte

	// round 0, must be 4 bytes
	round := bytes.Repeat([]byte{0x00}, 4)

	// no committer seals
	committers := []interface{}{}

	// pack all required info into data
	data = append(data, vanity)
	data = append(data, decodedValidators)
	data = append(data, vote)
	data = append(data, round)
	data = append(data, committers)

	// rlp encode data
	payload, err := rlp.EncodeToBytes(data)
	if err != nil {
		return extraData, err
	}

	return extraData + common.Bytes2Hex(payload), nil

}

// createGenesisFile creates genesis config file
func (r *NetworkReconciler) createGenesisFile(network *ethereumv1alpha1.Network) (config string, err error) {
	genesis := network.Spec.Genesis
	mixHash := genesis.MixHash
	nonce := genesis.Nonce
	difficulty := genesis.Difficulty
	result := map[string]interface{}{}

	var consensusConfig map[string]uint
	var extraData string
	var engine string

	if network.Spec.Consensus == ethereumv1alpha1.ProofOfWork {
		consensusConfig = map[string]uint{
			"fixeddifficulty": genesis.Ethash.FixedDifficulty,
		}
		engine = "ethash"
	}

	// clique PoA settings
	if network.Spec.Consensus == ethereumv1alpha1.ProofOfAuthority {
		consensusConfig = map[string]uint{
			"blockperiodseconds": genesis.Clique.BlockPeriod,
			"epochlength":        genesis.Clique.EpochLength,
		}
		engine = "clique"
		extraData = r.createExtraDataFromSigners(network.Spec.Genesis.Clique.Signers)
	}

	// clique ibft2 settings
	if network.Spec.Consensus == ethereumv1alpha1.IstanbulBFT {

		consensusConfig = map[string]uint{
			"blockperiodseconds":        genesis.IBFT2.BlockPeriod,
			"epochlength":               genesis.IBFT2.EpochLength,
			"requesttimeoutseconds":     genesis.IBFT2.RequestTimeout,
			"messageQueueLimit":         genesis.IBFT2.MessageQueueLimit,
			"duplicateMesageLimit":      genesis.IBFT2.DuplicateMesageLimit,
			"futureMessagesLimit":       genesis.IBFT2.FutureMessagesLimit,
			"futureMessagesMaxDistance": genesis.IBFT2.FutureMessagesMaxDistance,
		}
		engine = "ibft2"
		mixHash = "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365"
		nonce = "0x0"
		difficulty = "0x1"
		extraData, err = r.createExtraDataFromValidators(network.Spec.Genesis.IBFT2.Validators)
		if err != nil {
			return
		}
	}

	result["config"] = map[string]interface{}{
		"chainId":                genesis.ChainID,
		"homesteadBlock":         genesis.Forks.Homestead,
		"daoForkBlock":           genesis.Forks.DAO,
		"eip150Block":            genesis.Forks.EIP150,
		"eip150Hash":             genesis.Forks.EIP150Hash,
		"eip155Block":            genesis.Forks.EIP155,
		"eip158Block":            genesis.Forks.EIP158,
		"byzantiumBlock":         genesis.Forks.Byzantium,
		"constantinopleBlock":    genesis.Forks.Constantinople,
		"constantinopleFixBlock": genesis.Forks.Petersburg,
		"istanbulBlock":          genesis.Forks.Istanbul,
		"muirGlacierForkBlock":   genesis.Forks.MuirGlacier,
		engine:                   consensusConfig,
	}

	result["nonce"] = nonce
	result["timestamp"] = genesis.Timestamp
	result["gasLimit"] = genesis.GasLimit
	result["difficulty"] = difficulty
	result["coinbase"] = genesis.Coinbase
	result["mixHash"] = mixHash
	result["extraData"] = extraData

	alloc := map[ethereumv1alpha1.EthereumAddress]interface{}{}
	for _, account := range genesis.Accounts {
		alloc[account.Address] = map[string]interface{}{
			"balance": account.Balance,
			"code":    account.Code,
			"storage": account.Storage,
		}
	}

	result["alloc"] = alloc

	b, err := json.Marshal(result)
	if err != nil {
		return
	}

	config = string(b)

	return
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

// specNodeDataPVC update node data pvc spec
func (r *NetworkReconciler) specNodeDataPVC(pvc *corev1.PersistentVolumeClaim) {
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				// TODO: update storage per network i.e: mainnet, rinkeby, goreli ... etc
				corev1.ResourceStorage: resource.MustParse("10Gi"),
			},
		},
	}
}

// reconcileNodeDataPVC creates node data pvc if it doesn't exist
func (r *NetworkReconciler) reconcileNodeDataPVC(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) error {

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: network.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(network, pvc, r.Scheme); err != nil {
			return err
		}
		if pvc.CreationTimestamp.IsZero() {
			r.specNodeDataPVC(pvc)
		}
		return nil
	})

	return err
}

// createNodeVolumes creates all the required volumes for the node
func (r *NetworkReconciler) createNodeVolumes(node, network string, nodekey, customGenesis bool) []corev1.Volume {

	volumes := []corev1.Volume{}

	if nodekey {
		nodekeyVolume := corev1.Volume{
			Name: "nodekey",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: node,
				},
			},
		}
		volumes = append(volumes, nodekeyVolume)
	}

	if customGenesis {
		genesisVolume := corev1.Volume{
			Name: "genesis",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-genesis", network),
					},
				},
			},
		}
		volumes = append(volumes, genesisVolume)
	}

	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: node,
			},
		},
	}
	volumes = append(volumes, dataVolume)

	return volumes
}

// createNodeVolumeMounts creates all required volume mounts for the node
func (r *NetworkReconciler) createNodeVolumeMounts(node, network string, nodekey, customGenesis bool) []corev1.VolumeMount {

	volumeMounts := []corev1.VolumeMount{}

	if nodekey {
		nodekeyMount := corev1.VolumeMount{
			Name:      "nodekey",
			MountPath: nodekeyPath,
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, nodekeyMount)
	}

	if customGenesis {
		genesisMount := corev1.VolumeMount{
			Name:      "genesis",
			MountPath: genesisFilePath,
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, genesisMount)
	}

	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: blockchainDataPath,
	}
	volumeMounts = append(volumeMounts, dataMount)

	return volumeMounts
}

// specNodeDeployment updates node deployment spec
func (r *NetworkReconciler) specNodeDeployment(dep *appsv1.Deployment, args []string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	nodeName := dep.ObjectMeta.Name
	dep.ObjectMeta.Labels = map[string]string{
		"app":      "node",
		"instance": nodeName,
	}
	dep.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":      "node",
				"instance": nodeName,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":      "node",
					"instance": nodeName,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "node",
						Image: "hyperledger/besu:1.4.5",
						Command: []string{
							"besu",
						},
					},
				},
			},
		},
	}
	// attach the volumes to the deployment
	dep.Spec.Template.Spec.Volumes = volumes
	// mount the volumes
	dep.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	// TODO: recfactor this, will fail if container order change
	dep.Spec.Template.Spec.Containers[0].Args = args
}

// reconcileNodeDeployment creates creates node deployment if it doesn't exist, update it if it does exist
func (r *NetworkReconciler) reconcileNodeDeployment(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes, args []string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) error {

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: network.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, dep, func() error {
		if err := ctrl.SetControllerReference(network, dep, r.Scheme); err != nil {
			return err
		}
		r.specNodeDeployment(dep, args, volumes, volumeMounts)
		return nil
	})

	return err
}

// reconcileNodeSecret creates node secret if it doesn't exist, update it if it exists
func (r *NetworkReconciler) reconcileNodeSecret(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, nodekey string) (publicKey string, err error) {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: network.Namespace,
		},
	}

	privateKey, publicKey, err := helpers.CreateNodeKeypair(nodekey)
	if err != nil {
		return
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if err := ctrl.SetControllerReference(network, secret, r.Scheme); err != nil {
			return err
		}

		secret.StringData = map[string]string{
			"nodekey": privateKey,
		}

		return nil
	})

	if err != nil {
		return
	}

	return
}

// specNodeService updates node service spec
func (r *NetworkReconciler) specNodeService(svc *corev1.Service) {
	svc.Spec.Ports = []corev1.ServicePort{
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
	}

	svc.Spec.Selector = map[string]string{
		"app":      "node",
		"instance": svc.ObjectMeta.Name,
	}
}

// reconcileNodeService reconciles node service
func (r *NetworkReconciler) reconcileNodeService(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) (ip string, err error) {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: network.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err = ctrl.SetControllerReference(network, svc, r.Scheme); err != nil {
			return err
		}

		r.specNodeService(svc)

		return nil
	})

	if err != nil {
		return
	}

	ip = svc.Spec.ClusterIP

	return
}

// reconcileNode create a new node deployment if it doesn't exist
// updates existing deployments if node spec changed
func (r *NetworkReconciler) reconcileNode(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) (enodeURL string, err error) {

	customGenesis := network.Spec.Join == ""

	if err = r.reconcileNodeDataPVC(ctx, node, network); err != nil {
		return
	}

	args := r.createArgsForClient(node, network.Spec.Join, bootnodes, customGenesis)
	volumes := r.createNodeVolumes(node.Name, network.Name, node.WithNodekey(), customGenesis)
	mounts := r.createNodeVolumeMounts(node.Name, network.Name, node.WithNodekey(), customGenesis)

	if err = r.reconcileNodeDeployment(ctx, node, network, bootnodes, args, volumes, mounts); err != nil {
		return
	}

	if !node.WithNodekey() {
		return
	}

	var publicKey string
	from := string(node.Nodekey)[2:]

	if publicKey, err = r.reconcileNodeSecret(ctx, node, network, from); err != nil {
		return
	}

	if !node.IsBootnode() {
		return
	}

	ip, err := r.reconcileNodeService(ctx, node, network)
	if err != nil {
		return
	}

	enodeURL = fmt.Sprintf("enode://%s@%s:%d", publicKey, ip, node.P2PPort)

	return
}

// createArgsForClient create arguments to be passed to the node client from node specs
func (r *NetworkReconciler) createArgsForClient(node *ethereumv1alpha1.Node, join string, bootnodes []string, customGenesis bool) []string {
	args := []string{"--nat-method", "KUBERNETES"}
	// TODO: update after admissionmutating webhook
	// because it will default all args

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	if node.WithNodekey() {
		appendArg("--node-private-key-file", fmt.Sprintf("%s/nodekey", nodekeyPath))
	}

	if customGenesis {
		appendArg("--genesis-file", fmt.Sprintf("%s/genesis.json", genesisFilePath))
	}

	appendArg("--data-path", blockchainDataPath)

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

	if node.Coinbase != "" {
		appendArg("--miner-coinbase", string(node.Coinbase))
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

// SetupWithManager adds reconciler to the manager
func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Network{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
