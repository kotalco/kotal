package controllers

import (
	"context"

	bitcoinClients "github.com/kotalco/kotal/clients/bitcoin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=bitcoin.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bitcoin.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile Bitcoin node
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node bitcoinv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "bitcoind", string(node.Spec.Network))

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileService(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &node); err != nil {
		return
	}

	return
}

// updateStatus updates Bitcoin node status
func (r *NodeReconciler) updateStatus(ctx context.Context, node *bitcoinv1alpha1.Node) error {
	node.Status.Client = "bitcoincore"

	if err := r.Status().Update(ctx, node); err != nil {
		log.FromContext(ctx).Error(err, "unable to update node status")
		return err
	}

	return nil
}

// reconcilePVC reconciles Bitcoin node persistent volume claim
func (r *NodeReconciler) reconcilePVC(ctx context.Context, node *bitcoinv1alpha1.Node) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(node, pvc, r.Scheme); err != nil {
			return err
		}

		r.specPVC(node, pvc)

		return nil
	})

	return err
}

// specPVC updates Bitcoin node persistent volume claim
func (r *NodeReconciler) specPVC(node *bitcoinv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
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

// reconcileService reconciles Bitcoin node service
func (r *NodeReconciler) reconcileService(ctx context.Context, node *bitcoinv1alpha1.Node) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err := ctrl.SetControllerReference(node, svc, r.Scheme); err != nil {
			return err
		}
		r.specService(node, svc)
		return nil
	})

	return err
}

// specService updates Bitcoin node service spec
func (r *NodeReconciler) specService(node *bitcoinv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	if node.Spec.RPC {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromInt(int(node.Spec.RPCPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc.Spec.Selector = labels
}

// reconcileStatefulset reconciles node statefulset
func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *bitcoinv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := bitcoinClients.NewClient(node, r.Client)

	homeDir := client.HomeDir()
	cmd := client.Command()
	args := client.Args()
	env := client.Env()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specStatefulSet(node, sts, homeDir, env, cmd, args); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *bitcoinv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, env []corev1.EnvVar, cmd, args []string) error {

	sts.ObjectMeta.Labels = node.Labels

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
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   node.Spec.Image,
						Command: cmd,
						Args:    args,
						Env:     env,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: shared.PathData(homeDir),
							},
						},
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

	return nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&bitcoinv1alpha1.Node{}).
		WithEventFilter(pred).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
