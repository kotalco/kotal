package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// BeaconNodeReconciler reconciles a Node object
type BeaconNodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=beaconnodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=beaconnodes/status,verbs=get;update;patch

// Reconcile reconciles Ethereum 2.0 beacon node
func (r *BeaconNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node ethereum2v1alpha1.BeaconNode

	if err = r.Client.Get(context.Background(), req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&node)

	if err = r.reconcileNodeDataPVC(&node); err != nil {
		return
	}

	if err = r.reconcileNodeService(&node); err != nil {
		return
	}

	if err = r.reconcileNodeStatefulset(&node); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the node
func (r *BeaconNodeReconciler) updateLabels(node *ethereum2v1alpha1.BeaconNode) {

	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	node.Labels["name"] = "node"
	node.Labels["client"] = string(node.Spec.Client)
	node.Labels["protocol"] = "ethereum2"
	node.Labels["instance"] = node.Name
}

func (r *BeaconNodeReconciler) reconcileNodeService(node *ethereum2v1alpha1.BeaconNode) error {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, &svc, func() error {
		if err := ctrl.SetControllerReference(node, &svc, r.Scheme); err != nil {
			return err
		}

		r.specNodeService(&svc, node)

		return nil
	})

	return err
}

func (r *BeaconNodeReconciler) specNodeService(svc *corev1.Service, node *ethereum2v1alpha1.BeaconNode) {
	labels := node.GetLabels()

	svc.ObjectMeta.Labels = labels
	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "discovery",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			Name:       "p2p",
			Port:       int32(node.Spec.P2PPort),
			TargetPort: intstr.FromInt(int(node.Spec.P2PPort)),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	if node.Spec.RPCPort != 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "json-rpc",
			Port:       int32(node.Spec.RPCPort),
			TargetPort: intstr.FromInt(int(node.Spec.RPCPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.GRPCPort != 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "grpc",
			Port:       int32(node.Spec.GRPCPort),
			TargetPort: intstr.FromInt(int(node.Spec.GRPCPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if node.Spec.RESTPort != 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       "rest",
			Port:       int32(node.Spec.RESTPort),
			TargetPort: intstr.FromInt(int(node.Spec.RESTPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc.Spec.Selector = labels
}

func (r *BeaconNodeReconciler) reconcileNodeDataPVC(node *ethereum2v1alpha1.BeaconNode) error {
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
func (r *BeaconNodeReconciler) specNodeDataPVC(pvc *corev1.PersistentVolumeClaim, node *ethereum2v1alpha1.BeaconNode) {

	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(node.Spec.Resources.Storage),
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
		StorageClassName: node.Spec.Resources.StorageClass,
	}
}

// reconcileNodeStatefulset reconcile Ethereum 2.0 node
func (r *BeaconNodeReconciler) reconcileNodeStatefulset(node *ethereum2v1alpha1.BeaconNode) error {
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
func (r *BeaconNodeReconciler) specNodeStatefulset(sts *appsv1.StatefulSet, node *ethereum2v1alpha1.BeaconNode, args, command []string, img string) {

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
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPU),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Resources.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Spec.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(node.Spec.Resources.MemoryLimit),
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
func (r *BeaconNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereum2v1alpha1.BeaconNode{}).
		Complete(r)
}
