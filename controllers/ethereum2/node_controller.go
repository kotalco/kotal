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

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconcile Ethereum 2.0 node
func (r *NodeReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var node ethereum2v1alpha1.Node

	if err = r.Client.Get(context.Background(), req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&node)

	if err = r.reconcileNodeStatefulset(&node); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the node
func (r *NodeReconciler) updateLabels(node *ethereum2v1alpha1.Node) {

	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	// TODO: update with client
	node.Labels["name"] = "node"
	node.Labels["protocol"] = "ethereum2"
	node.Labels["instance"] = node.Name
}

// reconcileNodeStatefulset reconcile Ethereum 2.0 node
func (r *NodeReconciler) reconcileNodeStatefulset(node *ethereum2v1alpha1.Node) error {
	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client, err := NewEthereum2Client(node.Spec.Client)
	if err != nil {
		return err
	}

	args := client.GetArgs(node)

	_, err = ctrl.CreateOrUpdate(context.Background(), r.Client, &sts, func() error {
		if err := ctrl.SetControllerReference(node, &sts, r.Scheme); err != nil {
			return err
		}
		r.specNodeStatefulset(&sts, node, args)
		return nil
	})

	return err
}

func (r *NodeReconciler) specNodeStatefulset(sts *appsv1.StatefulSet, node *ethereum2v1alpha1.Node, args []string) {
	sts.Labels = node.GetLabels()
	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.GetLabels(),
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: node.GetLabels(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "node",
						// TODO: move image to TekuImage()
						Image: "consensys/teku:latest",
						Args:  args,
					},
				},
			},
		},
	}
}

// SetupWithManager adds reconciler to the manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereum2v1alpha1.Node{}).
		Complete(r)
}
