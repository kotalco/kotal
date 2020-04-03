/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereumv1alpha1 "github.com/mfarghaly/kotal/api/v1alpha1"
)

// NetworkReconciler reconciles a Network object
type NetworkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;create;update

// Reconcile reconciles ethereum networks
func (r *NetworkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("network", req.NamespacedName)

	var network ethereumv1alpha1.Network

	if err := r.Client.Get(ctx, req.NamespacedName, &network); err != nil {
		log.Error(err, "Unable to fetch Ethereum Network")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: delete nodes that has been removed from nodes list

	for _, node := range network.Spec.Nodes {
		if err := r.reconcileNode(ctx, &node, &network); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// reconcileNode create a new node deployment if it doesn't exist
// updates existing deployments if node spec changed
func (r *NetworkReconciler) reconcileNode(ctx context.Context, node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network) error {
	log := r.Log.WithValues("node", node.Name)
	dep := &appsv1.Deployment{}
	ns := network.ObjectMeta.Namespace

	err := r.Client.Get(ctx, client.ObjectKey{
		Name:      node.Name,
		Namespace: ns,
	}, dep)

	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("unable to find node (%s) deployment", node.Name))
			return err
		}

		log.Info(fmt.Sprintf("node (%s) deployment is not found", node.Name))
		log.Info(fmt.Sprintf("creating a new deployment for node (%s)", node.Name))

		// TODO: move into new function
		newDep := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      node.Name,
				Namespace: ns,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "node",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "node",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name:  "node",
								Image: "hyperledger/besu",
								Command: []string{
									"besu",
								},
								Args: []string{
									"--network",
									network.Spec.Join,
								},
							},
						},
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(network, &newDep, r.Scheme); err != nil {
			log.Error(err, "Unable to set controller reference")
			return err
		}

		if err := r.Client.Create(ctx, &newDep); err != nil {
			log.Error(err, "Unable to create node deployment")
			return err
		}

	}

	// TODO: update existing deployment if node spec change

	return nil
}

// SetupWithManager adds reconciler to the manager
func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Network{}).
		Complete(r)
}
