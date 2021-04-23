package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// ClusterPeerReconciler reconciles a ClusterPeer object
type ClusterPeerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=clusterpeers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=clusterpeers/status,verbs=get;update;patch

func (r *ClusterPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {

	var peer ipfsv1alpha1.ClusterPeer

	if err = r.Client.Get(ctx, req.NamespacedName, &peer); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the cluster peer if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		peer.Default()
	}

	r.updateLabels(&peer)

	if err = r.reconcileClusterPeerPVC(ctx, &peer); err != nil {
		return
	}

	if err = r.reconcileClusterPeerStatefulset(ctx, &peer); err != nil {
		return
	}

	return
}

// updateLabels updates IPFS cluster peer labels
func (r *ClusterPeerReconciler) updateLabels(peer *ipfsv1alpha1.ClusterPeer) {
	if peer.Labels == nil {
		peer.Labels = map[string]string{}
	}

	peer.Labels["name"] = "cluster-peer"
	peer.Labels["protocol"] = "ipfs"
	peer.Labels["client"] = "ipfs-cluster-service"
	peer.Labels["instance"] = peer.Name
}

// reconcileClusterPeerPVC reconciles IPFS cluster peer persistent volume claim
func (r *ClusterPeerReconciler) reconcileClusterPeerPVC(ctx context.Context, peer *ipfsv1alpha1.ClusterPeer) error {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peer.Name,
			Namespace: peer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		if err := ctrl.SetControllerReference(peer, &pvc, r.Scheme); err != nil {
			return err
		}

		r.specClusterPeerPVC(peer, &pvc)
		return nil
	})

	return err
}

// specClusterPeerPVC updates IPFS cluster peer persistent volume claim
func (r *ClusterPeerReconciler) specClusterPeerPVC(peer *ipfsv1alpha1.ClusterPeer, pvc *corev1.PersistentVolumeClaim) {
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

// reconcileClusterPeerStatefulset reconciles IPFS cluster peer statefulset
func (r *ClusterPeerReconciler) reconcileClusterPeerStatefulset(ctx context.Context, peer *ipfsv1alpha1.ClusterPeer) error {
	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      peer.Name,
			Namespace: peer.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
		if err := ctrl.SetControllerReference(peer, &sts, r.Scheme); err != nil {
			return err
		}

		r.specClusterPeerStatefulset(peer, &sts)

		return nil
	})

	return err
}

// specClusterPeerStatefulset updates IPFS cluster peer statefulset
func (r *ClusterPeerReconciler) specClusterPeerStatefulset(peer *ipfsv1alpha1.ClusterPeer, sts *appsv1.StatefulSet) {
	labels := peer.Labels

	sts.Labels = labels

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
					{
						Name:    "init-cluster-peer",
						Image:   "ipfs/ipfs-cluster:v0.13.2",
						Command: []string{"ipfs-cluster-service"},
						Args:    []string{"init"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/data/ipfs-cluster",
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:    "cluster-peer",
						Image:   "ipfs/ipfs-cluster:v0.13.2",
						Command: []string{"ipfs-cluster-service"},
						Args:    []string{"daemon"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/data/ipfs-cluster",
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
				},
			},
		},
	}
}

func (r *ClusterPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.ClusterPeer{}).
		Complete(r)
}