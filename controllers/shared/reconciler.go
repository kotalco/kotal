package shared

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) GetClient() client.Client {
	return r.Client
}

func (r *Reconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// ReconcileOwned reconciles k8s object according to custom resource spec
func (r Reconciler) ReconcileOwned(ctx context.Context, cr CustomResource, obj client.Object, updateFn func(client.Object) error) error {

	obj.SetName(cr.GetName())
	obj.SetNamespace(cr.GetNamespace())

	_, err := ctrl.CreateOrUpdate(ctx, r.GetClient(), obj, func() error {
		if err := ctrl.SetControllerReference(cr, obj, r.GetScheme()); err != nil {
			return err
		}

		updateFn(obj)

		return nil
	})

	return err
}
