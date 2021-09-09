package controllers

import (
	"context"

	"github.com/go-logr/logr"
	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	polkadotClients "github.com/kotalco/kotal/clients/polkadot"
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
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=polkadot.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=polkadot.kotal.io,resources=nodes/status,verbs=get;update;patch

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node polkadotv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&node)

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the node
func (r *NodeReconciler) updateLabels(node *polkadotv1alpha1.Node) {

	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	node.Labels["app.kubernetes.io/name"] = "polkadot"
	node.Labels["app.kubernetes.io/instance"] = node.Name
	node.Labels["app.kubernetes.io/component"] = "polkadot-node"
	node.Labels["app.kubernetes.io/managed-by"] = "kotal"
	node.Labels["app.kubernetes.io/created-by"] = "polkadot-node-controller"

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

	img := client.Image()
	args := client.Args()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specStatefulSet(node, sts, img, args); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *polkadotv1alpha1.Node, sts *appsv1.StatefulSet, image string, args []string) error {

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
						Name:  "node",
						Image: image,
						Args:  args,
					},
				},
			},
		},
	}

	return nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&polkadotv1alpha1.Node{}).
		Complete(r)
}
