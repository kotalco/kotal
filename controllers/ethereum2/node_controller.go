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

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconcile Ethereum 2.0 node
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node ethereum2v1alpha1.Node

	if err = r.Client.Get(context.Background(), req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&node)

	if err = r.reconcileNodeDataPVC(&node); err != nil {
		return
	}

	if err = r.reconcileNodeStatefulset(&node); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the node
func (r *NodeReconciler) updateLabels(node *ethereum2v1alpha1.Node) {

	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	node.Labels["name"] = "node"
	node.Labels["client"] = string(node.Spec.Client)
	node.Labels["protocol"] = "ethereum2"
	node.Labels["instance"] = node.Name
}

func (r *NodeReconciler) reconcileNodeDataPVC(node *ethereum2v1alpha1.Node) error {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, &pvc, func() error {
		if err := ctrl.SetControllerReference(node, &pvc, r.Scheme); err != nil {
			return err
		}

		r.specNodeDataPVC(&pvc, node)

		return nil
	})

	return err
}

// specNodeDataPVC updates node data PVC spec
func (r *NodeReconciler) specNodeDataPVC(pvc *corev1.PersistentVolumeClaim, node *ethereum2v1alpha1.Node) {

	request := corev1.ResourceList{
		// TODO: update node spec with resources
		corev1.ResourceStorage: resource.MustParse("100Gi"),
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
	}
}

// reconcileNodeStatefulset reconcile Ethereum 2.0 node
func (r *NodeReconciler) reconcileNodeStatefulset(node *ethereum2v1alpha1.Node) error {
	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, &sts, func() error {
		if err := ctrl.SetControllerReference(node, &sts, r.Scheme); err != nil {
			return err
		}

		client, err := NewEthereum2Client(node.Spec.Client)
		if err != nil {
			return err
		}

		args := client.Args(node)
		img := client.Image()
		command := client.Command()

		r.specNodeStatefulset(&sts, node, args, command, img)

		return nil
	})

	return err
}

// specNodeStatefulset updates node statefulset spec
func (r *NodeReconciler) specNodeStatefulset(sts *appsv1.StatefulSet, node *ethereum2v1alpha1.Node, args, command []string, img string) {

	sts.Labels = node.GetLabels()

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.GetLabels(),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: node.GetLabels(),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "node",
						Command: command,
						Args:    args,
						Image:   img,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: PathBlockchainData,
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
				},
			},
		},
	}
}

// SetupWithManager adds reconciler to the manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereum2v1alpha1.Node{}).
		Complete(r)
}
