package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
)

// ClusterPeerReconciler reconciles a ClusterPeer object
type ClusterPeerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=clusterpeers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=clusterpeers/status,verbs=get;update;patch

func (r *ClusterPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *ClusterPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.ClusterPeer{}).
		Complete(r)
}
