package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	ethereumClients "github.com/kotalco/kotal/clients/ethereum"
	"github.com/kotalco/kotal/controllers/shared"
	"github.com/kotalco/kotal/helpers"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	//go:embed geth_init_genesis.sh
	GethInitGenesisScript string
	//go:embed geth_import_account.sh
	gethImportAccountScript string
	//go:embed parity_import_account.sh
	parityImportAccountScript string
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

	var node ethereumv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	r.updateLabels(&node)
	r.updateStaticNodes(ctx, &node)

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

// updateLabels adds missing labels to the node
func (r *NodeReconciler) updateLabels(node *ethereumv1alpha1.Node) {

	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	node.Labels["app.kubernetes.io/name"] = string(node.Spec.Client)
	node.Labels["app.kubernetes.io/instance"] = node.Name
	node.Labels["app.kubernetes.io/component"] = "ethereum-node"
	node.Labels["app.kubernetes.io/managed-by"] = "kotal"
	node.Labels["app.kubernetes.io/created-by"] = "ethereum-node-controller"

}

// updateStaticNodes replaces Ethereum node references with their enodeURL
func (r *NodeReconciler) updateStaticNodes(ctx context.Context, node *ethereumv1alpha1.Node) {
	for i, enode := range node.Spec.StaticNodes {
		if !strings.HasPrefix(string(enode), "enode://") {
			staticNode := &ethereumv1alpha1.Node{}
			var name, namespace string
			// enode reference can have the format name.namespace
			// name is the node name
			// namespace is the node namespace
			if parts := strings.Split(string(enode), "."); len(parts) > 1 {
				name = parts[0]
				namespace = parts[1]
			} else {
				// nodes without . refered to nodes in the current node namespace
				name = string(enode)
				namespace = node.Namespace
			}
			namespacedName := types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}
			if err := r.Client.Get(ctx, namespacedName, staticNode); err != nil {
				// remove static node reference, so it won't be included into static nodes file
				node.Spec.StaticNodes = append(node.Spec.StaticNodes[:i], node.Spec.StaticNodes[i+1:]...)
				r.Log.Error(err, "failed to get static node")
				// don't return the error
				// node maybe not up and running yet
				continue
			}
			staticNodeURL := staticNode.Status.EnodeURL
			r.Log.Info("static node URL", string(enode), staticNodeURL)
			// replace reference with actual enode url
			if strings.HasPrefix(staticNodeURL, "enode://") {
				node.Spec.StaticNodes[i] = ethereumv1alpha1.Enode(staticNodeURL)
			} else {
				// remove static node reference, so it won't be included into static nodes file
				node.Spec.StaticNodes = append(node.Spec.StaticNodes[:i], node.Spec.StaticNodes[i+1:]...)
			}
		}
	}
}

// updateStatus updates network status
func (r *NodeReconciler) updateStatus(ctx context.Context, node *ethereumv1alpha1.Node, enodeURL string) error {

	if node.Spec.NodekeySecretName == "" {
		switch node.Spec.Client {
		case ethereumv1alpha1.BesuClient:
			enodeURL = "call net_enode JSON-RPC method"
		case ethereumv1alpha1.GethClient:
			enodeURL = "call admin_nodeInfo JSON-RPC method"
		case ethereumv1alpha1.ParityClient:
			enodeURL = "call parity_enode JSON-RPC method"
		case ethereumv1alpha1.NethermindClient:
			enodeURL = "call net_localEnode JSON-RPC method"
		}
	}

	node.Status.EnodeURL = enodeURL

	if err := r.Status().Update(ctx, node); err != nil {
		r.Log.Error(err, "unable to update node status")
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
	case ethereumv1alpha1.ParityClient:
		key = "static-nodes"
		importAccountScript = parityImportAccountScript
	case ethereumv1alpha1.NethermindClient:
		key = "static-nodes.json"
	}

	configmap.Data["genesis.json"] = genesis
	configmap.Data["geth-init-genesis.sh"] = GethInitGenesisScript
	configmap.Data["import-account.sh"] = importAccountScript
	configmap.Data["nethermind_convert_enode_privatekey.sh"] = nethermindConvertEnodePrivateKeyScript
	configmap.Data["nethermind_copy_keystore.sh"] = nethermindConvertCopyKeystoreScript

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
			r.Log.Error(err, "Unable to set controller reference on genesis configmap")
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

	// nodekey (node private key) projection
	if node.Spec.NodekeySecretName != "" {
		nodekeyProjection := corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.NodekeySecretName,
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

		// parity & nethermind : account keystore
		if node.Spec.Client == ethereumv1alpha1.ParityClient || node.Spec.Client == ethereumv1alpha1.NethermindClient {
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

	if node.Spec.NodekeySecretName != "" || node.Spec.Import != nil {
		nodekeyMount := corev1.VolumeMount{
			Name:      "secrets",
			MountPath: shared.PathSecrets(homedir),
			ReadOnly:  true,
		}
		volumeMounts = append(volumeMounts, nodekeyMount)
	}

	genesisMount := corev1.VolumeMount{
		Name:      "config",
		MountPath: shared.PathConfig(homedir),
		ReadOnly:  true,
	}
	volumeMounts = append(volumeMounts, genesisMount)

	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: shared.PathData(homedir),
	}
	volumeMounts = append(volumeMounts, dataMount)

	return volumeMounts
}

// getNodeAffinity returns affinity settings to be use by the node pod
func (r *NodeReconciler) getNodeAffinity(node *ethereumv1alpha1.Node) *corev1.Affinity {
	if node.Spec.HighlyAvailable {
		return &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"name":    "node",
								"network": node.Name,
							},
						},
						TopologyKey: node.Spec.TopologyKey,
					},
				},
			},
		}
	}
	return nil
}

// specStatefulset updates node statefulset spec
func (r *NodeReconciler) specStatefulset(node *ethereumv1alpha1.Node, sts *appsv1.StatefulSet, img, homedir string, args []string, volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, affinity *corev1.Affinity) {
	labels := node.GetLabels()
	// used by geth to init genesis and import account(s)
	initContainers := []corev1.Container{}
	// node client container
	nodeContainer := corev1.Container{
		Name:  "node",
		Image: img,
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
				Image: img,
				Env: []corev1.EnvVar{
					{
						Name:  EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  EnvConfigPath,
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
				Image: img,
				Env: []corev1.EnvVar{
					{
						Name:  EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  EnvSecretsPath,
						Value: shared.PathSecrets(homedir),
					},
				},
				Command:      []string{"/bin/sh"},
				Args:         []string{fmt.Sprintf("%s/import-account.sh", shared.PathConfig(homedir))},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, importAccount)
		}

	} else if node.Spec.Client == ethereumv1alpha1.ParityClient {
		if node.Spec.Import != nil {
			importAccount := corev1.Container{
				Name:  "import-account",
				Image: img,
				Env: []corev1.EnvVar{
					{
						Name:  EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  EnvConfigPath,
						Value: shared.PathConfig(homedir),
					},
					{
						Name:  EnvSecretsPath,
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
		if node.Spec.NodekeySecretName != "" {
			convertEnodePrivatekey := corev1.Container{
				Name:  "convert-enode-privatekey",
				Image: "busybox",
				Env: []corev1.EnvVar{
					{
						Name:  EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  EnvSecretsPath,
						Value: shared.PathSecrets(homedir),
					},
				},
				Command:      []string{"/bin/sh"},
				Args:         []string{fmt.Sprintf("%s/nethermind_convert_enode_privatekey.sh", shared.PathConfig(homedir))},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, convertEnodePrivatekey)
		}

		if node.Spec.Import != nil {
			copyKeystore := corev1.Container{
				Name:  "copy-keystore",
				Image: "busybox",
				Env: []corev1.EnvVar{
					{
						Name:  EnvDataPath,
						Value: shared.PathData(homedir),
					},
					{
						Name:  EnvSecretsPath,
						Value: shared.PathSecrets(homedir),
					},
					{
						Name:  "COINBASE",
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
		Affinity:        affinity,
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
	img := client.Image()
	homedir := client.HomeDir()
	args := client.Args()
	volumes := r.createNodeVolumes(node)
	mounts := r.createNodeVolumeMounts(node, homedir)
	affinity := r.getNodeAffinity(node)

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		r.specStatefulset(node, sts, img, homedir, args, volumes, mounts, affinity)
		return nil
	})

	return err
}

func (r *NodeReconciler) getSecret(ctx context.Context, name types.NamespacedName, key string) (value string, err error) {
	secret := &corev1.Secret{}

	if err = r.Client.Get(ctx, name, secret); err != nil {
		return
	}

	value = string(secret.Data[key])

	return
}

// specSecret creates keystore from account private key for parity client
func (r *NodeReconciler) specSecret(ctx context.Context, node *ethereumv1alpha1.Node, secret *corev1.Secret) error {
	secret.ObjectMeta.Labels = node.GetLabels()
	client := node.Spec.Client
	clientRequiresKeystore := client == ethereumv1alpha1.ParityClient || client == ethereumv1alpha1.NethermindClient
	if node.Spec.Import != nil && clientRequiresKeystore {
		key := types.NamespacedName{
			Name:      node.Spec.Import.PrivateKeySecretName,
			Namespace: node.Namespace,
		}

		privateKey, err := r.getSecret(ctx, key, "key")
		if err != nil {
			return err
		}

		key = types.NamespacedName{
			Name:      node.Spec.Import.PasswordSecretName,
			Namespace: node.Namespace,
		}

		password, err := r.getSecret(ctx, key, "password")
		if err != nil {
			return err
		}

		account, err := KeyStoreFromPrivatekey(privateKey, password)
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
	if node.Spec.NodekeySecretName != "" {
		key := types.NamespacedName{
			Name:      node.Spec.NodekeySecretName,
			Namespace: node.Namespace,
		}

		var nodekey string
		nodekey, err = r.getSecret(ctx, key, "key")
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

	if node.Spec.RPCPort != 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "json-rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromInt(int(node.Spec.RPCPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.WSPort != 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "ws",
			Port:       int32(node.Spec.WSPort),
			TargetPort: intstr.FromInt(int(node.Spec.WSPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.GraphQLPort != 0 {
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
