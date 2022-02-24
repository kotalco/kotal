package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=bitcoin.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bitcoin.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete

// Reconcile Bitcoin node
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node bitcoinv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "bitcoind")

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// reconcileStatefulset reconciles node statefulset
func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *bitcoinv1alpha1.Node) error {
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
		if err := r.specStatefulSet(node, sts); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *bitcoinv1alpha1.Node, sts *appsv1.StatefulSet) error {

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
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   "ruimarinho/bitcoin-core",
						Command: []string{"bitcoind"},
						Args:    []string{"-chain=test"},
					},
				},
			},
		},
	}

	return nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bitcoinv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
