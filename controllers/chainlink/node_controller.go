package controllers

import (
	"context"
	_ "embed"
	"fmt"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	chainlinkClients "github.com/kotalco/kotal/clients/chainlink"
	"github.com/kotalco/kotal/controllers/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {

	var node chainlinkv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "chainlink")

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcilePVC(ctx, &node); err != nil {
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

	// chainlink chmod the root dir
	// we mount data volume at home dir
	// chainlink root dir will be mounted at $data/kotal-data
	dataMountPath := homeDir

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
				InitContainers: []corev1.Container{
					{
						Name:    "copy-api-credentials",
						Image:   "busybox",
						Command: []string{"/bin/sh"},
						Env: []corev1.EnvVar{
							{
								Name:  "KOTAL_DATA_PATH",
								Value: shared.PathData(homeDir),
							},
							{
								Name:  "KOTAL_EMAIL",
								Value: node.Spec.APICredentials.Email,
							},
							{
								Name:  "KOTAL_SECRETS_PATH",
								Value: shared.PathSecrets(homeDir),
							},
						},
						Args: []string{fmt.Sprintf("%s/copy_api_credentials.sh", shared.PathConfig(homeDir))},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: dataMountPath,
							},
							{
								Name:      "config",
								MountPath: shared.PathConfig(homeDir),
							},
							{
								Name:      "secrets",
								MountPath: shared.PathSecrets(homeDir),
							},
						},
					},
				},
				SecurityContext: shared.SecurityContext(),
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   image,
						Command: command,
						Args:    args,
						Env:     env,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPU),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Resources.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Resources.MemoryLimit),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "secrets",
								MountPath: shared.PathSecrets(homeDir),
							},
							{
								Name:      "data",
								MountPath: dataMountPath,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: node.Name,
							},
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

// reconcilePVC reconciles chainlink node persistent volume claim
func (r *NodeReconciler) reconcilePVC(ctx context.Context, node *chainlinkv1alpha1.Node) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(node, pvc, r.Scheme); err != nil {
			return err
		}

		r.specPVC(node, pvc)

		return nil
	})

	return err
}

// specPVC updates chainlink persistent volume claim
func (r *NodeReconciler) specPVC(node *chainlinkv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = node.Labels
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: node.Spec.StorageClass,
	}
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chainlinkv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
