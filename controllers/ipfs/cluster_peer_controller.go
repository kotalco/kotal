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

// ClusterPeerReconciler reconciles a ClusterPeer object
type ClusterPeerReconciler struct {
	shared.Reconciler
}

var (
	//go:embed init_ipfs_cluster_config.sh
	initIPFSClusterConfig string
)

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=clusterpeers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=clusterpeers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;services;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *ClusterPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var peer ipfsv1alpha1.ClusterPeer

	if err = r.Client.Get(ctx, req.NamespacedName, &peer); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the cluster peer if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		peer.Default()
	}

	shared.UpdateLabels(&peer, "ipfs-cluster-service", "")

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

	// reconcile config map
	if err = r.ReconcileOwned(ctx, &peer, &corev1.ConfigMap{}, func(obj client.Object) error {
		r.specConfigmap(&peer, obj.(*corev1.ConfigMap))
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
		args := client.Args()
		args = append(args, peer.Spec.ExtraArgs.Encode(false)...)
		env := client.Env()
		homeDir := client.HomeDir()

		r.specStatefulset(&peer, obj.(*appsv1.StatefulSet), homeDir, env, command, args)
		return nil
	}); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &peer); err != nil {
		return
	}

	return
}

// updateStatus updates ipfs cluster peer status
func (r *ClusterPeerReconciler) updateStatus(ctx context.Context, peer *ipfsv1alpha1.ClusterPeer) error {
	// TODO: update after multi-client support
	peer.Status.Client = "ipfs-cluster-service"

	if err := r.Status().Update(ctx, peer); err != nil {
		log.FromContext(ctx).Error(err, "unable to update cluster peer status")
		return err
	}

	return nil
}

// specService updates ipfs peer service spec
func (r *ClusterPeerReconciler) specService(peer *ipfsv1alpha1.ClusterPeer, svc *corev1.Service) {
	labels := peer.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "swarm",
			Port:       9096,
			TargetPort: intstr.FromString("swarm"),
		},
		{
			Name:       "swarm-udp",
			Port:       9096,
			TargetPort: intstr.FromString("swarm-udp"),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			// Pinning service API
			// https://ipfscluster.io/documentation/reference/pinsvc_api/
			Name:       "api",
			Port:       5001,
			TargetPort: intstr.FromString("api"),
		},
		{
			// Proxy API
			// https://ipfscluster.io/documentation/reference/proxy/
			Name:       "proxy-api",
			Port:       9095,
			TargetPort: intstr.FromString("proxy-api"),
		},
		{
			// REST API
			//https://ipfscluster.io/documentation/reference/api/
			Name:       "rest",
			Port:       9094,
			TargetPort: intstr.FromString("rest"),
		},
		{
			Name:       "metrics",
			Port:       8888,
			TargetPort: intstr.FromString("metrics"),
		},
		{
			Name:       "tracing",
			Port:       6831,
			TargetPort: intstr.FromString("tracing"),
		},
	}

	svc.Spec.Selector = labels
}

// specConfigmap updates IPFS cluster peer configmap spec
func (r *ClusterPeerReconciler) specConfigmap(peer *ipfsv1alpha1.ClusterPeer, config *corev1.ConfigMap) {
	config.ObjectMeta.Labels = peer.Labels

	if config.Data == nil {
		config.Data = make(map[string]string)
	}

	config.Data["init_ipfs_cluster_config.sh"] = initIPFSClusterConfig
}

// specPVC updates IPFS cluster peer persistent volume claim
func (r *ClusterPeerReconciler) specPVC(peer *ipfsv1alpha1.ClusterPeer, pvc *corev1.PersistentVolumeClaim) {
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
		Resources: corev1.VolumeResourceRequirements{
			Requests: request,
		},
		StorageClassName: peer.Spec.Resources.StorageClass,
	}
}

// specStatefulset updates IPFS cluster peer statefulset
func (r *ClusterPeerReconciler) specStatefulset(peer *ipfsv1alpha1.ClusterPeer, sts *appsv1.StatefulSet, homeDir string, env []corev1.EnvVar, command, args []string) {
	labels := peer.Labels

	sts.Labels = labels

	// environment variables required by `ipfs-cluster-service init`
	initClusterPeerENV := []corev1.EnvVar{
		{
			Name:  ipfsClients.EnvIPFSClusterPath,
			Value: shared.PathData(homeDir),
		},
		{
			Name:  ipfsClients.EnvIPFSClusterConsensus,
			Value: string(peer.Spec.Consensus),
		},
		{
			Name:  ipfsClients.EnvIPFSClusterPeerEndpoint,
			Value: peer.Spec.PeerEndpoint,
		},
		{
			Name: ipfsClients.EnvIPFSClusterSecret,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: peer.Spec.ClusterSecretName,
					},
					Key: "secret",
				},
			},
		},
		{
			Name:  ipfsClients.EnvIPFSClusterTrustedPeers,
			Value: strings.Join(peer.Spec.TrustedPeers, ","),
		},
	}

	// if cluster peer ID (which implies private key) is provided
	// append cluster id and private key environment variables
	if peer.Spec.ID != "" {
		// cluster id
		initClusterPeerENV = append(initClusterPeerENV, corev1.EnvVar{
			Name:  ipfsClients.EnvIPFSClusterId,
			Value: peer.Spec.ID,
		})
		// cluster private key
		initClusterPeerENV = append(initClusterPeerENV, corev1.EnvVar{
			Name: ipfsClients.EnvIPFSClusterPrivateKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: peer.Spec.PrivateKeySecretName,
					},
					Key: "key",
				},
			},
		})
	}

	ports := []corev1.ContainerPort{
		{
			Name:          "swarm",
			ContainerPort: 9096,
		},
		{
			Name:          "swarm-udp",
			ContainerPort: 9096,
			Protocol:      corev1.ProtocolUDP,
		},
		{
			Name:          "api",
			ContainerPort: 5001,
		},
		{
			Name:          "proxy-api",
			ContainerPort: 9095,
		},
		{
			Name:          "rest",
			ContainerPort: 9094,
		},
		{
			Name:          "metrics",
			ContainerPort: 8888,
		},
		{
			Name:          "tracing",
			ContainerPort: 6831,
		},
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
				InitContainers: []corev1.Container{
					{
						Name:    "init-cluster-peer",
						Image:   peer.Spec.Image,
						Command: []string{"/bin/sh"},
						Env:     initClusterPeerENV,
						Args: []string{
							fmt.Sprintf("%s/init_ipfs_cluster_config.sh", shared.PathConfig(homeDir)),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: shared.PathData(homeDir),
							},
							{
								Name:      "config",
								MountPath: shared.PathConfig(homeDir),
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:    "cluster-peer",
						Image:   peer.Spec.Image,
						Command: command,
						Env:     env,
						Args:    args,
						Ports:   ports,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: shared.PathData(homeDir),
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(peer.Spec.CPU),
								corev1.ResourceMemory: resource.MustParse(peer.Spec.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(peer.Spec.CPULimit),
								corev1.ResourceMemory: resource.MustParse(peer.Spec.MemoryLimit),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
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
				},
			},
		},
	}
}

func (r *ClusterPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.ClusterPeer{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
