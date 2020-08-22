package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/helpers"
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
// +kubebuilder:rbac:groups=core,resources=secrets;services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles ethereum networks
func (r *NetworkReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()

	var network ethereumv1alpha1.Network

	// Get desired ethereum network
	if err = r.Client.Get(ctx, req.NamespacedName, &network); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// update network status
	if err = r.updateStatus(ctx, &network); err != nil {
		return
	}

	// reconcile genesis for private network with custom genesis
	if network.Spec.Genesis != nil {
		if err = r.reconcileGenesis(ctx, &network); err != nil {
			return
		}
	}

	// reconcile network nodes
	if err = r.reconcileNodes(ctx, &network); err != nil {
		return
	}

	return

}

// updateStatus updates network status
// TODO: don't update statuse on network deletion
func (r *NetworkReconciler) updateStatus(ctx context.Context, network *ethereumv1alpha1.Network) error {
	network.Status.NodesCount = len(network.Spec.Nodes)

	if err := r.Status().Update(ctx, network); err != nil {
		r.Log.Error(err, "unable to update network status")
		return err
	}

	return nil
}

// reconcileNodes creates or updates nodes according to nodes spec
// deletes nodes missing from nodes spec
func (r *NetworkReconciler) reconcileNodes(ctx context.Context, network *ethereumv1alpha1.Network) error {
	bootnodes := []string{}

	for _, node := range network.Spec.Nodes {

		bootnode, err := r.reconcileNode(ctx, &node, network, bootnodes)
		if err != nil {
			return err
		}

		if node.IsBootnode() {
			bootnodes = append(bootnodes, bootnode)
		}

	}

	if err := r.deleteRedundantNodes(ctx, network); err != nil {
		return err
	}

	return nil
}

// specGenesisConfigmap updates genesis config map spec
func (r *NetworkReconciler) specGenesisConfigmap(configmap *corev1.ConfigMap, data string) {
	configmap.Data = make(map[string]string)
	configmap.Data["genesis.json"] = data
}

// reconcileGenesis creates genesis config map if it doesn't exist or update it
func (r *NetworkReconciler) reconcileGenesis(ctx context.Context, network *ethereumv1alpha1.Network) error {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      network.Spec.Genesis.ConfigmapName(network.Name),
			Namespace: network.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(network, configmap, r.Scheme); err != nil {
			r.Log.Error(err, "Unable to set controller reference on genesis configmap")
			return err
		}

		data, err := r.createGenesisFile(network)
		if err != nil {
			return err
		}

		r.specGenesisConfigmap(configmap, data)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// createGenesisFile creates genesis config file
func (r *NetworkReconciler) createGenesisFile(network *ethereumv1alpha1.Network) (content string, err error) {
	genesis := network.Spec.Genesis
	mixHash := genesis.MixHash
	nonce := genesis.Nonce
	difficulty := genesis.Difficulty
	result := map[string]interface{}{}

	var consensusConfig map[string]uint
	var extraData string
	var engine string

	if network.Spec.Consensus == ethereumv1alpha1.ProofOfWork {
		consensusConfig = map[string]uint{}

		if genesis.Ethash.FixedDifficulty != nil {
			consensusConfig["fixeddifficulty"] = *genesis.Ethash.FixedDifficulty
		}

		engine = "ethash"
	}

	// clique PoA settings
	if network.Spec.Consensus == ethereumv1alpha1.ProofOfAuthority {
		consensusConfig = map[string]uint{
			// besu
			"blockperiodseconds": genesis.Clique.BlockPeriod,
			"epochlength":        genesis.Clique.EpochLength,
			// geth
			"period": genesis.Clique.BlockPeriod,
			"epoch":  genesis.Clique.EpochLength,
		}
		engine = "clique"
		extraData = createExtraDataFromSigners(network.Spec.Genesis.Clique.Signers)
	}

	// clique ibft2 settings
	if network.Spec.Consensus == ethereumv1alpha1.IstanbulBFT {

		consensusConfig = map[string]uint{
			"blockperiodseconds":        genesis.IBFT2.BlockPeriod,
			"epochlength":               genesis.IBFT2.EpochLength,
			"requesttimeoutseconds":     genesis.IBFT2.RequestTimeout,
			"messageQueueLimit":         genesis.IBFT2.MessageQueueLimit,
			"duplicateMessageLimit":     genesis.IBFT2.DuplicateMessageLimit,
			"futureMessagesLimit":       genesis.IBFT2.FutureMessagesLimit,
			"futureMessagesMaxDistance": genesis.IBFT2.FutureMessagesMaxDistance,
		}
		engine = "ibft2"
		mixHash = "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365"
		nonce = "0x0"
		difficulty = "0x1"
		extraData, err = createExtraDataFromValidators(network.Spec.Genesis.IBFT2.Validators)
		if err != nil {
			return
		}
	}

	config := map[string]interface{}{
		"chainId":             genesis.ChainID,
		"homesteadBlock":      genesis.Forks.Homestead,
		"eip150Block":         genesis.Forks.EIP150,
		"eip150Hash":          genesis.Forks.EIP150Hash,
		"eip155Block":         genesis.Forks.EIP155,
		"eip158Block":         genesis.Forks.EIP158,
		"byzantiumBlock":      genesis.Forks.Byzantium,
		"constantinopleBlock": genesis.Forks.Constantinople,
		"petersburgBlock":     genesis.Forks.Petersburg,
		"istanbulBlock":       genesis.Forks.Istanbul,
		"muirGlacierBlock":    genesis.Forks.MuirGlacier,
		engine:                consensusConfig,
	}

	if genesis.Forks.DAO != nil {
		config["daoForkBlock"] = genesis.Forks.DAO
		config["daoForkSupport"] = true //geth
	}

	result["config"] = config

	result["nonce"] = nonce
	result["timestamp"] = genesis.Timestamp
	result["gasLimit"] = genesis.GasLimit
	result["difficulty"] = difficulty
	result["coinbase"] = genesis.Coinbase
	result["mixHash"] = mixHash
	result["extraData"] = extraData

	alloc := map[ethereumv1alpha1.EthereumAddress]interface{}{}
	for _, account := range genesis.Accounts {
		m := map[string]interface{}{
			"balance": account.Balance,
		}

		if account.Code != "" {
			m["code"] = account.Code
		}

		if account.Storage != nil {
			m["storage"] = account.Storage
		}

		alloc[account.Address] = m
	}

	result["alloc"] = alloc

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	content = string(data)

	return
}

// deleteRedundantNode deletes all nodes that has been removed from spec
// network is the owner of the redundant resources (node deployment, svc, secret and pvc)
// removing nodes from spec won't remove these resources by grabage collection
// that's why we're deleting them manually
func (r *NetworkReconciler) deleteRedundantNodes(ctx context.Context, network *ethereumv1alpha1.Network) error {
	log := r.Log.WithName("delete redundant nodes")

	var deps appsv1.DeploymentList
	var pvcs corev1.PersistentVolumeClaimList
	var secrets corev1.SecretList
	var services corev1.ServiceList

	nodes := network.Spec.Nodes
	names := map[string]bool{}
	matchingLabels := client.MatchingLabels{
		"name":    "node",
		"network": network.Name,
	}
	inNamespace := client.InNamespace(network.Namespace)

	for _, node := range nodes {
		depName := node.DeploymentName(network.Name)
		names[depName] = true
	}

	// Node deployments
	if err := r.Client.List(ctx, &deps, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node deployments")
		return err
	}

	for _, dep := range deps.Items {
		name := dep.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) deployment", name))

			if err := r.Client.Delete(ctx, &dep); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) deployment", name))
				return err
			}
		}
	}

	// Node PVCs
	if err := r.Client.List(ctx, &pvcs, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node pvcs")
		return err
	}

	for _, pvc := range pvcs.Items {
		name := pvc.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) pvc", name))

			if err := r.Client.Delete(ctx, &pvc); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) pvc", name))
				return err
			}
		}
	}

	// Node Secrets
	if err := r.Client.List(ctx, &secrets, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node secrets")
		return err
	}

	for _, secret := range secrets.Items {
		name := secret.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) secret", name))

			if err := r.Client.Delete(ctx, &secret); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) secret", name))
				return err
			}
		}
	}

	// Node Services
	if err := r.Client.List(ctx, &services, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node services")
		return err
	}

	for _, service := range services.Items {
		name := service.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) service", name))

			if err := r.Client.Delete(ctx, &service); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) service", name))
				return err
			}
		}
	}

	return nil
}

// specNodeDataPVC update node data pvc spec
func (r *NetworkReconciler) specNodeDataPVC(pvc *corev1.PersistentVolumeClaim, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) {
	pvc.ObjectMeta.Labels = node.Labels(network.Name)
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(node.Resources.Storage),
			},
		},
	}
}

// reconcileNodeDataPVC creates node data pvc if it doesn't exist
func (r *NetworkReconciler) reconcileNodeDataPVC(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) error {

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.PVCName(network.Name),
			Namespace: network.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(network, pvc, r.Scheme); err != nil {
			return err
		}
		if pvc.CreationTimestamp.IsZero() {
			r.specNodeDataPVC(pvc, node, network)
		}
		return nil
	})

	return err
}

// createNodeVolumes creates all the required volumes for the node
func (r *NetworkReconciler) createNodeVolumes(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) []corev1.Volume {

	volumes := []corev1.Volume{}

	if node.WithNodekey() {
		nodekeyVolume := corev1.Volume{
			Name: "nodekey",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: node.SecretName(network.Name),
				},
			},
		}
		volumes = append(volumes, nodekeyVolume)
	}

	if network.Spec.Genesis != nil {
		genesisVolume := corev1.Volume{
			Name: "genesis",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: network.Spec.Genesis.ConfigmapName(network.Name),
					},
				},
			},
		}
		volumes = append(volumes, genesisVolume)
	}

	if node.Import != nil {
		importedAccount := corev1.Volume{
			Name: "imported-account",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: node.ImportedAccountName(network.Name),
				},
			},
		}
		volumes = append(volumes, importedAccount)
	}

	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: node.PVCName(network.Name),
			},
		},
	}
	volumes = append(volumes, dataVolume)

	return volumes
}

// createNodeVolumeMounts creates all required volume mounts for the node
func (r *NetworkReconciler) createNodeVolumeMounts(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) []corev1.VolumeMount {

	volumeMounts := []corev1.VolumeMount{}

	if node.WithNodekey() {
		nodekeyMount := corev1.VolumeMount{
			Name:      "nodekey",
			MountPath: PathNodekey,
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, nodekeyMount)
	}

	if network.Spec.Genesis != nil {
		genesisMount := corev1.VolumeMount{
			Name:      "genesis",
			MountPath: PathGenesisFile,
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, genesisMount)
	}

	if node.Import != nil {
		importedAccountMount := corev1.VolumeMount{
			Name:      "imported-account",
			MountPath: PathImportedAccount,
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, importedAccountMount)
	}

	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: PathBlockchainData,
	}
	volumeMounts = append(volumeMounts, dataMount)

	return volumeMounts
}

func (r *NetworkReconciler) getNodeAffinity(network *ethereumv1alpha1.Network) *corev1.Affinity {
	if network.Spec.HighlyAvailable {
		return &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"name": "node",
								// TODO: add network to restrict affinity effect to single network
							},
						},
						TopologyKey: network.Spec.TopologyKey,
					},
				},
			},
		}
	}
	return nil
}

// specNodeDeployment updates node deployment spec
func (r *NetworkReconciler) specNodeDeployment(dep *appsv1.Deployment, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, args []string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, affinity *corev1.Affinity) {
	labels := node.Labels(network.Name)
	// used by geth to init genesis and import account(s)
	initContainers := []corev1.Container{}
	// node client container
	nodeContainer := corev1.Container{
		Name: "node",
		Args: args,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(node.Resources.CPU),
				corev1.ResourceMemory: resource.MustParse(node.Resources.Memory),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(node.Resources.CPULimit),
				corev1.ResourceMemory: resource.MustParse(node.Resources.MemoryLimit),
			},
		},
		VolumeMounts: volumeMounts,
	}

	if node.Client == ethereumv1alpha1.GethClient {
		if network.Spec.Join == "" {
			initGenesis := corev1.Container{
				Name:  "init-genesis",
				Image: GethImage,
				Command: []string{
					"/bin/sh",
				},
				Args: []string{
					"-c",
					fmt.Sprintf(
						"if [ ! -d %s/geth ]; then geth init %s %s %s ;else echo \"%s\" ;fi",
						PathBlockchainData,
						GethDataDir,
						PathBlockchainData,
						fmt.Sprintf("%s/genesis.json", PathGenesisFile),
						"Genesis block has been initialized before!",
					),
				},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, initGenesis)
		}
		if node.Import != nil {
			importAccount := corev1.Container{
				Name:  "import-account",
				Image: GethImage,
				Command: []string{
					"/bin/sh",
				},
				Args: []string{
					"-c",
					fmt.Sprintf(
						"if [ -z \"$(ls -A %s/keystore)\" ]; then geth account import %s %s %s %s %s ;else echo \"%s\" ;fi",
						PathBlockchainData,
						GethDataDir,
						PathBlockchainData,
						GethPassword,
						fmt.Sprintf("%s/account.password", PathImportedAccount),
						fmt.Sprintf("%s/account.key", PathImportedAccount),
						"Account has been imported before!",
					),
				},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, importAccount)
		}

		nodeContainer.Image = GethImage
		nodeContainer.Command = []string{"geth"}

	} else if node.Client == ethereumv1alpha1.BesuClient {
		nodeContainer.Image = BesuImage
		nodeContainer.Command = []string{"besu"}
	}

	dep.ObjectMeta.Labels = labels
	if dep.Spec.Selector == nil {
		dep.Spec.Selector = &metav1.LabelSelector{}
	}
	dep.Spec.Selector.MatchLabels = labels
	dep.Spec.Template.ObjectMeta.Labels = labels
	dep.Spec.Template.Spec = corev1.PodSpec{
		Volumes:        volumes,
		InitContainers: initContainers,
		Containers:     []corev1.Container{nodeContainer},
		Affinity:       affinity,
	}
}

// reconcileNodeDeployment creates creates node deployment if it doesn't exist, update it if it does exist
func (r *NetworkReconciler) reconcileNodeDeployment(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes, args []string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, affinity *corev1.Affinity) error {

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.DeploymentName(network.Name),
			Namespace: network.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, dep, func() error {
		if err := ctrl.SetControllerReference(network, dep, r.Scheme); err != nil {
			return err
		}
		r.specNodeDeployment(dep, node, network, args, volumes, volumeMounts, affinity)
		return nil
	})

	return err
}

func (r *NetworkReconciler) specNodeSecret(secret *corev1.Secret, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) {
	privateKey := string(node.Nodekey)[2:]
	secret.ObjectMeta.Labels = node.Labels(network.Name)
	secret.StringData = map[string]string{
		"nodekey": privateKey,
	}
}

// reconcileNodeSecret creates node secret if it doesn't exist, update it if it exists
func (r *NetworkReconciler) reconcileNodeSecret(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) (publicKey string, err error) {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.SecretName(network.Name),
			Namespace: network.Namespace,
		},
	}

	// hex private key without the leading 0x
	privateKey := string(node.Nodekey)[2:]
	publicKey, err = helpers.DerivePublicKey(privateKey)
	if err != nil {
		return
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if err := ctrl.SetControllerReference(network, secret, r.Scheme); err != nil {
			return err
		}

		r.specNodeSecret(secret, node, network)

		return nil
	})

	if err != nil {
		return
	}

	return
}

// specNodeService updates node service spec
func (r *NetworkReconciler) specNodeService(svc *corev1.Service, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) {
	labels := node.Labels(network.Name)
	svc.ObjectMeta.Labels = labels
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

	svc.Spec.Selector = labels
}

// reconcileNodeService reconciles node service
func (r *NetworkReconciler) reconcileNodeService(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) (ip string, err error) {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.ServiceName(network.Name),
			Namespace: network.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err = ctrl.SetControllerReference(network, svc, r.Scheme); err != nil {
			return err
		}

		r.specNodeService(svc, node, network)

		return nil
	})

	if err != nil {
		return
	}

	ip = svc.Spec.ClusterIP

	return
}

func (r *NetworkReconciler) specImportedAccountSecret(secret *corev1.Secret, node *ethereumv1alpha1.Node) {
	// TODO: update labels for delete redundant nodes resources
	secret.StringData = map[string]string{
		"account.key":      string(node.Import.PrivateKey)[2:],
		"account.password": node.Import.Password,
	}
}

func (r *NetworkReconciler) reconcileImportedAccountSecret(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.ImportedAccountName(network.Name),
			Namespace: network.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if err := ctrl.SetControllerReference(network, secret, r.Scheme); err != nil {
			return err
		}

		r.specImportedAccountSecret(secret, node)

		return nil
	})

	return err
}

// reconcileNode create a new node deployment if it doesn't exist
// updates existing deployments if node spec changed
func (r *NetworkReconciler) reconcileNode(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) (enodeURL string, err error) {

	if err = r.reconcileNodeDataPVC(ctx, node, network); err != nil {
		return
	}

	args := r.createArgsForClient(node, network, bootnodes)
	volumes := r.createNodeVolumes(node, network)
	mounts := r.createNodeVolumeMounts(node, network)
	affinity := r.getNodeAffinity(network)

	if err = r.reconcileNodeDeployment(ctx, node, network, bootnodes, args, volumes, mounts, affinity); err != nil {
		return
	}

	if node.Import != nil {
		if err = r.reconcileImportedAccountSecret(ctx, node, network); err != nil {
			return
		}
	}

	if !node.WithNodekey() {
		return
	}

	var publicKey string

	if publicKey, err = r.reconcileNodeSecret(ctx, node, network); err != nil {
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
func (r *NetworkReconciler) createArgsForClient(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) []string {
	switch node.Client {
	case ethereumv1alpha1.BesuClient:
		return r.createArgsForBesu(node, network, bootnodes)
	case ethereumv1alpha1.GethClient:
		return r.createArgsForGeth(node, network, bootnodes)
	}
	return []string{}
}

// createArgsForGeth create arguments to be passed to the node client from node specs
func (r *NetworkReconciler) createArgsForGeth(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) []string {
	args := []string{"--nousb"}

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	if network.Spec.ID != 0 {
		appendArg(GethNetworkID, fmt.Sprintf("%d", network.Spec.ID))
	}

	if node.WithNodekey() {
		appendArg(GethNodeKey, fmt.Sprintf("%s/nodekey", PathNodekey))
	}

	appendArg(GethDataDir, PathBlockchainData)

	// TODO: restrict networks to be rinkeby, ropsten, and goerli if client is geth
	if network.Spec.Join != "" && network.Spec.Join != ethereumv1alpha1.MainNetwork {
		appendArg(fmt.Sprintf("--%s", network.Spec.Join))
	}

	if node.P2PPort != 0 {
		appendArg(GethP2PPort, fmt.Sprintf("%d", node.P2PPort))
	}

	if len(bootnodes) != 0 {
		commaSeperatedBootnodes := strings.Join(bootnodes, ",")
		appendArg(GethBootnodes, commaSeperatedBootnodes)
	}

	if node.SyncMode != "" {
		appendArg(GethSyncMode, string(node.SyncMode))
	}

	if node.Miner {
		appendArg(GethMinerEnabled)
	}

	if node.Coinbase != "" {
		appendArg(GethMinerCoinbase, string(node.Coinbase))
		appendArg(GethUnlock, string(node.Coinbase))
		appendArg(GethPassword, fmt.Sprintf("%s/account.password", PathImportedAccount))
	}

	if node.RPC {
		appendArg(GethRPCHTTPEnabled)
	}

	if node.RPCPort != 0 {
		appendArg(GethRPCHTTPPort, fmt.Sprintf("%d", node.RPCPort))
	}

	if node.RPCHost != "" {
		appendArg(GethRPCHTTPHost, node.RPCHost)
	}

	if len(node.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(GethRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.WS {
		appendArg(GethRPCWSEnabled)
	}

	if node.WSPort != 0 {
		appendArg(GethRPCWSPort, fmt.Sprintf("%d", node.WSPort))
	}

	if node.WSHost != "" {
		appendArg(GethRPCWSHost, node.WSHost)
	}

	if len(node.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(GethRPCWSAPI, commaSeperatedAPIs)
	}

	if node.GraphQL {
		appendArg(GethGraphQLHTTPEnabled)
	}

	if node.GraphQLPort != 0 {
		appendArg(GethGraphQLHTTPPort, fmt.Sprintf("%d", node.GraphQLPort))
	}

	if node.GraphQLHost != "" {
		appendArg(GethGraphQLHTTPHost, node.GraphQLHost)
	}

	if len(node.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Hosts, ",")
		if node.RPC {
			appendArg(GethRPCHostWhitelist, commaSeperatedHosts)
		}
		if node.GraphQL {
			appendArg(GethGraphQLHostWhitelist, commaSeperatedHosts)
		}
	}

	if len(node.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.CORSDomains, ",")
		if node.RPC {
			appendArg(GethRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.GraphQL {
			appendArg(GethGraphQLHTTPCorsOrigins, commaSeperatedDomains)
		}
	}

	return args
}

// createArgsForBesu create arguments to be passed to the node client from node specs
func (r *NetworkReconciler) createArgsForBesu(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) []string {
	args := []string{BesuNatMethod, "KUBERNETES"}
	// TODO: update after admissionmutating webhook
	// because it will default all args

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	if network.Spec.ID != 0 {
		appendArg(BesuNetworkID, fmt.Sprintf("%d", network.Spec.ID))
	}

	if node.WithNodekey() {
		appendArg(BesuNodePrivateKey, fmt.Sprintf("%s/nodekey", PathNodekey))
	}

	if network.Spec.Genesis != nil {
		appendArg(BesuGenesisFile, fmt.Sprintf("%s/genesis.json", PathGenesisFile))
	}

	appendArg(BesuDataPath, PathBlockchainData)

	if network.Spec.Join != "" {
		appendArg(BesuNetwork, network.Spec.Join)
	}

	if node.P2PPort != 0 {
		appendArg(BesuP2PPort, fmt.Sprintf("%d", node.P2PPort))
	}

	if len(bootnodes) != 0 {
		commaSeperatedBootnodes := strings.Join(bootnodes, ",")
		appendArg(BesuBootnodes, commaSeperatedBootnodes)
	}

	if node.SyncMode != "" {
		appendArg(BesuSyncMode, string(node.SyncMode))
	}

	if node.Miner {
		appendArg(BesuMinerEnabled)
	}

	if node.Coinbase != "" {
		appendArg(BesuMinerCoinbase, string(node.Coinbase))
	}

	if node.RPC {
		appendArg(BesuRPCHTTPEnabled)
	}

	if node.RPCPort != 0 {
		appendArg(BesuRPCHTTPPort, fmt.Sprintf("%d", node.RPCPort))
	}

	if node.RPCHost != "" {
		appendArg(BesuRPCHTTPHost, node.RPCHost)
	}

	if len(node.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(BesuRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.WS {
		appendArg(BesuRPCWSEnabled)
	}

	if node.WSPort != 0 {
		appendArg(BesuRPCWSPort, fmt.Sprintf("%d", node.WSPort))
	}

	if node.WSHost != "" {
		appendArg(BesuRPCWSHost, node.WSHost)
	}

	if len(node.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(BesuRPCWSAPI, commaSeperatedAPIs)
	}

	if node.GraphQL {
		appendArg(BesuGraphQLHTTPEnabled)
	}

	if node.GraphQLPort != 0 {
		appendArg(BesuGraphQLHTTPPort, fmt.Sprintf("%d", node.GraphQLPort))
	}

	if node.GraphQLHost != "" {
		appendArg(BesuGraphQLHTTPHost, node.GraphQLHost)
	}

	if len(node.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Hosts, ",")
		appendArg(BesuHostWhitelist, commaSeperatedHosts)
	}

	if len(node.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.CORSDomains, ",")
		if node.RPC {
			appendArg(BesuRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.GraphQL {
			appendArg(BesuGraphQLHTTPCorsOrigins, commaSeperatedDomains)
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
