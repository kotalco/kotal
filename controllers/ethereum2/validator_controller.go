package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// ValidatorReconciler reconciles a Validator object
type ValidatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=validators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=validators/status,verbs=get;update;patch

// Reconcile reconciles Ethereum 2.0 validator client node
func (r *ValidatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var validator ethereum2v1alpha1.Validator

	if err = r.Client.Get(ctx, req.NamespacedName, &validator); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&validator)

	if err = r.reconcileValidatorStatefulset(&validator); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the validator
func (r *ValidatorReconciler) updateLabels(validator *ethereum2v1alpha1.Validator) {

	if validator.Labels == nil {
		validator.Labels = map[string]string{}
	}

	validator.Labels["name"] = "node"
	validator.Labels["protocol"] = "ethereum2"
	validator.Labels["instance"] = validator.Name
}

// specValidatorStatefulset updates node statefulset spec
func (r *ValidatorReconciler) specValidatorStatefulset(validator *ethereum2v1alpha1.Validator, sts *appsv1.StatefulSet) {

	sts.Labels = validator.GetLabels()

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: validator.GetLabels(),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validator.GetLabels(),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "validator",
						Image: "consensys/teku",
						Args:  []string{"vc"},
					},
				},
			},
		},
	}
}

// reconcileValidatorStatefulset reconciles node statefulset
func (r *ValidatorReconciler) reconcileValidatorStatefulset(validator *ethereum2v1alpha1.Validator) error {
	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      validator.Name,
			Namespace: validator.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, &sts, func() error {
		if err := ctrl.SetControllerReference(validator, &sts, r.Scheme); err != nil {
			return err
		}

		r.specValidatorStatefulset(validator, &sts)

		return nil
	})

	return err
}

// SetupWithManager adds reconciler to the manager
func (r *ValidatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereum2v1alpha1.Validator{}).
		Complete(r)
}
