package controllers

import (
	"context"
	_ "embed"
	"fmt"

	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	nearClients "github.com/kotalco/kotal/clients/near"
	"github.com/kotalco/kotal/controllers/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	//go:embed init_near_node.sh
	InitNearNode string
)

// +kubebuilder:rbac:groups=near.kotal.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=near.kotal.io,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=watch;get;list;create;update;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var node nearv1alpha1.Node

	if err = r.Client.Get(ctx, req.NamespacedName, &node); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	// default the node if webhooks are disabled
	if !shared.IsWebhookEnabled() {
		node.Default()
	}

	shared.UpdateLabels(&node, "nearcore")

	if err = r.reconcilePVC(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileConfigmap(ctx, &node); err != nil {
		return
	}

	if err = r.reconcileStatefulset(ctx, &node); err != nil {
		return
	}

	return
}

// reconcilePVC reconciles NEAR node persistent volume claim
func (n *NodeReconciler) reconcilePVC(ctx context.Context, node *nearv1alpha1.Node) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, n.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(node, pvc, n.Scheme); err != nil {
			return err
		}

		n.specPVC(node, pvc)

		return nil
	})

	return err
}

// specPVC updates NEAR node persistent volume claim
func (n *NodeReconciler) specPVC(peer *nearv1alpha1.Node, pvc *corev1.PersistentVolumeClaim) {
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
		StorageClassName: peer.Spec.Resources.StorageClass,
	}
}

// specConfigmap updates node configmap
func (n *NodeReconciler) specConfigmap(node *nearv1alpha1.Node, configmap *corev1.ConfigMap) {
	configmap.ObjectMeta.Labels = node.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	configmap.Data["init_near_node.sh"] = InitNearNode

}

// reconcileConfigmap reconciles node configmap
func (r *NodeReconciler) reconcileConfigmap(ctx context.Context, node *nearv1alpha1.Node) error {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(node, configmap, r.Scheme); err != nil {
			return err
		}
		r.specConfigmap(node, configmap)
		return nil
	})

	return err
}

func (r *NodeReconciler) reconcileStatefulset(ctx context.Context, node *nearv1alpha1.Node) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
		},
	}

	client := nearClients.NewClient(node)

	img := client.Image()
	homeDir := client.HomeDir()
	args := client.Args()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		if err := ctrl.SetControllerReference(node, sts, r.Scheme); err != nil {
			return err
		}
		r.specStatefulSet(node, sts, img, homeDir, args)
		return nil
	})

	return err
}

// specStatefulSet updates node statefulset spec
func (r *NodeReconciler) specStatefulSet(node *nearv1alpha1.Node, sts *appsv1.StatefulSet, img, homeDir string, args []string) {

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
				// TODO: use shared security context
				// TODO: update paths to use shared path pkg and non-root directories
				InitContainers: []corev1.Container{
					{
						Name:  "init-near-node",
						Image: img,
						Env: []corev1.EnvVar{
							{
								Name:  EnvDataPath,
								Value: homeDir,
							},
							{
								Name:  EnvNetwork,
								Value: node.Spec.Network,
							},
						},
						Command: []string{"/bin/sh"},
						Args:    []string{fmt.Sprintf("%s/init_near_node.sh", "/root/config")},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: homeDir,
							},
							{
								Name:      "config",
								MountPath: "/root/config",
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:  "node",
						Image: img,
						Args:  args,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data",
								MountPath: homeDir,
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
				},
			},
		},
	}

}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nearv1alpha1.Node{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
