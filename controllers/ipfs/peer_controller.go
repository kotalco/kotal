package controllers

import (
	"context"
	_ "embed"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
)

// PeerReconciler reconciles a Peer object
type PeerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//go:embed init_ipfs_config.sh
var initIPFSConfig string

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=peers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=peers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *PeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var peer ipfsv1alpha1.Peer

	if err = r.Client.Get(ctx, req.NamespacedName, &peer); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&peer)

	if err = r.reconcilePeerConfig(ctx, &peer); err != nil {
		return
	}

	if err = r.reconcilePeerPVC(ctx, &peer); err != nil {
		return
	}

	if err = r.reconcilePeerStatefulSet(ctx, &peer); err != nil {
		return
	}

	return
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
		if pvc.CreationTimestamp.IsZero() {
			r.specPeerPVC(peer, pvc)
		}
		return nil
	})

	return err
}

// specPeerPVC updates ipfs peer persistent volume claim
func (r *PeerReconciler) specPeerPVC(peer *ipfsv1alpha1.Peer, pvc *corev1.PersistentVolumeClaim) {
	pvc.ObjectMeta.Labels = peer.Labels

	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(peer.Spec.Resources.Storage),
			},
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

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(peer, sts, r.Scheme); err != nil {
			return err
		}
		r.specPeerStatefulSet(peer, sts)
		return nil
	})

	return err
}

// specPeerStatefulSet updates ipfs peer statefulset spec
func (r *PeerReconciler) specPeerStatefulSet(peer *ipfsv1alpha1.Peer, sts *appsv1.StatefulSet) {
	labels := peer.Labels

	sts.ObjectMeta.Labels = labels

	initIPFS := corev1.Container{
		Name:    "init-ipfs",
		Image:   "ipfs/go-ipfs:v0.8.0",
		Command: []string{"/bin/sh"},
		Args:    []string{"/scripts/init_ipfs_config.sh"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: "/data/ipfs",
			},
			{
				Name:      "script",
				MountPath: "/scripts",
			},
		},
	}

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					initIPFS,
				},
				Containers: []corev1.Container{
					{
						Name:    "peer",
						Image:   "ipfs/go-ipfs:v0.8.0",
						Command: []string{"ipfs"},
						Args:    []string{"daemon"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/data/ipfs",
							},
						},
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
						Name: "script",
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

// SetupWithManager registers the controller to be started with the given manager
func (r *PeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.Peer{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
