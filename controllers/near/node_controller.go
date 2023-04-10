package controllers

import (
	"context"
	_ "embed"
	"fmt"

	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	nearClients "github.com/kotalco/kotal/clients/near"
	"github.com/kotalco/kotal/controllers/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	envNetwork = "KOTAL_NEAR_NETWORK"
)

var (
	//go:embed init_near_node.sh
	InitNearNode string
	//go:embed copy_node_key.sh
	CopyNodeKey string
	//go:embed copy_validator_key.sh
	CopyValidatorKey string
)

// +kubebuilder:rbac:groups=near.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=near.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;persistentvolumeclaims;services,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node nearv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "nearcore", node.Spec.Network)

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileService(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &node); err != nil {
		return
	}

	return
}

// updateStatus updates NEAR node status
func (r *NodeReconciler) updateStatus(ctx context.Context, peer *nearv1alpha1.Node) error {
	peer.Status.Client = "nearcore"

	if err := r.Status().Update(ctx, peer); err != nil {
		log.FromContext(ctx).Error(err, "unable to update node status")
		return err
	}

	return nil
}

// reconcileService reconciles NEAR node service
func (r *NodeReconciler) reconcileService(ctx context.Context, node *nearv1alpha1.Node) error {
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

// specService updates NEAR node service spec
func (r *NodeReconciler) specService(node *nearv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       "discovery",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolUDP,
		},
	}

	if node.Spec.RPC {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromInt(int(node.Spec.RPCPort)),
			Protocol:   corev1.ProtocolTCP,
		})
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "prometheus",
			Port:       int32(node.Spec.PrometheusPort),
			TargetPort: intstr.FromInt(int(node.Spec.PrometheusPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc.Spec.Selector = labels
}

// reconcilePVC reconciles NEAR node persistent volume claim
func (n *NodeReconciler) reconcilePVC(ctx context.Context, node *nearv1alpha1.Node) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, n.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(node, pvc, n.Scheme); err != nil {
			return err
		}

		n.specPVC(node, pvc)

		return nil
	})

	return err
}

// specPVC updates NEAR node persistent volume claim
func (n *NodeReconciler) specPVC(peer *nearv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(peer.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = peer.Labels
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: peer.Spec.Resources.StorageClass,
	}
}

// specConfigmap updates node configmap
func (n *NodeReconciler) specConfigmap(node *nearv1alpha1.Node, configmap *corev1.ConfigMap) {
	configmap.ObjectMeta.Labels = node.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	configmap.Data["init_near_node.sh"] = InitNearNode
	configmap.Data["copy_node_key.sh"] = CopyNodeKey
	configmap.Data["copy_validator_key.sh"] = CopyValidatorKey

}

// reconcileConfigmap reconciles node configmap
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *nearv1alpha1.Node) error {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(node, configmap, r.Scheme); err != nil {
			return err
		}
		r.specConfigmap(node, configmap)
		return nil
	})

	return err
}

func (r *NodeReconciler) createVolumes(node *nearv1alpha1.Node) []corev1.Volume {

	var volumeProjections []corev1.VolumeProjection

	volumes := []corev1.Volume{
		{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: node.Name,
				},
			},
		},
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: node.Name,
					},
				},
			},
		},
	}

	if node.Spec.NodePrivateKeySecretName != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.NodePrivateKeySecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "key",
						Path: "node_key.json",
					},
				},
			},
		})
	}

	if node.Spec.ValidatorSecretName != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.ValidatorSecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "key",
						Path: "validator_key.json",
					},
				},
			},
		})
	}

	secretsVolume := corev1.Volume{
		Name: "secrets",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: volumeProjections,
			},
		},
	}
	volumes = append(volumes, secretsVolume)

	return volumes
}

func (r *NodeReconciler) createVolumeMounts(node *nearv1alpha1.Node, homeDir string) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: shared.PathData(homeDir),
		},
		{
			Name:      "config",
			MountPath: shared.PathConfig(homeDir),
		},
	}

	if node.Spec.NodePrivateKeySecretName != "" || node.Spec.ValidatorSecretName != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "secrets",
			MountPath: shared.PathSecrets(homeDir),
		})
	}

	return mounts
}

func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *nearv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := nearClients.NewClient(node)

	homeDir := client.HomeDir()
	args := client.Args()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		r.specStatefulSet(node, sts, homeDir, args)
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *nearv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, args []string) {

	sts.ObjectMeta.Labels = node.Labels

	initContainers := []corev1.Container{
		{
			Name:  "init-near-node",
			Image: node.Spec.Image,
			Env: []corev1.EnvVar{
				{
					Name:  shared.EnvDataPath,
					Value: shared.PathData(homeDir),
				},
				{
					Name:  envNetwork,
					Value: node.Spec.Network,
				},
			},
			Command:      []string{"/bin/sh"},
			Args:         []string{fmt.Sprintf("%s/init_near_node.sh", shared.PathConfig(homeDir))},
			VolumeMounts: r.createVolumeMounts(node, homeDir),
		},
	}

	if node.Spec.NodePrivateKeySecretName != "" {
		initContainers = append(initContainers, corev1.Container{
			Name:    "copy-node-key",
			Image:   shared.BusyboxImage,
			Command: []string{"/bin/sh"},
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
			Args:         []string{fmt.Sprintf("%s/copy_node_key.sh", shared.PathConfig(homeDir))},
			VolumeMounts: r.createVolumeMounts(node, homeDir),
		})
	}

	if node.Spec.ValidatorSecretName != "" {
		initContainers = append(initContainers, corev1.Container{
			Name:    "copy-validator-key",
			Image:   shared.BusyboxImage,
			Command: []string{"/bin/sh"},
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
			Args:         []string{fmt.Sprintf("%s/copy_validator_key.sh", shared.PathConfig(homeDir))},
			VolumeMounts: r.createVolumeMounts(node, homeDir),
		})
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
				SecurityContext: shared.SecurityContext(),
				InitContainers:  initContainers,
				Containers: []corev1.Container{
					{
						Name:         "node",
						Image:        node.Spec.Image,
						Args:         args,
						VolumeMounts: r.createVolumeMounts(node, homeDir),
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.CPU),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.CPULimit),
								corev1.ResourceMemory: resource.MustParse(node.Spec.MemoryLimit),
							},
						},
					},
				},
				Volumes: r.createVolumes(node),
			},
		},
	}

}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&nearv1alpha1.Node{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
