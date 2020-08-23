package controllers

import (
	"bytes"
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
	"text/template"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
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
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=watch;get;create;update;list;delete

// Reconcile reconciles ipfs swarm
func (r *SwarmReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	var _ = context.Background()
	_ = r.Log.WithValues("swarm", req.NamespacedName)

	var swarm ipfsv1alpha1.Swarm

	if err = r.Client.Get(context.Background(), req.NamespacedName, &swarm); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.updateStatus(&swarm); err != nil {
		return
	}

	if err = r.reconcileNodes(&swarm); err != nil {
		return
	}

	return
}

// updateStatus updates swarm status
func (r *SwarmReconciler) updateStatus(swarm *ipfsv1alpha1.Swarm) error {
	swarm.Status.NodesCount = len(swarm.Spec.Nodes)

	if err := r.Status().Update(context.Background(), swarm); err != nil {
		r.Log.Error(err, "unable to update swarm status")
		return err
	}

	return nil
}

// reconcileNodes reconcile ipfs swarm nodes
func (r *SwarmReconciler) reconcileNodes(swarm *ipfsv1alpha1.Swarm) error {
	peers := []string{}

	for _, node := range swarm.Spec.Nodes {
		addr, err := r.reconcileNode(&node, swarm, peers)
		if err != nil {
			return err
		}
		peers = append(peers, addr)
	}

	if err := r.deleteRedundantNodes(swarm); err != nil {
		return err
	}

	return nil
}

// deleteRedundantNodes deletes redundant ipfs node that has been removed from spec
// swarm is the owner of the redundant resources (node deployment, svc, secret and pvc)
// removing nodes from spec won't remove these resources by grabage collection
// that's why we're deleting them manually
func (r *SwarmReconciler) deleteRedundantNodes(swarm *ipfsv1alpha1.Swarm) error {
	log := r.Log.WithName("delete redundant nodes")

	var deps appsv1.DeploymentList
	var pvcs corev1.PersistentVolumeClaimList
	var services corev1.ServiceList

	nodes := swarm.Spec.Nodes
	names := map[string]bool{}
	matchingLabels := client.MatchingLabels{
		"name":  "node",
		"swarm": swarm.Name,
	}
	inNamespace := client.InNamespace(swarm.Namespace)

	for _, node := range nodes {
		depName := node.DeploymentName(swarm.Name)
		names[depName] = true
	}

	// Node deployments
	if err := r.Client.List(context.Background(), &deps, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node deployments")
		return err
	}

	for _, dep := range deps.Items {
		name := dep.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) deployment", name))

			if err := r.Client.Delete(context.Background(), &dep); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) deployment", name))
				return err
			}
		}
	}

	// Node PVCs
	if err := r.Client.List(context.Background(), &pvcs, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node pvcs")
		return err
	}

	for _, pvc := range pvcs.Items {
		name := pvc.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) pvc", name))

			if err := r.Client.Delete(context.Background(), &pvc); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) pvc", name))
				return err
			}
		}
	}

	// Node Services
	if err := r.Client.List(context.Background(), &services, matchingLabels, inNamespace); err != nil {
		log.Error(err, "unable to list all node services")
		return err
	}

	for _, service := range services.Items {
		name := service.GetName()
		if exist := names[name]; !exist {
			log.Info(fmt.Sprintf("deleting node (%s) service", name))

			if err := r.Client.Delete(context.Background(), &service); err != nil {
				log.Error(err, fmt.Sprintf("unable to delete node (%s) service", name))
				return err
			}
		}
	}

	return nil
}

// reconcileNode reconciles a single ipfs node
// it creates node deployment, service and data pvc if it doesn't exist
func (r *SwarmReconciler) reconcileNode(node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm, peers []string) (addr string, err error) {
	var ip string

	if err = r.reconcileNodePVC(node, swarm); err != nil {
		return
	}

	if err = r.reconcileNodeConfig(node, swarm, peers); err != nil {
		return
	}

	if ip, err = r.reconcileNodeService(node, swarm); err != nil {
		return
	}

	if err = r.reconcileNodeDeployment(node, swarm, peers); err != nil {
		return
	}

	addr = node.SwarmAddress(ip)

	return
}

// reconcileNodeConfig reconciles ipfs node config map
func (r *SwarmReconciler) reconcileNodeConfig(node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm, peers []string) error {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.ConfigName(swarm.Name),
			Namespace: swarm.Namespace,
		},
	}

	script, err := generateInitScript(node, peers)
	if err != nil {
		return err
	}

	_, err = ctrl.CreateOrUpdate(context.Background(), r.Client, config, func() error {
		if err := ctrl.SetControllerReference(swarm, config, r.Scheme); err != nil {
			return err
		}

		r.specNodeConfig(config, node, swarm, script)
		return nil
	})

	return err
}

// generateInitScript generates init script from node spec
func generateInitScript(node *ipfsv1alpha1.Node, peers []string) (script string, err error) {

	type Input struct {
		Profiles []ipfsv1alpha1.Profile
		Peers    []string
	}

	input := &Input{
		Profiles: node.Profiles,
		Peers:    peers,
	}

	tmpl, err := template.New("master").Parse(initScriptTemplate)
	if err != nil {
		return
	}

	buff := new(bytes.Buffer)
	if err = tmpl.Execute(buff, input); err != nil {
		return
	}

	script = buff.String()

	return
}

// specNodeConfig updates node config map
func (r *SwarmReconciler) specNodeConfig(config *corev1.ConfigMap, node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm, script string) {

	config.ObjectMeta.Labels = node.Labels(swarm.Name)
	config.Data = make(map[string]string)
	config.Data["init.sh"] = script

}

// reconcileNodePVC reconciles ipfs node data persistent volume claim
func (r *SwarmReconciler) reconcileNodePVC(node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.PVCName(swarm.Name),
			Namespace: swarm.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, pvc, func() error {
		if err := ctrl.SetControllerReference(swarm, pvc, r.Scheme); err != nil {
			return err
		}
		if pvc.CreationTimestamp.IsZero() {
			r.specNodePVC(pvc, node, swarm)
		}
		return nil
	})

	return err
}

// specNodePVC updates node persistent volume spec
func (r *SwarmReconciler) specNodePVC(pvc *corev1.PersistentVolumeClaim, node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm) {

	pvc.ObjectMeta.Labels = node.Labels(swarm.Name)

	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(node.Resources.Storage),
			},
		},
	}

}

// reconcileNodeService reconciles node service
func (r *SwarmReconciler) reconcileNodeService(node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm) (string, error) {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.ServiceName(swarm.Name),
			Namespace: swarm.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, svc, func() error {
		if err := ctrl.SetControllerReference(swarm, svc, r.Scheme); err != nil {
			return err
		}
		r.specNodeService(svc, node, swarm)
		return nil
	})

	return svc.Spec.ClusterIP, err
}

// specNodeService updates node service spec
func (r *SwarmReconciler) specNodeService(svc *corev1.Service, node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm) {

	labels := node.Labels(swarm.Name)
	svc.ObjectMeta.Labels = labels

	svc.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "swarm",
			Port:       4001,
			TargetPort: intstr.FromInt(4001),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       "swarm-udp",
			Port:       4002,
			TargetPort: intstr.FromInt(4002),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			Name:       "api",
			Port:       5001,
			TargetPort: intstr.FromInt(5001),
			Protocol:   corev1.ProtocolUDP,
		},
		{
			Name:       "gateway",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
			Protocol:   corev1.ProtocolUDP,
		},
	}

	svc.Spec.Selector = labels

}

// reconcileNodeDeployment reconciles node deployment
func (r *SwarmReconciler) reconcileNodeDeployment(node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm, peers []string) error {

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.DeploymentName(swarm.Name),
			Namespace: swarm.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, dep, func() error {
		if err := ctrl.SetControllerReference(swarm, dep, r.Scheme); err != nil {
			return err
		}
		r.specNodeDeployment(dep, node, swarm, peers)
		return nil
	})

	return err
}

// specNodeDeployment updates node deployment spec
func (r *SwarmReconciler) specNodeDeployment(dep *appsv1.Deployment, node *ipfsv1alpha1.Node, swarm *ipfsv1alpha1.Swarm, peers []string) {
	labels := node.Labels(swarm.Name)

	dep.ObjectMeta.Labels = labels

	initNode := corev1.Container{
		Name:  "init-node",
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
		Command: []string{"/bin/sh"},
		Args:    []string{"/script/init.sh"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: "/data/ipfs",
			},
			{
				Name:      "script",
				MountPath: "/script",
			},
		},
	}

	dep.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					initNode,
				},
				Containers: []corev1.Container{
					{
						Name:    "node",
						Image:   "ipfs/go-ipfs:v0.6.0",
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
								corev1.ResourceCPU:    resource.MustParse(node.Resources.CPU),
								corev1.ResourceMemory: resource.MustParse(node.Resources.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(node.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(node.Resources.MemoryLimit),
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: node.PVCName(swarm.Name),
							},
						},
					},
					{
						Name: "script",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: node.ConfigName(swarm.Name),
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
func (r *SwarmReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipfsv1alpha1.Swarm{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
