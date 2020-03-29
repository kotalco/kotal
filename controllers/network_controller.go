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

	"github.com/go-logr/logr"
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

// Reconcile reconciles ethereum networks
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum.kotal.io,resources=networks/status,verbs=get;update;patch
func (r *NetworkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("network", req.NamespacedName)

	var network ethereumv1alpha1.Network
	if err := r.Client.Get(ctx, req.NamespacedName, &network); err != nil {
		log.Error(err, "Unable to fetch Ethereum Network")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//TODO: reconcile the network

	return ctrl.Result{}, nil
}

// SetupWithManager adds reconciler to the manager
func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereumv1alpha1.Network{}).
		Complete(r)
}
