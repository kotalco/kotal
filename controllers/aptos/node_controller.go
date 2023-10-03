package controllers

import (
	"context"
	_ "embed"
	"fmt"

	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	aptosClients "github.com/kotalco/kotal/clients/aptos"
	"github.com/kotalco/kotal/controllers/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	shared.Reconciler
}

var (
	//go:embed download_waypoint.sh
	downloadWaypoint string
	//go:embed download_genesis_block.sh
	downloadGenesisBlock string
)

// +kubebuilder:rbac:groups=aptos.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aptos.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node aptosv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "aptos-core", string(node.Spec.Network))

	// reconcile config map
	r.ReconcileOwned(ctx, &node, &corev1.ConfigMap{}, func(obj client.Object) error {
		r.specConfigmap(&node, obj.(*corev1.ConfigMap))
		return nil
	})

	// reconcile service
	r.ReconcileOwned(ctx, &node, &corev1.Service{}, func(obj client.Object) error {
		r.specService(&node, obj.(*corev1.Service))
		return nil
	})

	// reconcile persistent volume claim
	r.ReconcileOwned(ctx, &node, &corev1.PersistentVolumeClaim{}, func(obj client.Object) error {
		r.specPVC(&node, obj.(*corev1.PersistentVolumeClaim))
		return nil
	})

	// reconcile statefulset
	r.ReconcileOwned(ctx, &node, &appsv1.StatefulSet{}, func(obj client.Object) error {
		client := aptosClients.NewClient(&node)

		homeDir := client.HomeDir()
		cmd := client.Command()
		args := client.Args()
		env := client.Env()

		return r.specStatefulSet(&node, obj.(*appsv1.StatefulSet), homeDir, env, cmd, args)
	})

	return
}

// specConfigmap updates node configmap
func (n *NodeReconciler) specConfigmap(node *aptosv1alpha1.Node, configmap *corev1.ConfigMap) {
	configmap.ObjectMeta.Labels = node.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	config, err := ConfigFromSpec(node, n.Client)
	if err != nil {
		return
	}

	configmap.Data["config.yaml"] = config
	configmap.Data["download_waypoint.sh"] = downloadWaypoint
	configmap.Data["download_genesis_block.sh"] = downloadGenesisBlock

}

// specPVC updates Aptos node persistent volume claim
func (r *NodeReconciler) specPVC(node *aptosv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
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
	}
}

// specService updates Aptos node service spec
func (r *NodeReconciler) specService(node *aptosv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromString("p2p"),
		},
		{
			Name:       "metrics",
			Port:       int32(node.Spec.MetricsPort),
			TargetPort: intstr.FromString("metrics"),
		},
	}

	if node.Spec.API {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "api",
			Port:       int32(node.Spec.APIPort),
			TargetPort: intstr.FromString("api"),
		})
	}

	svc.Spec.Selector = labels
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *aptosv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, env []corev1.EnvVar, cmd, args []string) error {

	sts.ObjectMeta.Labels = node.Labels

	initContainers := []corev1.Container{}

	if node.Spec.Waypoint == "" {
		initContainers = append(initContainers, corev1.Container{
			Name:  "download-waypoint",
			Image: "curlimages/curl:8.00.1",
			Env: []corev1.EnvVar{
				{
					Name:  "KOTAL_NETWORK",
					Value: string(node.Spec.Network),
				},
				{
					Name:  shared.EnvDataPath,
					Value: shared.PathData(homeDir),
				},
			},
			Command: []string{"/bin/sh"},
			Args:    []string{fmt.Sprintf("%s/download_waypoint.sh", shared.PathConfig(homeDir))},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "config",
					MountPath: shared.PathConfig(homeDir),
					ReadOnly:  true,
				},
				{
					Name:      "data",
					MountPath: shared.PathData(homeDir),
				},
			},
		})
	}

	sources := []corev1.VolumeProjection{
		{
			// config.yaml
			ConfigMap: &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Name,
				},
			},
		},
	}

	if node.Spec.GenesisConfigmapName == "" {
		initContainers = append(initContainers, corev1.Container{
			Name:  "download-genesis-block",
			Image: "curlimages/curl:8.00.1",
			Env: []corev1.EnvVar{
				{
					Name:  "KOTAL_NETWORK",
					Value: string(node.Spec.Network),
				},
				{
					Name:  shared.EnvDataPath,
					Value: shared.PathData(homeDir),
				},
			},
			Command: []string{"/bin/sh"},
			Args:    []string{fmt.Sprintf("%s/download_genesis_block.sh", shared.PathConfig(homeDir))},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "config",
					MountPath: shared.PathConfig(homeDir),
					ReadOnly:  true,
				},
				{
					Name:      "data",
					MountPath: shared.PathData(homeDir),
				},
			},
		})
	} else {
		sources = append(sources, corev1.VolumeProjection{
			// genesis.blob
			ConfigMap: &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.GenesisConfigmapName,
				},
			},
		})
	}

	ports := []corev1.ContainerPort{
		{
			Name:          "p2p",
			ContainerPort: int32(node.Spec.P2PPort),
		},
		{
			Name:          "metrics",
			ContainerPort: int32(node.Spec.MetricsPort),
		},
	}

	if node.Spec.API {
		ports = append(ports, corev1.ContainerPort{
			Name:          "api",
			ContainerPort: int32(node.Spec.APIPort),
		})
	}

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
				InitContainers:  initContainers,
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   node.Spec.Image,
						Command: cmd,
						Args:    args,
						Env:     env,
						Ports:   ports,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.CPU),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.CPULimit),
								corev1.ResourceMemory: resource.MustParse(node.Spec.MemoryLimit),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								MountPath: shared.PathConfig(homeDir),
								ReadOnly:  true,
							},
							{
								Name:      "data",
								MountPath: shared.PathData(homeDir),
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
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: sources,
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
		For(&aptosv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
