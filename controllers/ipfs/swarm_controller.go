package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ipfsv1alpha1 "github.com/mfarghaly/kotal/apis/ipfs/v1alpha1"
)

// SwarmReconciler reconciles a Swarm object
type SwarmReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=swarms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipfs.kotal.io,resources=swarms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=watch;get;list;create;update;delete

// Reconcile reconciles ipfs swarm
func (r *SwarmReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	_ = r.Log.WithValues("swarm", req.NamespacedName)

	var swarm ipfsv1alpha1.Swarm

	if err = r.Client.Get(ctx, req.NamespacedName, &swarm); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.reconcileNodes(ctx, &swarm); err != nil {
		return
	}

	return
}

// reconcileNodes reconcile ipfs swarm nodes
func (r *SwarmReconciler) reconcileNodes(ctx context.Context, swarm *ipfsv1alpha1.Swarm) error {
	for _, node := range swarm.Spec.Nodes {
		if err := r.reconcileNode(ctx, &node, swarm); err != nil {
			return err
		}
	}
	return nil
}

func (r *SwarmReconciler) specNodeDeployment(dep *appsv1.Deployment, node *ipfsv1alpha1.Node) {

	dep.ObjectMeta.Labels = map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

	dep.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name":     "node",
				"instance": node.Name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"name":     "node",
					"instance": node.Name,
				},
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					{
						Name:  "init",
						Image: "kotalco/go-ipfs:v0.6.0",
						Env: []corev1.EnvVar{
							{
								Name:  "IPFS_PEER_ID",
								Value: node.ID,
							},
							{
								Name:  "IPFS_PRIVATE_KEY",
								Value: node.PrivateKey,
							},
						},
						Command: []string{"ipfs"},
						Args:    []string{"init"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/data/ipfs",
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   "kotalco/go-ipfs:v0.6.0",
						Command: []string{"ipfs"},
						Args:    []string{"daemon"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/data/ipfs",
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
				},
			},
		},
	}
}

// reconcileNodes reconcile a single ipfs node
func (r *SwarmReconciler) reconcileNode(ctx context.Context, node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm) error {

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: swarm.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, dep, func() error {
		r.specNodeDeployment(dep, node)
		return nil
	})

	return err
}

// SetupWithManager registers the controller to be started with the given manager
func (r *SwarmReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.Swarm{}).
		Complete(r)
}
