package controllers

import (
	"context"
	"fmt"

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
	"github.com/kotalco/kotal/controllers/shared"
)

// BeaconNodeReconciler reconciles a Node object
type BeaconNodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=beaconnodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=beaconnodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles Ethereum 2.0 beacon node
func (r *BeaconNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node ethereum2v1alpha1.BeaconNode

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the beacon node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	r.updateLabels(&node)

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileService(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the node
func (r *BeaconNodeReconciler) updateLabels(node *ethereum2v1alpha1.BeaconNode) {

	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	node.Labels["app.kubernetes.io/name"] = string(node.Spec.Client)
	node.Labels["app.kubernetes.io/instance"] = node.Name
	node.Labels["app.kubernetes.io/component"] = "ethereum2-beacon-node"
	node.Labels["app.kubernetes.io/managed-by"] = "kotal"
	node.Labels["app.kubernetes.io/created-by"] = "ethereum2-beacon-node-controller"
}

// reconcileService reconciles beacon node service
func (r *BeaconNodeReconciler) reconcileService(ctx context.Context, node *ethereum2v1alpha1.BeaconNode) error {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		if err := ctrl.SetControllerReference(node, &svc, r.Scheme); err != nil {
			return err
		}

		r.specService(node, &svc)

		return nil
	})

	return err
}

func (r *BeaconNodeReconciler) specService(node *ethereum2v1alpha1.BeaconNode, svc *corev1.Service) {
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

// reconcilePVC reconciles beacon node persistent volume claim
func (r *BeaconNodeReconciler) reconcilePVC(ctx context.Context, node *ethereum2v1alpha1.BeaconNode) error {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		if err := ctrl.SetControllerReference(node, &pvc, r.Scheme); err != nil {
			return err
		}

		r.specPVC(node, &pvc)

		return nil
	})

	return err
}

// specPVC updates beacon node persistent volume claim spec
func (r *BeaconNodeReconciler) specPVC(node *ethereum2v1alpha1.BeaconNode, pvc *corev1.PersistentVolumeClaim) {

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

// reconcileStatefulset reconcile Ethereum 2.0 beacon node
func (r *BeaconNodeReconciler) reconcileStatefulset(ctx context.Context, node *ethereum2v1alpha1.BeaconNode) error {
	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
		if err := ctrl.SetControllerReference(node, &sts, r.Scheme); err != nil {
			return err
		}

		client, err := NewEthereum2Client(node)
		if err != nil {
			return err
		}

		args := client.Args()
		img := client.Image()
		command := client.Command()
		homeDir := client.HomeDir()

		r.specStatefulset(node, &sts, args, command, img, homeDir)

		return nil
	})

	return err
}

// specStatefulset updates beacon node statefulset spec
func (r *BeaconNodeReconciler) specStatefulset(node *ethereum2v1alpha1.BeaconNode, sts *appsv1.StatefulSet, args, command []string, img, homeDir string) {

	sts.Labels = node.GetLabels()

	mountPath := shared.PathData(homeDir)

	// Nimbus required changing permission of the data dir to be
	// read and write by owner only
	// that's why we mount volume at $HOME
	// but data dir is atatched at $HOME/kota-data
	if node.Spec.Client == ethereum2v1alpha1.NimbusClient {
		mountPath = homeDir
	}

	mounts := []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: mountPath,
		},
	}

	initContainers := []corev1.Container{}

	// Nimbus client requires data dir path to be read and write only by the owner 0700
	if node.Spec.Client == ethereum2v1alpha1.NimbusClient {
		fixPermissionContainer := corev1.Container{
			Name:  "fix-datadir-permission",
			Image: img,
			Command: []string{
				"/bin/sh",
				"-c",
			},
			Args: []string{
				fmt.Sprintf(`
					mkdir -p %s &&
					chmod 700 %s`,
					shared.PathData(homeDir),
					shared.PathData(homeDir),
				),
			},
			VolumeMounts: mounts,
		}
		initContainers = append(initContainers, fixPermissionContainer)
	}

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: node.GetLabels(),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: node.GetLabels(),
			},
			Spec: corev1.PodSpec{
				SecurityContext: shared.SecurityContext(),
				InitContainers:  initContainers,
				Containers: []corev1.Container{
					{
						Name:         "node",
						Command:      command,
						Args:         args,
						Image:        img,
						VolumeMounts: mounts,
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
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
