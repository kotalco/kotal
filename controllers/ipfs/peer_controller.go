package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	ipfsClients "github.com/kotalco/kotal/clients/ipfs"
	"github.com/kotalco/kotal/controllers/shared"
)

// PeerReconciler reconciles a Peer object
type PeerReconciler struct {
	shared.Reconciler
}

var (
	//go:embed init_ipfs_config.sh
	initIPFSConfigScript string
	//go:embed copy_swarm_key.sh
	copySwarmKeyScript string
	//go:embed config_ipfs.sh
	configIPFSScript string
)

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=peers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=peers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *PeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var peer ipfsv1alpha1.Peer

	if err = r.Client.Get(ctx, req.NamespacedName, &peer); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the peer if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		peer.Default()
	}

	shared.UpdateLabels(&peer, "kubo", "")

	// reconcile config map
	if err = r.ReconcileOwned(ctx, &peer, &corev1.ConfigMap{}, func(obj client.Object) error {
		r.specConfigmap(&peer, obj.(*corev1.ConfigMap))
		return nil
	}); err != nil {
		return
	}

	// reconcile service
	if err = r.ReconcileOwned(ctx, &peer, &corev1.Service{}, func(obj client.Object) error {
		r.specService(&peer, obj.(*corev1.Service))
		return nil
	}); err != nil {
		return
	}

	// reconcile persistent volume claim
	if err = r.ReconcileOwned(ctx, &peer, &corev1.PersistentVolumeClaim{}, func(obj client.Object) error {
		r.specPVC(&peer, obj.(*corev1.PersistentVolumeClaim))
		return nil
	}); err != nil {
		return
	}

	// reconcile stateful set
	if err = r.ReconcileOwned(ctx, &peer, &appsv1.StatefulSet{}, func(obj client.Object) error {
		client, err := ipfsClients.NewClient(&peer)
		if err != nil {
			return err
		}

		command := client.Command()
		env := client.Env()
		args := client.Args()
		args = append(args, peer.Spec.ExtraArgs.Encode(false)...)
		homeDir := client.HomeDir()

		r.specStatefulSet(&peer, obj.(*appsv1.StatefulSet), homeDir, env, command, args)
		return nil
	}); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &peer); err != nil {
		return
	}

	return
}

// updateStatus updates ipfs peer status
func (r *PeerReconciler) updateStatus(ctx context.Context, peer *ipfsv1alpha1.Peer) error {
	// TODO: update after multi-client support
	peer.Status.Client = "kubo"

	if err := r.Status().Update(ctx, peer); err != nil {
		log.FromContext(ctx).Error(err, "unable to update peer status")
		return err
	}

	return nil
}

// specService updates ipfs peer service spec
func (r *PeerReconciler) specService(peer *ipfsv1alpha1.Peer, svc *corev1.Service) {
	labels := peer.Labels

	svc.ObjectMeta.Labels = labels

	ports := []corev1.ServicePort{
		{
			Name:       "swarm",
			Port:       4001,
			TargetPort: intstr.FromString("swarm"),
		},
		{
			Name:       "swarm-udp",
			Port:       4001,
			TargetPort: intstr.FromString("swarm-udp"),
			Protocol:   corev1.ProtocolUDP,
		},
	}

	if peer.Spec.API {
		ports = append(ports, corev1.ServicePort{
			Name:       "api",
			Port:       int32(peer.Spec.APIPort),
			TargetPort: intstr.FromString("api"),
		})
	}

	if peer.Spec.Gateway {
		ports = append(ports, corev1.ServicePort{
			Name:       "gateway",
			Port:       int32(peer.Spec.GatewayPort),
			TargetPort: intstr.FromString("gateway"),
		})
	}

	svc.Spec.Ports = ports

	svc.Spec.Selector = labels
}

// specConfigmap updates ipfs peer config spec
func (r *PeerReconciler) specConfigmap(peer *ipfsv1alpha1.Peer, config *corev1.ConfigMap) {
	config.ObjectMeta.Labels = peer.Labels
	if config.Data == nil {
		config.Data = make(map[string]string)
	}
	config.Data["init_ipfs_config.sh"] = initIPFSConfigScript
	config.Data["copy_swarm_key.sh"] = copySwarmKeyScript
	config.Data["config_ipfs.sh"] = configIPFSScript
}

// specPVC updates ipfs peer persistent volume claim
func (r *PeerReconciler) specPVC(peer *ipfsv1alpha1.Peer, pvc *corev1.PersistentVolumeClaim) {
	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(peer.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = peer.Labels
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: peer.Spec.Resources.StorageClass,
	}
}

// specStatefulSet updates ipfs peer statefulset spec
func (r *PeerReconciler) specStatefulSet(peer *ipfsv1alpha1.Peer, sts *appsv1.StatefulSet, homeDir string, env []corev1.EnvVar, command, args []string) {
	labels := peer.Labels

	sts.ObjectMeta.Labels = labels

	volumes := []corev1.Volume{
		{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: peer.Name,
				},
			},
		},
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: peer.Name,
					},
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: shared.PathData(homeDir),
		},
		{
			Name:      "config",
			MountPath: shared.PathConfig(homeDir),
		},
	}

	initContainers := []corev1.Container{}

	// copy swarm key before init ipfs
	if peer.Spec.SwarmKeySecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "swarm-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: peer.Spec.SwarmKeySecretName,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "swarm-key",
			MountPath: shared.PathSecrets(homeDir),
		})

		initContainers = append(initContainers, corev1.Container{
			Name:  "copy-swarm-key",
			Image: shared.BusyboxImage,
			Env: []corev1.EnvVar{
				{
					Name:  ipfsClients.EnvIPFSPath,
					Value: shared.PathData(homeDir),
				},
				{
					Name:  shared.EnvSecretsPath,
					Value: shared.PathSecrets(homeDir),
				},
			},
			Command: []string{"/bin/sh"},
			Args: []string{
				fmt.Sprintf("%s/copy_swarm_key.sh", shared.PathConfig(homeDir)),
			},
			VolumeMounts: volumeMounts,
		})

	}

	// init ipfs config
	initProfiles := []string{}
	for _, profile := range peer.Spec.InitProfiles {
		initProfiles = append(initProfiles, string(profile))
	}
	initContainers = append(initContainers, corev1.Container{
		Name:  "init-ipfs",
		Image: peer.Spec.Image,
		Env: []corev1.EnvVar{
			{
				Name:  ipfsClients.EnvIPFSPath,
				Value: shared.PathData(homeDir),
			},
			{
				Name:  ipfsClients.EnvIPFSInitProfiles,
				Value: strings.Join(initProfiles, ","),
			},
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			fmt.Sprintf("%s/init_ipfs_config.sh", shared.PathConfig(homeDir)),
		},
		VolumeMounts: volumeMounts,
	})

	// init ipfs config
	profiles := []string{}
	for _, profile := range peer.Spec.Profiles {
		profiles = append(profiles, string(profile))
	}
	// config ipfs
	initContainers = append(initContainers, corev1.Container{
		Name:  "config-ipfs",
		Image: peer.Spec.Image,
		Env: []corev1.EnvVar{
			{
				Name:  ipfsClients.EnvIPFSPath,
				Value: shared.PathData(homeDir),
			},
			{
				Name:  ipfsClients.EnvIPFSAPIPort,
				Value: fmt.Sprintf("%d", peer.Spec.APIPort),
			},
			{
				Name:  ipfsClients.EnvIPFSAPIHost,
				Value: shared.Host(peer.Spec.API),
			},
			{
				Name:  ipfsClients.EnvIPFSGatewayPort,
				Value: fmt.Sprintf("%d", peer.Spec.GatewayPort),
			},
			{
				Name:  ipfsClients.EnvIPFSGatewayHost,
				Value: shared.Host(peer.Spec.Gateway),
			},
			{
				Name:  ipfsClients.EnvIPFSProfiles,
				Value: strings.Join(profiles, ";"),
			},
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			fmt.Sprintf("%s/config_ipfs.sh", shared.PathConfig(homeDir)),
		},
		VolumeMounts: volumeMounts,
	})

	ports := []corev1.ContainerPort{
		{
			Name:          "swarm",
			ContainerPort: 4001,
		},
		{
			Name:          "swarm-udp",
			ContainerPort: 4001,
			Protocol:      corev1.ProtocolUDP,
		},
	}

	if peer.Spec.API {
		ports = append(ports, corev1.ContainerPort{
			Name:          "api",
			ContainerPort: int32(peer.Spec.APIPort),
		})
	}

	if peer.Spec.Gateway {
		ports = append(ports, corev1.ContainerPort{
			Name:          "gateway",
			ContainerPort: int32(peer.Spec.GatewayPort),
		})
	}

	replicas := int32(*peer.Spec.Replicas)

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Replicas: &replicas,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				SecurityContext: shared.SecurityContext(),
				InitContainers:  initContainers,
				Containers: []corev1.Container{
					{
						Name:         "peer",
						Image:        peer.Spec.Image,
						Env:          env,
						Command:      command,
						Args:         args,
						Ports:        ports,
						VolumeMounts: volumeMounts,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(peer.Spec.Resources.CPU),
								corev1.ResourceMemory: resource.MustParse(peer.Spec.Resources.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(peer.Spec.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(peer.Spec.Resources.MemoryLimit),
							},
						},
					},
				},
				Volumes: volumes,
			},
		},
	}
}

// SetupWithManager registers the controller to be started with the given manager
func (r *PeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.Peer{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
