package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// PeerReconciler reconciles a Peer object
type PeerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	//go:embed init_ipfs_config.sh
	initIPFSConfig string
	//go:embed copy_swarm_key.sh
	copySwarmKey string
	//go:embed config_ipfs.sh
	configIPFS string
)

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=peers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=peers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *PeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var peer ipfsv1alpha1.Peer

	if err = r.Client.Get(ctx, req.NamespacedName, &peer); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the peer if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		peer.Default()
	}

	r.updateLabels(&peer)

	if err = r.reconcilePeerConfig(ctx, &peer); err != nil {
		return
	}

	if err = r.reconcilePeerService(ctx, &peer); err != nil {
		return
	}

	if err = r.reconcilePeerPVC(ctx, &peer); err != nil {
		return
	}

	if err = r.reconcilePeerStatefulSet(ctx, &peer); err != nil {
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
	peer.Status.Client = "go-ipfs"

	if err := r.Status().Update(ctx, peer); err != nil {
		r.Log.Error(err, "unable to update peer status")
		return err
	}

	return nil
}

// updateLabels adds meta labels to the peer
func (r *PeerReconciler) updateLabels(peer *ipfsv1alpha1.Peer) {

	if peer.Labels == nil {
		peer.Labels = map[string]string{}
	}

	// TODO: change client implementing multi-clients
	peer.Labels["name"] = "peer"
	peer.Labels["protocol"] = "ipfs"
	peer.Labels["client"] = "go-ipfs"
	peer.Labels["instance"] = peer.Name

}

// reconcilePeerService reconciles ipfs peer service
func (r *PeerReconciler) reconcilePeerService(ctx context.Context, peer *ipfsv1alpha1.Peer) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peer.Name,
			Namespace: peer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err := ctrl.SetControllerReference(peer, svc, r.Scheme); err != nil {
			return err
		}
		r.specPeerService(peer, svc)
		return nil
	})

	return err
}

// specPeerService updates ipfs peer service spec
func (r *PeerReconciler) specPeerService(peer *ipfsv1alpha1.Peer, svc *corev1.Service) {
	labels := peer.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "swarm",
			Port:       4001,
			TargetPort: intstr.FromInt(4001),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       "swarm-udp",
			Port:       4001,
			TargetPort: intstr.FromInt(4002),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			Name:       "api",
			Port:       int32(peer.Spec.APIPort),
			TargetPort: intstr.FromInt(int(peer.Spec.APIPort)),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       "gateway",
			Port:       int32(peer.Spec.GatewayPort),
			TargetPort: intstr.FromInt(int(peer.Spec.GatewayPort)),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	svc.Spec.Selector = labels
}

// reconcilePeerConfig reconciles ipfs peer config map
func (r *PeerReconciler) reconcilePeerConfig(ctx context.Context, peer *ipfsv1alpha1.Peer) error {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peer.Name,
			Namespace: peer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, config, func() error {
		if err := ctrl.SetControllerReference(peer, config, r.Scheme); err != nil {
			return err
		}

		r.specPeerConfig(peer, config)

		return nil
	})

	return err

}

// specPeerConfig updates ipfs peer config spec
func (r *PeerReconciler) specPeerConfig(peer *ipfsv1alpha1.Peer, config *corev1.ConfigMap) {
	config.ObjectMeta.Labels = peer.Labels
	if config.Data == nil {
		config.Data = make(map[string]string)
	}
	config.Data["init_ipfs_config.sh"] = initIPFSConfig
	config.Data["copy_swarm_key.sh"] = copySwarmKey
	config.Data["config_ipfs.sh"] = configIPFS
}

func (r *PeerReconciler) reconcilePeerPVC(ctx context.Context, peer *ipfsv1alpha1.Peer) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peer.Name,
			Namespace: peer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(peer, pvc, r.Scheme); err != nil {
			return err
		}

		r.specPeerPVC(peer, pvc)

		return nil
	})

	return err
}

// specPeerPVC updates ipfs peer persistent volume claim
func (r *PeerReconciler) specPeerPVC(peer *ipfsv1alpha1.Peer, pvc *corev1.PersistentVolumeClaim) {
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
	}
}

// reconcilePeerStatefulSet reconciles ipfs peer statefulset
func (r *PeerReconciler) reconcilePeerStatefulSet(ctx context.Context, peer *ipfsv1alpha1.Peer) error {

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peer.Name,
			Namespace: peer.Namespace,
		},
	}

	client := NewIPFSClient(peer)
	img := client.Image()
	command := client.Command()
	args := client.Args()
	homeDir := client.HomeDir()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(peer, sts, r.Scheme); err != nil {
			return err
		}
		r.specPeerStatefulSet(peer, sts, img, homeDir, command, args)
		return nil
	})

	return err
}

// specPeerStatefulSet updates ipfs peer statefulset spec
func (r *PeerReconciler) specPeerStatefulSet(peer *ipfsv1alpha1.Peer, sts *appsv1.StatefulSet, img, homeDir string, command, args []string) {
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
			Name: "script",
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
			Name:      "script",
			MountPath: shared.PathConfig(homeDir),
		},
	}

	initContainers := []corev1.Container{}

	// copy swarm key before init ipfs
	if peer.Spec.SwarmKeySecret != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "swarm-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: peer.Spec.SwarmKeySecret,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "swarm-key",
			MountPath: shared.PathSecrets(homeDir),
		})

		initContainers = append(initContainers, corev1.Container{
			Name:  "copy-swarm-key",
			Image: "busybox",
			Env: []corev1.EnvVar{
				{
					Name:  EnvIPFSPath,
					Value: shared.PathData(homeDir),
				},
				{
					Name:  EnvSecretsPath,
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
		Image: img,
		Env: []corev1.EnvVar{
			{
				Name:  EnvIPFSPath,
				Value: shared.PathData(homeDir),
			},
			{
				Name:  EnvIPFSInitProfiles,
				Value: strings.Join(initProfiles, ","),
			},
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			fmt.Sprintf("%s/init_ipfs_config.sh", shared.PathConfig(homeDir)),
		},
		VolumeMounts: volumeMounts,
	})

	// config ipfs
	initContainers = append(initContainers, corev1.Container{
		Name:  "config-ipfs",
		Image: img,
		Env: []corev1.EnvVar{
			{
				Name:  EnvIPFSPath,
				Value: shared.PathData(homeDir),
			},
			{
				Name:  EnvIPFSAPIPort,
				Value: fmt.Sprintf("%d", peer.Spec.APIPort),
			},
			{
				Name:  EnvIPFSAPIHost,
				Value: peer.Spec.APIHost,
			},
			{
				Name:  EnvIPFSGatewayPort,
				Value: fmt.Sprintf("%d", peer.Spec.GatewayPort),
			},
			{
				Name:  EnvIPFSGatewayHost,
				Value: peer.Spec.GatewayHost,
			},
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			fmt.Sprintf("%s/config_ipfs.sh", shared.PathConfig(homeDir)),
		},
		VolumeMounts: volumeMounts,
	})

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				InitContainers: initContainers,
				Containers: []corev1.Container{
					{
						Name:  "peer",
						Image: img,
						Env: []corev1.EnvVar{
							{
								Name:  EnvIPFSPath,
								Value: shared.PathData(homeDir),
							},
						},
						Command:      command,
						Args:         args,
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.Peer{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
