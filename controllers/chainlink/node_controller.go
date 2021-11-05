package controllers

import (
	"context"
	_ "embed"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	chainlinkClients "github.com/kotalco/kotal/clients/chainlink"
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
	Scheme *runtime.Scheme
}

var (
	//go:embed copy_api_credentials.sh
	CopyAPICredentials string
)

// +kubebuilder:rbac:groups=chainlink.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chainlink.kotal.io,resources=nodes/status,verbs=get;update;patch

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {

	var node chainlinkv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// reconcileConfigmap reconciles chainlink node configmap
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *chainlinkv1alpha1.Node) error {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, config, func() error {
		if err := ctrl.SetControllerReference(node, config, r.Scheme); err != nil {
			return err
		}

		r.specConfigmap(node, config)

		return nil
	})

	return err

}

// specConfigmap updates chainlink node configmap spec
func (r *NodeReconciler) specConfigmap(node *chainlinkv1alpha1.Node, config *corev1.ConfigMap) {
	config.ObjectMeta.Labels = node.Labels

	if config.Data == nil {
		config.Data = make(map[string]string)
	}

	config.Data["copy_api_credentials.sh"] = CopyAPICredentials
}

// reconcileStatefulset reconciles node statefulset
func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *chainlinkv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := chainlinkClients.NewClient(node)

	img := client.Image()
	command := client.Command()
	args := client.Args()
	env := client.Env()
	homeDir := client.HomeDir()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specStatefulSet(node, sts, img, homeDir, command, args, env); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *chainlinkv1alpha1.Node, sts *appsv1.StatefulSet, image, homeDir string, command, args []string, env []corev1.EnvVar) error {

	// TODO: use shared node labels
	labels := map[string]string{
		"name": node.Name,
	}

	sts.ObjectMeta.Labels = labels

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		ServiceName: node.Name,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					{
						Name:    "copy-api-credentials",
						Image:   "busybox",
						Command: []string{"/bin/sh"},
						Env: []corev1.EnvVar{
							{
								Name:  "KOTAL_DATA_PATH",
								Value: "/.chainlink",
							},
							{
								Name:  "KOTAL_EMAIL",
								Value: node.Spec.APICredentials.Email,
							},
							{
								Name:  "KOTAL_SECRETS_PATH",
								Value: "/secrets",
							},
						},
						Args: []string{"/config/copy_api_credentials.sh"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/.chainlink",
							},
							{
								Name:      "config",
								MountPath: "/config",
							},
							{
								Name:      "secrets",
								MountPath: "/secrets",
							},
						},
					},
				},
				// TODO: use shared security context
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   image,
						Command: command,
						Args:    args,
						Env:     env,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "secrets",
								MountPath: "/secrets",
							},
							{
								Name:      "data",
								MountPath: "/.chainlink",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
					{
						Name: "secrets",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										Secret: &corev1.SecretProjection{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: node.Spec.KeystorePasswordSecretName,
											},
											Items: []corev1.KeyToPath{
												{
													Key:  "password",
													Path: "keystore-password",
												},
											},
										},
									},
									{
										Secret: &corev1.SecretProjection{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: node.Spec.APICredentials.PasswordSecretName,
											},
											Items: []corev1.KeyToPath{
												{
													Key:  "password",
													Path: "api-password",
												},
											},
										},
									},
								},
							},
						},
					},
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
		For(&chainlinkv1alpha1.Node{}).
		Complete(r)
}
