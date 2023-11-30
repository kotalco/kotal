package controllers

import (
	"context"

	stacksClients "github.com/kotalco/kotal/clients/stacks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	shared.Reconciler
}

// +kubebuilder:rbac:groups=stacks.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stacks.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	defer shared.IgnoreConflicts(&err)

	var node stacksv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "stacks-node", string(node.Spec.Network))

	// reconcile config map
	if err = r.ReconcileOwned(ctx, &node, &corev1.ConfigMap{}, func(obj client.Object) error {
		configToml, err := ConfigFromSpec(&node, r.Client)
		if err != nil {
			return err
		}
		r.specConfigmap(&node, obj.(*corev1.ConfigMap), configToml)
		return nil
	}); err != nil {
		return
	}

	// reconcile persistent volume claim
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

	// reconcile stateful set
	if err = r.ReconcileOwned(ctx, &node, &appsv1.StatefulSet{}, func(obj client.Object) error {
		client := stacksClients.NewClient(&node)

		homeDir := client.HomeDir()
		cmd := client.Command()
		args := client.Args()
		args = append(args, node.Spec.ExtraArgs.Encode(false)...)
		env := client.Env()

		return r.specStatefulSet(&node, obj.(*appsv1.StatefulSet), homeDir, env, cmd, args)
	}); err != nil {
		return
	}

	if err = r.updateStatus(ctx, &node); err != nil {
		return
	}

	return
}

// updateStatus updates Stacks node status
func (r *NodeReconciler) updateStatus(ctx context.Context, node *stacksv1alpha1.Node) error {
	node.Status.Client = "stacks"

	if err := r.Status().Update(ctx, node); err != nil {
		log.FromContext(ctx).Error(err, "unable to update node status")
		return err
	}

	return nil
}

// specConfigmap updates node statefulset spec
func (r *NodeReconciler) specConfigmap(node *stacksv1alpha1.Node, configmap *corev1.ConfigMap, configToml string) {
	configmap.ObjectMeta.Labels = node.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	configmap.Data["config.toml"] = configToml

}

// specPVC updates Stacks node persistent volume claim
func (r *NodeReconciler) specPVC(node *stacksv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
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

// specService updates Bitcoin node service spec
func (r *NodeReconciler) specService(node *stacksv1alpha1.Node, svc *corev1.Service) {
	labels := node.Labels

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromString("p2p"),
		},
		{
			Name:       "rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromString("rpc"),
		},
	}

	svc.Spec.Selector = labels
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *stacksv1alpha1.Node, sts *appsv1.StatefulSet, homeDir string, env []corev1.EnvVar, cmd, args []string) error {

	sts.ObjectMeta.Labels = node.Labels

	ports := []corev1.ContainerPort{
		{
			Name:          "p2p",
			ContainerPort: int32(node.Spec.P2PPort),
		},
		{
			Name:          "rpc",
			ContainerPort: int32(node.Spec.RPCPort),
		},
	}

	replicas := int32(*node.Spec.Replicas)

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.Labels,
		},
		ServiceName: node.Name,
		Replicas:    &replicas,
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
						Ports:   ports,
						Env:     env,
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
								Name:      "data",
								MountPath: shared.PathData(homeDir),
							},
							{
								Name:      "config",
								ReadOnly:  true,
								MountPath: shared.PathConfig(homeDir),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: node.Name,
								},
							},
						},
					},
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
		For(&stacksv1alpha1.Node{}).
		WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
