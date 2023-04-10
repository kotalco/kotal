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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileService(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// reconcileConfigmap reconciles polkadot node configmap
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *polkadotv1alpha1.Node) error {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, config, func() error {
		if err := ctrl.SetControllerReference(node, config, r.Scheme); err != nil {
			return err
		}

		r.specConfigmap(node, config)

		return nil
	})

	return err

}

// specConfigmap updates polkadot node configmap spec
func (r *NodeReconciler) specConfigmap(node *polkadotv1alpha1.Node, config *corev1.ConfigMap) {
	config.ObjectMeta.Labels = node.Labels

	if config.Data == nil {
		config.Data = make(map[string]string)
	}

	config.Data["convert_node_private_key.sh"] = convertNodePrivateKeyScript
}

// reconcileService reconciles polkadot node service
func (r *NodeReconciler) reconcileService(ctx context.Context, node *polkadotv1alpha1.Node) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err := ctrl.SetControllerReference(node, svc, r.Scheme); err != nil {
			return err
		}
		r.specService(node, svc)
		return nil
	})

	return err
}

// specService updates polkadot node service spec
func (r *NodeReconciler) specService(node *polkadotv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	if node.Spec.Prometheus {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "prometheus",
			Port:       int32(node.Spec.PrometheusPort),
			TargetPort: intstr.FromInt(int(node.Spec.PrometheusPort)),
			Protocol:   corev1.ProtocolTCP,
		})
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

	svc.Spec.Selector = labels
}

// reconcileStatefulset reconciles node statefulset
func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *polkadotv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := polkadotClients.NewClient(node)

	args := client.Args()
	homeDir := client.HomeDir()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specStatefulSet(node, sts, homeDir, args); err != nil {
			return err
		}
		return nil
	})

	return err
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

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.Labels,
		},
		ServiceName: node.Name,
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

// reconcilePVC reconciles polkadot node persistent volume claim
func (r *NodeReconciler) reconcilePVC(ctx context.Context, node *polkadotv1alpha1.Node) error {
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
		Resources: corev1.ResourceRequirements{
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
