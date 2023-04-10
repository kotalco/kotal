package controllers

import (
	"context"
	_ "embed"
	"fmt"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	filecoinClients "github.com/kotalco/kotal/clients/filecoin"
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

var (
	//go:embed copy_config_toml.sh
	CopyConfigToml string
)

// +kubebuilder:rbac:groups=filecoin.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=filecoin.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;services;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles Filecoin network node
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node filecoinv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "lotus", string(node.Spec.Network))

	if err = r.reconcileService(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulSet(ctx, &node); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &node); err != nil {
		return
	}

	return
}

// updateStatus updates filecoin node status
func (r *NodeReconciler) updateStatus(ctx context.Context, node *filecoinv1alpha1.Node) error {
	node.Status.Client = "lotus"

	if err := r.Status().Update(ctx, node); err != nil {
		log.FromContext(ctx).Error(err, "unable to update filecoin node status")
		return err
	}

	return nil
}

// reconcilePVC reconciles node pvc
func (r *NodeReconciler) reconcilePVC(ctx context.Context, node *filecoinv1alpha1.Node) error {
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

// specPVC updates node PVC spec
func (r *NodeReconciler) specPVC(node *filecoinv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Resources.Storage),
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
		StorageClassName: node.Spec.Resources.StorageClass,
	}
}

// reconcileService reconciles node service
func (r *NodeReconciler) reconcileService(ctx context.Context, node *filecoinv1alpha1.Node) error {

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

// specConfigmap updates node statefulset spec
func (r *NodeReconciler) specConfigmap(node *filecoinv1alpha1.Node, configmap *corev1.ConfigMap, configToml string) {
	configmap.ObjectMeta.Labels = node.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	configmap.Data["config.toml"] = configToml
	configmap.Data["copy_config_toml.sh"] = CopyConfigToml

}

// reconcileConfigmap reconciles node configmap
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *filecoinv1alpha1.Node) error {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	// filecoin generates config.toml file from node spec
	configToml, err := ConfigFromSpec(node)
	if err != nil {
		return err
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(node, configmap, r.Scheme); err != nil {
			return err
		}
		r.specConfigmap(node, configmap, configToml)
		return nil
	})

	return err
}

// reconcileStatefulSet reconciles node stateful set
func (r *NodeReconciler) reconcileStatefulSet(ctx context.Context, node *filecoinv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := filecoinClients.NewClient(node)
	args := client.Args()
	env := client.Env()
	cmd := client.Command()
	homeDir := client.HomeDir()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specStatefulSet(node, sts, homeDir, cmd, args, env); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specService updates node statefulset spec
func (r *NodeReconciler) specService(node *filecoinv1alpha1.Node, svc *corev1.Service) {
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

	if node.Spec.API {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "api",
			Port:       int32(node.Spec.APIPort),
			TargetPort: intstr.FromInt(int(node.Spec.APIPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc.Spec.Selector = labels
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *filecoinv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, cmd, args []string, env []corev1.EnvVar) error {
	labels := node.Labels

	sts.ObjectMeta.Labels = labels

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		ServiceName: node.Name,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				SecurityContext: shared.SecurityContext(),
				InitContainers: []corev1.Container{
					{
						Name:  "copy-config-toml",
						Image: shared.BusyboxImage,
						Env: []corev1.EnvVar{
							{
								Name:  shared.EnvDataPath,
								Value: shared.PathData(homeDir),
							},
							{
								Name:  shared.EnvConfigPath,
								Value: shared.PathConfig(homeDir),
							},
						},
						Command: []string{"/bin/sh"},
						Args:    []string{fmt.Sprintf("%s/copy_config_toml.sh", shared.PathConfig(homeDir))},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: shared.PathData(homeDir),
							},
							{
								Name:      "config",
								MountPath: shared.PathConfig(homeDir),
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   node.Spec.Image,
						Args:    args,
						Command: cmd,
						Env:     env,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: shared.PathData(homeDir),
							},
						},
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
					},
				},
				Volumes: []corev1.Volume{
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
				},
			},
		},
	}

	return nil
}

// SetupWithManager adds reconciler to the manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&filecoinv1alpha1.Node{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
