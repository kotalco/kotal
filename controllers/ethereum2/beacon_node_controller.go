package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	ethereum2Clients "github.com/kotalco/kotal/clients/ethereum2"
	"github.com/kotalco/kotal/controllers/shared"
)

// BeaconNodeReconciler reconciles a Node object
type BeaconNodeReconciler struct {
	shared.Reconciler
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=beaconnodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=beaconnodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles Ethereum 2.0 beacon node
func (r *BeaconNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node ethereum2v1alpha1.BeaconNode

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the beacon node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, string(node.Spec.Client), node.Spec.Network)

	// reconcile persistent volume clain
	if err = r.ReconcileOwned(ctx, &node, &corev1.PersistentVolumeClaim{}, func(obj client.Object) error {
		r.specPVC(&node, obj.(*corev1.PersistentVolumeClaim))
		return nil
	}); err != nil {
		return
	}

	// reconcile service
	if err = r.ReconcileOwned(ctx, &node, &corev1.Service{}, func(obj client.Object) error {
		r.specService(&node, obj.(*corev1.Service))
		return nil
	}); err != nil {
		return
	}

	// reconcile service
	if err = r.ReconcileOwned(ctx, &node, &appsv1.StatefulSet{}, func(obj client.Object) error {
		client, err := ethereum2Clients.NewClient(&node)
		if err != nil {
			return err
		}

		args := client.Args()
		command := client.Command()
		homeDir := client.HomeDir()

		r.specStatefulset(&node, obj.(*appsv1.StatefulSet), args, command, homeDir)
		return nil
	}); err != nil {
		return
	}

	return
}

func (r *BeaconNodeReconciler) specService(node *ethereum2v1alpha1.BeaconNode, svc *corev1.Service) {
	labels := node.GetLabels()

	svc.ObjectMeta.Labels = labels
	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "discovery",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromString("discovery"),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromString("p2p"),
		},
	}

	if node.Spec.RPC {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromString("rpc"),
		})
	}

	if node.Spec.GRPC {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "grpc",
			Port:       int32(node.Spec.GRPCPort),
			TargetPort: intstr.FromString("grpc"),
		})
	}

	if node.Spec.REST {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rest",
			Port:       int32(node.Spec.RESTPort),
			TargetPort: intstr.FromString("rest"),
		})
	}

	svc.Spec.Selector = labels
}

// specPVC updates beacon node persistent volume claim spec
func (r *BeaconNodeReconciler) specPVC(node *ethereum2v1alpha1.BeaconNode, pvc *corev1.PersistentVolumeClaim) {

	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.Labels = node.GetLabels()

	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: node.Spec.Resources.StorageClass,
	}
}

// nodeVolumes returns node volumes
func (r *BeaconNodeReconciler) nodeVolumes(node *ethereum2v1alpha1.BeaconNode) (volumes []corev1.Volume) {
	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: node.Name,
			},
		},
	}
	volumes = append(volumes, dataVolume)

	// projected volume sources
	volumeProjections := []corev1.VolumeProjection{
		{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.JWTSecretName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "secret",
						Path: "jwt.secret",
					},
				},
			},
		},
	}

	if node.Spec.CertSecretName != "" {
		volumeProjections = append(volumeProjections, corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: node.Spec.CertSecretName,
				},
			},
		})
	}

	volumes = append(volumes, corev1.Volume{
		Name: "secrets",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: volumeProjections,
			},
		},
	})

	return
}

// nodeVolumeMounts returns node volume mounts
func (r *BeaconNodeReconciler) nodeVolumeMounts(node *ethereum2v1alpha1.BeaconNode, homeDir string) (mounts []corev1.VolumeMount) {
	dataDir := shared.PathData(homeDir)

	// Nimbus required changing permission of the data dir to be
	// read and write by owner only
	// that's why we mount volume at $HOME
	// but data dir is atatched at $HOME/kota-data
	if node.Spec.Client == ethereum2v1alpha1.NimbusClient {
		dataDir = homeDir
	}

	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: dataDir,
	}
	mounts = append(mounts, dataMount)

	secretMount := corev1.VolumeMount{
		Name:      "secrets",
		MountPath: shared.PathSecrets(homeDir),
	}
	mounts = append(mounts, secretMount)

	return
}

// specStatefulset updates beacon node statefulset spec
func (r *BeaconNodeReconciler) specStatefulset(node *ethereum2v1alpha1.BeaconNode, sts *appsv1.StatefulSet, args, command []string, homeDir string) {

	sts.Labels = node.GetLabels()

	volumes := r.nodeVolumes(node)
	volumeMounts := r.nodeVolumeMounts(node, homeDir)

	initContainers := []corev1.Container{}

	if node.Spec.Client == ethereum2v1alpha1.NimbusClient {
		// Nimbus client requires data dir path to be read and write only by the owner 0700
		fixPermissionContainer := corev1.Container{
			Name:  "fix-datadir-permission",
			Image: node.Spec.Image,
			Command: []string{
				"/bin/sh",
				"-c",
			},
			Args: []string{
				fmt.Sprintf(`
					mkdir -p %s &&
					chmod 700 %s`,
					shared.PathData(homeDir),
					shared.PathData(homeDir),
				),
			},
			VolumeMounts: volumeMounts,
		}
		initContainers = append(initContainers, fixPermissionContainer)

		if node.Spec.CheckpointSyncURL != "" {
			checkpointSyncContainer := corev1.Container{
				Name:    "checkpoint-sync",
				Image:   node.Spec.Image,
				Command: []string{"nimbus_beacon_node", "trustedNodeSync"},
				Args: []string{
					fmt.Sprintf("--network=%s", node.Spec.Network),
					fmt.Sprintf("--data-dir=%s", shared.PathData(homeDir)),
					fmt.Sprintf("--trusted-node-url=%s", node.Spec.CheckpointSyncURL),
				},
				VolumeMounts: volumeMounts,
			}
			initContainers = append(initContainers, checkpointSyncContainer)
		}
	}

	ports := []corev1.ContainerPort{
		{
			Name:          "discovery",
			ContainerPort: int32(node.Spec.P2PPort),
			Protocol:      corev1.ProtocolUDP,
		},
		{
			Name:          "p2p",
			ContainerPort: int32(node.Spec.P2PPort),
		},
	}

	if node.Spec.RPC {
		ports = append(ports, corev1.ContainerPort{
			Name:          "rpc",
			ContainerPort: int32(node.Spec.RPCPort),
		})
	}

	if node.Spec.GRPC {
		ports = append(ports, corev1.ContainerPort{
			Name:          "grpc",
			ContainerPort: int32(node.Spec.GRPCPort),
		})
	}

	if node.Spec.REST {
		ports = append(ports, corev1.ContainerPort{
			Name:          "rest",
			ContainerPort: int32(node.Spec.RESTPort),
		})
	}

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.GetLabels(),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: node.GetLabels(),
			},
			Spec: corev1.PodSpec{
				SecurityContext: shared.SecurityContext(),
				InitContainers:  initContainers,
				Containers: []corev1.Container{
					{
						Name:         "node",
						Command:      command,
						Args:         args,
						Ports:        ports,
						Image:        node.Spec.Image,
						VolumeMounts: volumeMounts,
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
					},
				},
				Volumes: volumes,
			},
		},
	}
}

// SetupWithManager adds reconciler to the manager
func (r *BeaconNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereum2v1alpha1.BeaconNode{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
