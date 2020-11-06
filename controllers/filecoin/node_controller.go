package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
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

// Reconcile reconciles Filecoin network node
func (r *NodeReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var node filecoinv1alpha1.Node

	if err = r.Client.Get(context.Background(), req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.reconcileNodeStatefulSet(&node); err != nil {
		return
	}

	return
}

// reconcileNodeStatefulSet reconciles node stateful set
func (r *NodeReconciler) reconcileNodeStatefulSet(node *filecoinv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, sts, func() error {
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
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "node",
						Image: image,
						Args:  []string{"daemon"},
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
		Complete(r)
}
