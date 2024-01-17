package controllers

import (
	"context"
	_ "embed"
	"fmt"

	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	polkadotClients "github.com/kotalco/kotal/clients/polkadot"
	"github.com/kotalco/kotal/controllers/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	shared.Reconciler
}

var (
	//go:embed convert_node_private_key.sh
	convertNodePrivateKeyScript string
)

// +kubebuilder:rbac:groups=polkadot.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=polkadot.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node polkadotv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "polkadot", node.Spec.Network)

	// reconcile config map
	if err = r.ReconcileOwned(ctx, &node, &corev1.ConfigMap{}, func(obj client.Object) error {
		r.specConfigmap(&node, obj.(*corev1.ConfigMap))
		return nil
	}); err != nil {
		return
	}

	// reconcile persistent volume claim
	if err = r.ReconcileOwned(ctx, &node, &corev1.PersistentVolumeClaim{}, func(obj client.Object) error {
		r.specPVC(&node, obj.(*corev1.PersistentVolumeClaim))
		return nil
	}); err != nil {
		return
	}

	// reconcile service
	if err = r.ReconcileOwned(ctx, &node, &corev1.Service{}, func(obj client.Object) error {
		r.specService(&node, obj.(*corev1.Service))
		return nil
	}); err != nil {
		return
	}

	// reconcile stateful set
	if err = r.ReconcileOwned(ctx, &node, &appsv1.StatefulSet{}, func(obj client.Object) error {
		client := polkadotClients.NewClient(&node)
		args := client.Args()
		args = append(args, node.Spec.ExtraArgs.Encode(false)...)
		homeDir := client.HomeDir()

		return r.specStatefulSet(&node, obj.(*appsv1.StatefulSet), homeDir, args)
	}); err != nil {
		return
	}

	return
}

// specConfigmap updates polkadot node configmap spec
func (r *NodeReconciler) specConfigmap(node *polkadotv1alpha1.Node, config *corev1.ConfigMap) {
	config.ObjectMeta.Labels = node.Labels

	if config.Data == nil {
		config.Data = make(map[string]string)
	}

	config.Data["convert_node_private_key.sh"] = convertNodePrivateKeyScript
}

// specService updates polkadot node service spec
func (r *NodeReconciler) specService(node *polkadotv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromString("p2p"),
		},
	}

	if node.Spec.Prometheus {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "prometheus",
			Port:       int32(node.Spec.PrometheusPort),
			TargetPort: intstr.FromString("prometheus"),
		})
	}

	if node.Spec.RPC {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromString("rpc"),
		})
	}

	if node.Spec.WS {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "ws",
			Port:       int32(node.Spec.WSPort),
			TargetPort: intstr.FromString("ws"),
		})
	}

	svc.Spec.Selector = labels
}

// nodeVolumes returns node volumes
func (r *NodeReconciler) nodeVolumes(node *polkadotv1alpha1.Node) (volumes []corev1.Volume) {
	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: node.Name,
			},
		},
	}
	volumes = append(volumes, dataVolume)

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

	if node.Spec.NodePrivateKeySecretName != "" {
		secretVolume := corev1.Volume{
			Name: "secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: node.Spec.NodePrivateKeySecretName,
					Items: []corev1.KeyToPath{
						{
							Key:  "key",
							Path: "nodekey",
						},
					},
				},
			},
		}
		volumes = append(volumes, secretVolume)
	}

	return
}

// nodeVolumeMounts returns node volume mounts
func (r *NodeReconciler) nodeVolumeMounts(node *polkadotv1alpha1.Node, homeDir string) (mounts []corev1.VolumeMount) {
	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: shared.PathData(homeDir),
	}
	mounts = append(mounts, dataMount)

	configMount := corev1.VolumeMount{
		Name:      "config",
		MountPath: shared.PathConfig(homeDir),
	}
	mounts = append(mounts, configMount)

	if node.Spec.NodePrivateKeySecretName != "" {
		secretMount := corev1.VolumeMount{
			Name:      "secret",
			MountPath: shared.PathSecrets(homeDir),
		}
		mounts = append(mounts, secretMount)
	}
	return
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *polkadotv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, args []string) error {

	sts.ObjectMeta.Labels = node.Labels

	var initContainers []corev1.Container

	if node.Spec.NodePrivateKeySecretName != "" {
		convertEnodePrivateKey := corev1.Container{
			Name:  "convert-node-private-key",
			Image: shared.BusyboxImage,
			Env: []corev1.EnvVar{
				{
					Name:  shared.EnvDataPath,
					Value: shared.PathData(homeDir),
				},
				{
					Name:  shared.EnvSecretsPath,
					Value: shared.PathSecrets(homeDir),
				},
			},
			Command:      []string{"/bin/sh"},
			Args:         []string{fmt.Sprintf("%s/convert_node_private_key.sh", shared.PathConfig(homeDir))},
			VolumeMounts: r.nodeVolumeMounts(node, homeDir),
		}
		initContainers = append(initContainers, convertEnodePrivateKey)
	}

	ports := []corev1.ContainerPort{
		{
			Name:          "p2p",
			ContainerPort: int32(node.Spec.P2PPort),
		},
	}

	if node.Spec.Prometheus {
		ports = append(ports, corev1.ContainerPort{
			Name:          "prometheus",
			ContainerPort: int32(node.Spec.PrometheusPort),
		})
	}

	if node.Spec.RPC {
		ports = append(ports, corev1.ContainerPort{
			Name:          "rpc",
			ContainerPort: int32(node.Spec.RPCPort),
		})
	}

	if node.Spec.WS {
		ports = append(ports, corev1.ContainerPort{
			Name:          "ws",
			ContainerPort: int32(node.Spec.RPCPort),
		})
	}

	replicas := int32(*node.Spec.Replicas)

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.Labels,
		},
		ServiceName: node.Name,
		Replicas:    &replicas,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: node.Labels,
			},
			Spec: corev1.PodSpec{
				InitContainers:  initContainers,
				SecurityContext: shared.SecurityContext(),
				Containers: []corev1.Container{
					{
						Name:         "node",
						Image:        node.Spec.Image,
						Args:         args,
						Ports:        ports,
						VolumeMounts: r.nodeVolumeMounts(node, homeDir),
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.CPU),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Memory),
							},
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.CPULimit),
								corev1.ResourceMemory: resource.MustParse(node.Spec.MemoryLimit),
							},
						},
					},
				},
				Volumes: r.nodeVolumes(node),
			},
		},
	}

	return nil
}

// specPVC updates ipfs peer persistent volume claim
func (r *NodeReconciler) specPVC(node *polkadotv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = node.Labels
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.VolumeResourceRequirements{
			Requests: request,
		},
		StorageClassName: node.Spec.StorageClass,
	}
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&polkadotv1alpha1.Node{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
