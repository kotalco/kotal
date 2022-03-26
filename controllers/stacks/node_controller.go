package controllers

import (
	"context"

	stacksClients "github.com/kotalco/kotal/clients/stacks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=stacks.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stacks.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node stacksv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// TODO: default the node if webhooks are disabled

	shared.UpdateLabels(&node, "stacks-node")

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// specConfigmap updates node statefulset spec
func (r *NodeReconciler) specConfigmap(node *stacksv1alpha1.Node, configmap *corev1.ConfigMap, configToml string) {
	configmap.ObjectMeta.Labels = node.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	configmap.Data["config.toml"] = configToml

}

// reconcileConfigmap reconciles node configmap
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *stacksv1alpha1.Node) error {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	// filecoin generates config.toml file from node spec
	configToml, err := ConfigFromSpec(node)
	if err != nil {
		return err
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(node, configmap, r.Scheme); err != nil {
			return err
		}
		r.specConfigmap(node, configmap, configToml)
		return nil
	})

	return err
}

// reconcileStatefulset reconciles node statefulset
func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *stacksv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := stacksClients.NewClient(node)

	img := client.Image()
	homeDir := client.HomeDir()
	cmd := client.Command()
	args := client.Args()
	env := client.Env()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specStatefulSet(node, sts, img, homeDir, env, cmd, args); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *stacksv1alpha1.Node, sts *appsv1.StatefulSet, img, homeDir string, env []corev1.EnvVar, cmd, args []string) error {

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
				SecurityContext: shared.SecurityContext(),
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   img,
						Command: cmd,
						Args:    args,
						Env:     env,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								ReadOnly:  true,
								MountPath: shared.PathConfig(homeDir),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: node.Name,
								},
							},
						},
					},
				},
			},
		},
	}

	return nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stacksv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
