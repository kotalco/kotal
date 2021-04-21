package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=filecoin.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=filecoin.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles Filecoin network node
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node filecoinv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	if err = r.reconcileNodeService(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileNodePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileNodeStatefulSet(ctx, &node); err != nil {
		return
	}

	return
}

// reconcileNodePVC reconciles node pvc
func (r *NodeReconciler) reconcileNodePVC(ctx context.Context, node *filecoinv1alpha1.Node) error {
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
		r.specNodePVC(pvc, node)
		return nil
	})

	return err
}

// specNodePVC updates node PVC spec
func (r *NodeReconciler) specNodePVC(pvc *corev1.PersistentVolumeClaim, node *filecoinv1alpha1.Node) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

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

// reconcileNodeService reconciles node service
func (r *NodeReconciler) reconcileNodeService(ctx context.Context, node *filecoinv1alpha1.Node) error {

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
		r.specNodeService(svc, node)
		return nil
	})

	return err
}

// reconcileNodeStatefulSet reconciles node stateful set
func (r *NodeReconciler) reconcileNodeStatefulSet(ctx context.Context, node *filecoinv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specNodeStatefulSet(sts, node); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specNodeService updates node statefulset spec
func (r *NodeReconciler) specNodeService(svc *corev1.Service, node *filecoinv1alpha1.Node) {
	labels := map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "api",
			Port:       int32(1234),
			TargetPort: intstr.FromInt(1234),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	svc.Spec.Selector = labels
}

// specNodeStatefulSet updates node statefulset spec
func (r *NodeReconciler) specNodeStatefulSet(sts *appsv1.StatefulSet, node *filecoinv1alpha1.Node) error {
	labels := map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

	image, err := LotusImage(node.Spec.Network)
	if err != nil {
		return err
	}

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
				Containers: []corev1.Container{
					{
						Name:  "node",
						Image: image,
						Args:  []string{"daemon"},
						Env: []corev1.EnvVar{
							{
								Name:  "LOTUS_PATH",
								Value: "/mnt/data",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      node.Name,
								MountPath: "/mnt/data",
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
						Name: node.Name,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: node.Name,
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&filecoinv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
