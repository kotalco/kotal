package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	ethereumClients "github.com/kotalco/kotal/clients/ethereum"
	"github.com/kotalco/kotal/controllers/shared"
	"github.com/kotalco/kotal/helpers"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	envCoinbase = "KOTAL_COINBASE"
)

var (
	//go:embed geth_init_genesis.sh
	GethInitGenesisScript string
	//go:embed geth_import_account.sh
	gethImportAccountScript string
	//go:embed nethermind_convert_enode_privatekey.sh
	nethermindConvertEnodePrivateKeyScript string
	//go:embed nethermind_copy_keystore.sh
	nethermindConvertCopyKeystoreScript string
)

// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=secrets;services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles ethereum networks
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node ethereumv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, string(node.Spec.Client), node.Spec.Network)
	r.updateStaticNodes(ctx, &node)
	r.updateBootnodes(ctx, &node)

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	ip, err := r.reconcileService(ctx, &node)
	if err != nil {
		return
	}

	if err = r.reconcileStatefulSet(ctx, &node); err != nil {
		return
	}

	var publicKey string
	if publicKey, err = r.reconcileSecret(ctx, &node); err != nil {
		return
	}

	enodeURL := fmt.Sprintf("enode://%s@%s:%d", publicKey, ip, node.Spec.P2PPort)

	if err = r.updateStatus(ctx, &node, enodeURL); err != nil {
		return
	}

	return ctrl.Result{}, nil
}

// getEnodeURL fetch enodeURL from enode that has the format of node.namespace
// name is the node name, and namespace is the node namespace
func (r *NodeReconciler) getEnodeURL(ctx context.Context, enode, ns string) (string, error) {
	node := &ethereumv1alpha1.Node{}
	var name, namespace string

	if parts := strings.Split(enode, "."); len(parts) > 1 {
		name = parts[0]
		namespace = parts[1]
	} else {
		// nodes without . refered to nodes in the current node namespace
		name = enode
		namespace = ns
	}

	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := r.Client.Get(ctx, namespacedName, node); err != nil {
		return "", err
	}

	return node.Status.EnodeURL, nil
}

// updateStaticNodes replaces Ethereum node references with their enodeURL
func (r *NodeReconciler) updateStaticNodes(ctx context.Context, node *ethereumv1alpha1.Node) {
	log := log.FromContext(ctx)
	for i, enode := range node.Spec.StaticNodes {
		if !strings.HasPrefix(string(enode), "enode://") {
			enodeURL, err := r.getEnodeURL(ctx, string(enode), node.Namespace)
			if err != nil {
				// remove static node reference, so it won't be included into static nodes file
				// don't return the error, node maybe not up and running yet
				node.Spec.StaticNodes = append(node.Spec.StaticNodes[:i], node.Spec.StaticNodes[i+1:]...)
				log.Error(err, "failed to get static node")
				continue
			}
			log.Info("static node enodeURL", string(enode), enodeURL)
			// replace reference with actual enode url
			if strings.HasPrefix(enodeURL, "enode://") {
				node.Spec.StaticNodes[i] = ethereumv1alpha1.Enode(enodeURL)
			} else {
				// remove static node reference, so it won't be included into static nodes file
				node.Spec.StaticNodes = append(node.Spec.StaticNodes[:i], node.Spec.StaticNodes[i+1:]...)
			}
		}
	}
}

// updateBootnodes replaces Ethereum node references with their enodeURL
func (r *NodeReconciler) updateBootnodes(ctx context.Context, node *ethereumv1alpha1.Node) {
	log := log.FromContext(ctx)
	for i, enode := range node.Spec.Bootnodes {
		if !strings.HasPrefix(string(enode), "enode://") {
			enodeURL, err := r.getEnodeURL(ctx, string(enode), node.Namespace)
			if err != nil {
				// remove bootnode reference, so it won't be included into bootnodes
				// don't return the error, node maybe not up and running yet
				node.Spec.Bootnodes = append(node.Spec.Bootnodes[:i], node.Spec.Bootnodes[i+1:]...)
				log.Error(err, "failed to get bootnode")
				continue
			}
			log.Info("bootnode enodeURL", string(enode), enodeURL)
			// replace reference with actual enode url
			if strings.HasPrefix(enodeURL, "enode://") {
				node.Spec.Bootnodes[i] = ethereumv1alpha1.Enode(enodeURL)
			} else {
				// remove bootnode reference, so it won't be included into bootnodes
				node.Spec.Bootnodes = append(node.Spec.Bootnodes[:i], node.Spec.Bootnodes[i+1:]...)
			}
		}
	}
}

// updateStatus updates network status
func (r *NodeReconciler) updateStatus(ctx context.Context, node *ethereumv1alpha1.Node, enodeURL string) error {
	var consensus, network string

	log := log.FromContext(ctx)

	if node.Spec.Genesis == nil {
		switch node.Spec.Network {
		case ethereumv1alpha1.MainNetwork,
			ethereumv1alpha1.RopstenNetwork,
			ethereumv1alpha1.XDaiNetwork,
			ethereumv1alpha1.GoerliNetwork:
			consensus = "pos"
		case ethereumv1alpha1.RinkebyNetwork:
			consensus = "poa"
		}
	} else {
		if node.Spec.Genesis.Ethash != nil {
			consensus = "pow"
		} else if node.Spec.Genesis.Clique != nil {
			consensus = "poa"
		} else if node.Spec.Genesis.IBFT2 != nil {
			consensus = "ibft2"
		}
	}

	node.Status.Consensus = consensus

	if network = node.Spec.Network; network == "" {
		network = "private"
	}

	node.Status.Network = network

	if node.Spec.NodePrivateKeySecretName == "" {
		switch node.Spec.Client {
		case ethereumv1alpha1.BesuClient:
			enodeURL = "call net_enode JSON-RPC method"
		case ethereumv1alpha1.GethClient:
			enodeURL = "call admin_nodeInfo JSON-RPC method"
		case ethereumv1alpha1.NethermindClient:
			enodeURL = "call net_localEnode JSON-RPC method"
		}
	}

	node.Status.EnodeURL = enodeURL

	if err := r.Status().Update(ctx, node); err != nil {
		log.Error(err, "unable to update node status")
		return err
	}

	return nil
}

// specConfigmap updates genesis configmap spec
func (r *NodeReconciler) specConfigmap(node *ethereumv1alpha1.Node, configmap *corev1.ConfigMap, genesis, staticNodes string) {
	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	var key, importAccountScript string

	switch node.Spec.Client {
	case ethereumv1alpha1.GethClient:
		key = "config.toml"
		importAccountScript = gethImportAccountScript
	case ethereumv1alpha1.BesuClient:
		key = "static-nodes.json"
	case ethereumv1alpha1.NethermindClient:
		key = "static-nodes.json"
	}

	if node.Spec.Genesis != nil {
		configmap.Data["genesis.json"] = genesis
		if node.Spec.Client == ethereumv1alpha1.GethClient {
			configmap.Data["geth-init-genesis.sh"] = GethInitGenesisScript
		}
	}

	if node.Spec.Import != nil {
		configmap.Data["import-account.sh"] = importAccountScript
	}

	if node.Spec.Client == ethereumv1alpha1.NethermindClient {
		configmap.Data["nethermind_convert_enode_privatekey.sh"] = nethermindConvertEnodePrivateKeyScript
		configmap.Data["nethermind_copy_keystore.sh"] = nethermindConvertCopyKeystoreScript
	}

	currentStaticNodes := configmap.Data[key]
	// update static nodes config if it's empty
	// update static nodes config if more static nodes has been created
	if currentStaticNodes == "" || len(currentStaticNodes) < len(staticNodes) {
		configmap.Data[key] = staticNodes
	}

	// create empty config for ptivate networks so it won't be ovverriden by
	if node.Spec.Client == ethereumv1alpha1.NethermindClient && node.Spec.Genesis != nil {
		configmap.Data["empty.cfg"] = "{}"
	}

}

// reconcileConfigmap creates genesis config map if it doesn't exist or update it
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *ethereumv1alpha1.Node) error {

	var genesis string

	log := log.FromContext(ctx)

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client, err := ethereumClients.NewClient(node)
	if err != nil {
		return err
	}

	staticNodes := client.EncodeStaticNodes()

	// private network with custom genesis
	if node.Spec.Genesis != nil {

		// create client specific genesis configuration
		if genesis, err = client.Genesis(); err != nil {
			return err
		}
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(node, configmap, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference on genesis configmap")
			return err
		}

		r.specConfigmap(node, configmap, genesis, staticNodes)

		return nil
	})

	return err
}

// specPVC update node data pvc spec
func (r *NodeReconciler) specPVC(node *ethereumv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = node.GetLabels()
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: node.Spec.Resources.StorageClass,
	}
}

// reconcilePVC creates node data pvc if it doesn't exist
func (r *NodeReconciler) reconcilePVC(ctx context.Context, node *ethereumv1alpha1.Node) error {

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(node, pvc, r.Scheme); err != nil {
			return err
		}
		r.specPVC(node, pvc)
		return nil
	})

	return err
}

// createNodeVolumes creates all the required volumes for the node
func (r *NodeReconciler) createNodeVolumes(node *ethereumv1alpha1.Node) []corev1.Volume {

	volumes := []corev1.Volume{}
	projections := []corev1.VolumeProjection{}

	// authenticated APIs jwt secret
	if node.Spec.JWTSecretName != "" {
		jwtSecretProjection := corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.JWTSecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "secret",
						Path: "jwt.secret",
					},
				},
			},
		}
		projections = append(projections, jwtSecretProjection)
	}

	// nodekey (node private key) projection
	if node.Spec.NodePrivateKeySecretName != "" {
		nodekeyProjection := corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.NodePrivateKeySecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "key",
						Path: "nodekey",
					},
				},
			},
		}
		projections = append(projections, nodekeyProjection)
	}

	// importing ethereum account
	if node.Spec.Import != nil {
		// account private key projection
		privateKeyProjection := corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.Import.PrivateKeySecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "key",
						Path: "account.key",
					},
				},
			},
		}
		projections = append(projections, privateKeyProjection)

		// account password projection
		passwordProjection := corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.Import.PasswordSecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "password",
						Path: "account.password",
					},
				},
			},
		}
		projections = append(projections, passwordProjection)

		// nethermind : account keystore
		if node.Spec.Client == ethereumv1alpha1.NethermindClient {
			accountKeystoreProjection := corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: node.Name,
					},
				},
			}
			projections = append(projections, accountKeystoreProjection)
		}
	}

	if len(projections) != 0 {
		secretsVolume := corev1.Volume{
			Name: "secrets",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: projections,
				},
			},
		}
		volumes = append(volumes, secretsVolume)
	}

	configVolume := corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Name,
				},
			},
		},
	}
	volumes = append(volumes, configVolume)

	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: node.Name,
			},
		},
	}
	volumes = append(volumes, dataVolume)

	return volumes
}

// createNodeVolumeMounts creates all required volume mounts for the node
func (r *NodeReconciler) createNodeVolumeMounts(node *ethereumv1alpha1.Node, homedir string) []corev1.VolumeMount {

	volumeMounts := []corev1.VolumeMount{}

	if node.Spec.NodePrivateKeySecretName != "" || node.Spec.Import != nil || node.Spec.JWTSecretName != "" {
		secretsMount := corev1.VolumeMount{
			Name:      "secrets",
			MountPath: shared.PathSecrets(homedir),
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, secretsMount)
	}

	configMount := corev1.VolumeMount{
		Name:      "config",
		MountPath: shared.PathConfig(homedir),
		ReadOnly:  true,
	}
	volumeMounts = append(volumeMounts, configMount)

	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: shared.PathData(homedir),
	}
	volumeMounts = append(volumeMounts, dataMount)

	return volumeMounts
}

// specStatefulset updates node statefulset spec
func (r *NodeReconciler) specStatefulset(node *ethereumv1alpha1.Node, sts *appsv1.StatefulSet, homedir string, args []string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	labels := node.GetLabels()
	// used by geth to init genesis and import account(s)
	initContainers := []corev1.Container{}
	// node client container
	nodeContainer := corev1.Container{
		Name:  "node",
		Image: node.Spec.Image,
		Args:  args,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPU),
				corev1.ResourceMemory: resource.MustParse(node.Spec.Resources.Memory),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPULimit),
				corev1.ResourceMemory: resource.MustParse(node.Spec.Resources.MemoryLimit),
			},
		},
		VolumeMounts: volumeMounts,
	}

	if node.Spec.Client == ethereumv1alpha1.GethClient {
		if node.Spec.Genesis != nil {
			initGenesis := corev1.Container{
				Name:  "init-geth-genesis",
				Image: node.Spec.Image,
				Env: []corev1.EnvVar{
					{
						Name:  shared.EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  shared.EnvConfigPath,
						Value: shared.PathConfig(homedir),
					},
				},
				Command:      []string{"/bin/sh"},
				Args:         []string{fmt.Sprintf("%s/geth-init-genesis.sh", shared.PathConfig(homedir))},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, initGenesis)
		}
		if node.Spec.Import != nil {
			importAccount := corev1.Container{
				Name:  "import-account",
				Image: node.Spec.Image,
				Env: []corev1.EnvVar{
					{
						Name:  shared.EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  shared.EnvSecretsPath,
						Value: shared.PathSecrets(homedir),
					},
				},
				Command:      []string{"/bin/sh"},
				Args:         []string{fmt.Sprintf("%s/import-account.sh", shared.PathConfig(homedir))},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, importAccount)
		}

	} else if node.Spec.Client == ethereumv1alpha1.NethermindClient {
		if node.Spec.NodePrivateKeySecretName != "" {
			convertEnodePrivateKey := corev1.Container{
				Name:  "convert-enode-privatekey",
				Image: shared.BusyboxImage,
				Env: []corev1.EnvVar{
					{
						Name:  shared.EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  shared.EnvSecretsPath,
						Value: shared.PathSecrets(homedir),
					},
				},
				Command:      []string{"/bin/sh"},
				Args:         []string{fmt.Sprintf("%s/nethermind_convert_enode_privatekey.sh", shared.PathConfig(homedir))},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, convertEnodePrivateKey)
		}

		if node.Spec.Import != nil {
			copyKeystore := corev1.Container{
				Name:  "copy-keystore",
				Image: shared.BusyboxImage,
				Env: []corev1.EnvVar{
					{
						Name:  shared.EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  shared.EnvSecretsPath,
						Value: shared.PathSecrets(homedir),
					},
					{
						Name:  envCoinbase,
						Value: strings.ToLower(string(node.Spec.Coinbase))[2:],
					},
				},
				Command:      []string{"/bin/sh"},
				Args:         []string{fmt.Sprintf("%s/nethermind_copy_keystore.sh", shared.PathConfig(homedir))},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, copyKeystore)
		}
	}

	sts.ObjectMeta.Labels = labels
	if sts.Spec.Selector == nil {
		sts.Spec.Selector = &metav1.LabelSelector{}
	}
	sts.Spec.ServiceName = node.Name
	sts.Spec.Selector.MatchLabels = labels
	sts.Spec.Template.ObjectMeta.Labels = labels
	sts.Spec.Template.Spec = corev1.PodSpec{
		SecurityContext: shared.SecurityContext(),
		Volumes:         volumes,
		InitContainers:  initContainers,
		Containers:      []corev1.Container{nodeContainer},
	}
}

// reconcileStatefulSet creates node statefulset if it doesn't exist, update it if it does exist
func (r *NodeReconciler) reconcileStatefulSet(ctx context.Context, node *ethereumv1alpha1.Node) error {

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client, err := ethereumClients.NewClient(node)
	if err != nil {
		return err
	}
	homedir := client.HomeDir()
	args := client.Args()
	volumes := r.createNodeVolumes(node)
	mounts := r.createNodeVolumeMounts(node, homedir)

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		r.specStatefulset(node, sts, homedir, args, volumes, mounts)
		return nil
	})

	return err
}

// specSecret creates keystore from account private key for nethermind client
func (r *NodeReconciler) specSecret(ctx context.Context, node *ethereumv1alpha1.Node, secret *corev1.Secret) error {
	secret.ObjectMeta.Labels = node.GetLabels()

	if node.Spec.Import != nil && node.Spec.Client == ethereumv1alpha1.NethermindClient {
		key := types.NamespacedName{
			Name:      node.Spec.Import.PrivateKeySecretName,
			Namespace: node.Namespace,
		}

		privateKey, err := shared.GetSecret(ctx, r.Client, key, "key")
		if err != nil {
			return err
		}

		key = types.NamespacedName{
			Name:      node.Spec.Import.PasswordSecretName,
			Namespace: node.Namespace,
		}

		password, err := shared.GetSecret(ctx, r.Client, key, "password")
		if err != nil {
			return err
		}

		account, err := KeyStoreFromPrivateKey(privateKey, password)
		if err != nil {
			return err
		}

		secret.Data = map[string][]byte{
			"account": account,
		}
	}

	return nil
}

// reconcileSecret creates node secret if it doesn't exist, update it if it exists
func (r *NodeReconciler) reconcileSecret(ctx context.Context, node *ethereumv1alpha1.Node) (publicKey string, err error) {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	// pubkey is required by the caller
	// 1. read the private key secret content
	// 2. derive public key from the private key
	if node.Spec.NodePrivateKeySecretName != "" {
		key := types.NamespacedName{
			Name:      node.Spec.NodePrivateKeySecretName,
			Namespace: node.Namespace,
		}

		var nodekey string
		nodekey, err = shared.GetSecret(ctx, r.Client, key, "key")
		if err != nil {
			return
		}

		// hex private key without the leading 0x
		publicKey, err = helpers.DerivePublicKey(nodekey)
		if err != nil {
			return
		}
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if err := ctrl.SetControllerReference(node, secret, r.Scheme); err != nil {
			return err
		}

		return r.specSecret(ctx, node, secret)
	})

	return
}

// specService updates node service spec
func (r *NodeReconciler) specService(node *ethereumv1alpha1.Node, svc *corev1.Service) {
	labels := node.GetLabels()
	client := node.Spec.Client

	svc.ObjectMeta.Labels = labels
	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "discovery",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	if node.Spec.RPC {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromInt(int(node.Spec.RPCPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.WS {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "ws",
			Port:       int32(node.Spec.WSPort),
			TargetPort: intstr.FromInt(int(node.Spec.WSPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.Engine {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "engine",
			Port:       int32(node.Spec.EnginePort),
			TargetPort: intstr.FromInt(int(node.Spec.EnginePort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.GraphQL {
		targetPort := node.Spec.GraphQLPort
		if client == ethereumv1alpha1.GethClient {
			targetPort = node.Spec.RPCPort
		}
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "graphql",
			Port:       int32(node.Spec.GraphQLPort),
			TargetPort: intstr.FromInt(int(targetPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc.Spec.Selector = labels
}

// reconcileService reconciles node service
func (r *NodeReconciler) reconcileService(ctx context.Context, node *ethereumv1alpha1.Node) (ip string, err error) {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err = ctrl.SetControllerReference(node, svc, r.Scheme); err != nil {
			return err
		}

		r.specService(node, svc)

		return nil
	})

	if err != nil {
		return
	}

	ip = svc.Spec.ClusterIP

	return
}

// SetupWithManager adds reconciler to the manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Node{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
