package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// NetworkReconciler reconciles a Network object
type NetworkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles ethereum networks
func (r *NetworkReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {

	var network ethereumv1alpha1.Network

	// Get desired ethereum network
	if err = r.Client.Get(context.Background(), req.NamespacedName, &network); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// update network status
	if err = r.updateStatus(&network); err != nil {
		return
	}

	// reconcile network nodes
	if err = r.reconcileNodes(&network); err != nil {
		return
	}

	return

}

// updateStatus updates network status
// TODO: don't update statuse on network deletion
func (r *NetworkReconciler) updateStatus(network *ethereumv1alpha1.Network) error {
	network.Status.NodesCount = len(network.Spec.Nodes)

	if err := r.Status().Update(context.Background(), network); err != nil {
		r.Log.Error(err, "unable to update network status")
		return err
	}

	return nil
}

// reconcileNodes creates or updates nodes according to nodes spec
// deletes nodes missing from nodes spec
func (r *NetworkReconciler) reconcileNodes(network *ethereumv1alpha1.Network) error {

	var staticNodes []string

	for i := range network.Spec.Nodes {
		enodeURL, err := r.reconcileNode(network, network.Spec.Nodes[i], staticNodes)
		if err != nil {
			return err
		}
		if enodeURL != "" {
			staticNodes = append(staticNodes, enodeURL)
		}
	}

	if err := r.deleteRedundantNodes(network); err != nil {
		return err
	}

	return nil
}

// reconcileNode reconciles a single etheruem node from within the network.spec.nodes
func (r *NetworkReconciler) reconcileNode(network *ethereumv1alpha1.Network, spec ethereumv1alpha1.NetworkNodeSpec, staticNodes []string) (enodeURL string, err error) {
	node := ethereumv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", network.Name, spec.Name),
			Namespace: network.Namespace,
		},
	}

	ctrl.CreateOrUpdate(context.Background(), r.Client, &node, func() error {
		if err := ctrl.SetControllerReference(network, &node, r.Scheme); err != nil {
			return err
		}
		r.specNode(network, &node, spec, staticNodes)
		return nil
	})

	enodeURL = node.Status.EnodeURL

	return
}

// specNode updates ethereum node spec
func (r *NetworkReconciler) specNode(network *ethereumv1alpha1.Network, node *ethereumv1alpha1.Node, spec ethereumv1alpha1.NetworkNodeSpec, staticNodes []string) {
	node.Labels = map[string]string{
		"name":     "node",
		"instance": spec.Name,
		"network":  network.Name,
		"protocol": "ethereum",
		"client":   string(spec.Client),
	}
	node.Annotations = map[string]string{
		"kotal.io/static-nodes": strings.Join(staticNodes, ";"),
	}
	node.Spec = spec.NodeSpec
	// override node's network and availability config
	node.Spec.NetworkConfig = network.Spec.NetworkConfig
	node.Spec.AvailabilityConfig = network.Spec.AvailabilityConfig
}

// deleteRedundantNode deletes all nodes that has been removed from spec
func (r *NetworkReconciler) deleteRedundantNodes(network *ethereumv1alpha1.Network) error {
	log := r.Log.WithName("delete redundant nodes")

	var nodeList ethereumv1alpha1.NodeList
	names := map[string]bool{}
	matchingLabels := client.MatchingLabels{
		"name":     "node",
		"network":  network.Name,
		"protocol": "ethereum",
	}
	inNamespace := client.InNamespace(network.Namespace)

	for _, node := range network.Spec.Nodes {
		name := fmt.Sprintf("%s-%s", network.Name, node.Name)
		names[name] = true
	}

	// Nodes
	if err := r.Client.List(context.Background(), &nodeList, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node statefulsets")
		return err
	}

	for _, node := range nodeList.Items {
		name := node.Name
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node %s", name))

			if err := r.Client.Delete(context.Background(), &node); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node %s", name))
				return err
			}
		}
	}

	return nil
}

// SetupWithManager adds reconciler to the manager
func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Network{}).
		Owns(&ethereumv1alpha1.Node{}).
		Complete(r)
}
