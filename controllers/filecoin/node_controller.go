package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=filecoin.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=filecoin.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete

// Reconcile reconciles Filecoin network node
func (r *NodeReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var node filecoinv1alpha1.Node

	if err = r.Client.Get(context.Background(), req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.reconcileNodeService(&node); err != nil {
		return
	}

	if err = r.reconcileNodePVC(&node); err != nil {
		return
	}

	if err = r.reconcileNodeStatefulSet(&node); err != nil {
		return
	}

	return
}

// reconcileNodePVC reconciles node pvc
func (r *NodeReconciler) reconcileNodePVC(node *filecoinv1alpha1.Node) error {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(node, pvc, r.Scheme); err != nil {
			return err
		}
		r.specNodePVC(pvc, node)
		return nil
	})

	return err
}

// specNodePVC updates node PVC spec
func (r *NodeReconciler) specNodePVC(pvc *v1.PersistentVolumeClaim, node *filecoinv1alpha1.Node) {
	request := v1.ResourceList{
		v1.ResourceStorage: resource.MustParse(node.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.ObjectMeta.Labels = map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

	pvc.Spec = v1.PersistentVolumeClaimSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteOnce,
		},
		Resources: v1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: node.Spec.Resources.StorageClass,
	}
}

// reconcileNodeService reconciles node service
func (r *NodeReconciler) reconcileNodeService(node *filecoinv1alpha1.Node) error {

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, svc, func() error {
		if err := ctrl.SetControllerReference(node, svc, r.Scheme); err != nil {
			return err
		}
		r.specNodeService(svc, node)
		return nil
	})

	return err
}

// reconcileNodeStatefulSet reconciles node stateful set
func (r *NodeReconciler) reconcileNodeStatefulSet(node *filecoinv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		if err := r.specNodeStatefulSet(sts, node); err != nil {
			return err
		}
		return nil
	})

	return err
}

// specNodeService updates node statefulset spec
func (r *NodeReconciler) specNodeService(svc *v1.Service, node *filecoinv1alpha1.Node) {
	labels := map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []v1.ServicePort{
		{
			Name:       "api",
			Port:       int32(1234),
			TargetPort: intstr.FromInt(1234),
			Protocol:   v1.ProtocolTCP,
		},
	}

	svc.Spec.Selector = labels
}

// specNodeStatefulSet updates node statefulset spec
func (r *NodeReconciler) specNodeStatefulSet(sts *appsv1.StatefulSet, node *filecoinv1alpha1.Node) error {
	labels := map[string]string{
		"name":     "node",
		"instance": node.Name,
	}

	image, err := LotusImage(node.Spec.Network)
	if err != nil {
		return err
	}

	sts.ObjectMeta.Labels = labels

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		ServiceName: node.Name,
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "node",
						Image: image,
						Args:  []string{"daemon"},
						Env: []v1.EnvVar{
							{
								Name:  "LOTUS_PATH",
								Value: "/mnt/data",
							},
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      node.Name,
								MountPath: "/mnt/data",
							},
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPU),
								v1.ResourceMemory: resource.MustParse(node.Spec.Resources.Memory),
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPULimit),
								v1.ResourceMemory: resource.MustParse(node.Spec.Resources.MemoryLimit),
							},
						},
					},
				},
				Volumes: []v1.Volume{
					{
						Name: node.Name,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
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

// SetupWithManager adds reconciler to the manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&filecoinv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
