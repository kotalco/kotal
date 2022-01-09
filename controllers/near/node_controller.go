package controllers

import (
	"context"

	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=near.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=near.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node nearv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// TODO: default the node if webhooks are disabled

	shared.UpdateLabels(&node, "nearcore")

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *nearv1alpha1.Node) error {
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
		r.specStatefulSet(node, sts)
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *nearv1alpha1.Node, sts *appsv1.StatefulSet) {

	sts.ObjectMeta.Labels = node.Labels

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
				// TODO: use shared security context
				Containers: []corev1.Container{
					{
						Name:  "node",
						Image: "nearprotocol/nearup",
						Args:  []string{"run", node.Spec.Network},
						// TODO: mount data pvc
						VolumeMounts: []corev1.VolumeMount{},
					},
				},
				// TODO: use persistent volume claim
				Volumes: []corev1.Volume{
					{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				},
			},
		},
	}

}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nearv1alpha1.Node{}).
		Complete(r)
}
