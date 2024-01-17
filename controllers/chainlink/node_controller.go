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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	shared.Reconciler
}

const (
	envApiEmail = "KOTAL_API_EMAIL"
)

var (
	//go:embed copy_api_credentials.sh
	CopyAPICredentials string
)

// +kubebuilder:rbac:groups=chainlink.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chainlink.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node chainlinkv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "chainlink", "")

	// reconcile service
	if err = r.ReconcileOwned(ctx, &node, &corev1.Service{}, func(obj client.Object) error {
		r.specService(&node, obj.(*corev1.Service))
		return nil
	}); err != nil {
		return
	}

	// reconcile config map
	if err = r.ReconcileOwned(ctx, &node, &corev1.ConfigMap{}, func(obj client.Object) error {
		homeDir := chainlinkClients.NewClient(&node).HomeDir()

		configToml, err := ConfigFromSpec(&node, homeDir)
		if err != nil {
			return err
		}

		secretsConfigToml, err := SecretsFromSpec(&node, homeDir, r.Client)
		if err != nil {
			return err
		}
		r.specConfigmap(&node, obj.(*corev1.ConfigMap), configToml, secretsConfigToml)
		return nil
	}); err != nil {
		return
	}

	// reconcile persistent volume claim
	if err = r.ReconcileOwned(ctx, &node, &corev1.PersistentVolumeClaim{}, func(obj client.Object) error {
		r.specPVC(&node, obj.(*corev1.PersistentVolumeClaim))
		return nil
	}); err != nil {
		return
	}

	// reconcile stateful set
	if err = r.ReconcileOwned(ctx, &node, &appsv1.StatefulSet{}, func(obj client.Object) error {
		client := chainlinkClients.NewClient(&node)

		command := client.Command()
		args := client.Args()
		args = append(args, node.Spec.ExtraArgs.Encode(false)...)
		env := client.Env()
		homeDir := client.HomeDir()
		return r.specStatefulSet(&node, obj.(*appsv1.StatefulSet), homeDir, command, args, env)
	}); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &node); err != nil {
		return
	}

	return
}

// updateStatus updates chainlink node status
func (r *NodeReconciler) updateStatus(ctx context.Context, node *chainlinkv1alpha1.Node) error {
	node.Status.Client = "chainlink"

	if err := r.Status().Update(ctx, node); err != nil {
		log.FromContext(ctx).Error(err, "unable to update node status")
		return err
	}

	return nil
}

// specService updates node service spec
func (r *NodeReconciler) specService(node *chainlinkv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromString("p2p"),
		},
	}

	if node.Spec.API {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "api",
			Port:       int32(node.Spec.APIPort),
			TargetPort: intstr.FromString("api"),
		})
	}

	if node.Spec.TLSPort != 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "tls",
			Port:       int32(node.Spec.TLSPort),
			TargetPort: intstr.FromString("tls"),
		})
	}

	svc.Spec.Selector = labels
}

// specConfigmap updates chainlink node configmap spec
func (r *NodeReconciler) specConfigmap(node *chainlinkv1alpha1.Node, config *corev1.ConfigMap, configToml, secretsConfigToml string) {
	config.ObjectMeta.Labels = node.Labels

	if config.Data == nil {
		config.Data = make(map[string]string)
	}

	config.Data["config.toml"] = configToml
	config.Data["secrets.toml"] = secretsConfigToml
	config.Data["copy_api_credentials.sh"] = CopyAPICredentials
}

func (r *NodeReconciler) createVolumes(node *chainlinkv1alpha1.Node) []corev1.Volume {
	volumes := []corev1.Volume{}

	// data volume
	volumes = append(volumes, corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: node.Name,
			},
		},
	})

	// config volume
	volumes = append(volumes, corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Name,
				},
			},
		},
	})

	// projected volume sources
	sources := []corev1.VolumeProjection{
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
	}

	if node.Spec.CertSecretName != "" {
		sources = append(sources, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.CertSecretName,
				},
			},
		})
	}

	// secrets volume
	volumes = append(volumes, corev1.Volume{
		Name: "secrets",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: sources,
			},
		},
	})

	return volumes
}

func (r *NodeReconciler) createVolumeMounts(node *chainlinkv1alpha1.Node, homeDir string) []corev1.VolumeMount {
	// chainlink chmod the root dir
	// we mount data volume at home dir
	// chainlink root dir will be mounted at $data/kotal-data
	return []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: homeDir,
		},
		{
			Name:      "config",
			MountPath: shared.PathConfig(homeDir),
		},
		{
			Name:      "secrets",
			MountPath: shared.PathSecrets(homeDir),
		},
	}
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *chainlinkv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, command, args []string, env []corev1.EnvVar) error {

	sts.ObjectMeta.Labels = node.Labels

	ports := []corev1.ContainerPort{
		{
			Name:          "p2p",
			ContainerPort: int32(node.Spec.P2PPort),
		},
	}

	if node.Spec.API {
		ports = append(ports, corev1.ContainerPort{
			Name:          "api",
			ContainerPort: int32(node.Spec.APIPort),
		})
	}

	if node.Spec.TLSPort != 0 {
		ports = append(ports, corev1.ContainerPort{
			Name:          "tls",
			ContainerPort: int32(node.Spec.TLSPort),
		})
	}

	replicas := int32(*node.Spec.Replicas)

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.Labels,
		},
		Replicas:    &replicas,
		ServiceName: node.Name,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: node.Labels,
			},
			Spec: corev1.PodSpec{
				SecurityContext: shared.SecurityContext(),
				InitContainers: []corev1.Container{
					{
						Name:    "copy-api-credentials",
						Image:   shared.BusyboxImage,
						Command: []string{"/bin/sh"},
						Env: []corev1.EnvVar{
							{
								Name:  shared.EnvDataPath,
								Value: shared.PathData(homeDir),
							},
							{
								Name:  envApiEmail,
								Value: node.Spec.APICredentials.Email,
							},
							{
								Name:  shared.EnvSecretsPath,
								Value: shared.PathSecrets(homeDir),
							},
						},
						Args:         []string{fmt.Sprintf("%s/copy_api_credentials.sh", shared.PathConfig(homeDir))},
						VolumeMounts: r.createVolumeMounts(node, homeDir),
					},
				},
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   node.Spec.Image,
						Command: command,
						Args:    args,
						Env:     env,
						Ports:   ports,
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
						VolumeMounts: r.createVolumeMounts(node, homeDir),
					},
				},
				Volumes: r.createVolumes(node),
			},
		},
	}

	return nil
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
		Resources: corev1.VolumeResourceRequirements{
			Requests: request,
		},
		StorageClassName: node.Spec.StorageClass,
	}
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&chainlinkv1alpha1.Node{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
