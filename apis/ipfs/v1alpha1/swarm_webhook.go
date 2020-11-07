package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var swarmlog = logf.Log.WithName("swarm-resource")

// SetupWebhookWithManager registers webhook to be started wth the given manager
func (s *Swarm) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-ipfs-kotal-io-v1alpha1-swarm,mutating=true,failurePolicy=fail,groups=ipfs.kotal.io,resources=swarms,verbs=create;update,versions=v1alpha1,name=mswarm.kb.io

var _ webhook.Defaulter = &Swarm{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (s *Swarm) Default() {
	swarmlog.Info("default", "name", s.Name)

	for i := range s.Spec.Nodes {
		s.DefaultNode(&s.Spec.Nodes[i])
	}
}

// DefaultNode defaults a single ipfs node spec
func (s *Swarm) DefaultNode(node *Node) {

	if node.Resources.CPU == "" {
		node.Resources.CPU = DefaultNodeCPURequest
	}

	if node.Resources.CPULimit == "" {
		node.Resources.CPULimit = DefaultNodeCPULimit
	}

	if node.Resources.Memory == "" {
		node.Resources.Memory = DefaultNodeMemoryRequest
	}

	if node.Resources.MemoryLimit == "" {
		node.Resources.MemoryLimit = DefaultNodeMemoryLimit
	}

	if node.Resources.Storage == "" {
		node.Resources.Storage = DefaultNodeStorageRequest
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-swarm,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=swarms,versions=v1alpha1,name=vswarm.kb.io

var _ webhook.Validator = &Swarm{}

// ValidateNodeNameUniqeness validates that all node names are unique
func (s *Swarm) ValidateNodeNameUniqeness() field.ErrorList {

	var uniquenessErrors field.ErrorList
	names := map[string]int{}
	msg := "already used by spec.nodes[%d].name"
	nodesPath := field.NewPath("spec").Child("nodes")

	for i, node := range s.Spec.Nodes {
		if j, exists := names[node.Name]; exists {
			path := nodesPath.Index(i).Child("name")
			err := field.Invalid(path, node.Name, fmt.Sprintf(msg, j))
			uniquenessErrors = append(uniquenessErrors, err)
		} else {
			names[node.Name] = i
		}
	}
	return uniquenessErrors
}

// ValidateNode validates a single ipfs node
func (s *Swarm) ValidateNode(i int) field.ErrorList {
	var nodeErrors field.ErrorList
	node := s.Spec.Nodes[i]
	nodePath := field.NewPath("spec").Child("nodes").Index(i)

	cpu := resource.MustParse(node.Resources.CPU)
	cpuLimit := resource.MustParse(node.Resources.CPULimit)

	// validate cpuLimit can't be less than cpu request
	if cpuLimit.Cmp(cpu) == -1 {
		msg := fmt.Sprintf("must be greater than or equal to cpu %s", string(node.Resources.CPU))
		err := field.Invalid(nodePath.Child("resources").Child("cpuLimit"), node.Resources.CPULimit, msg)
		nodeErrors = append(nodeErrors, err)
	}

	memory := resource.MustParse(node.Resources.Memory)
	memoryLimit := resource.MustParse(node.Resources.MemoryLimit)

	// validate memoryLimit can't be less than memory request
	if memoryLimit.Cmp(memory) == -1 {
		msg := fmt.Sprintf("must be greater than or equal to memory %s", string(node.Resources.Memory))
		err := field.Invalid(nodePath.Child("resources").Child("memoryLimit"), node.Resources.MemoryLimit, msg)
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors
}

// Validate is the shared validation between create and update
func (s *Swarm) Validate() field.ErrorList {
	var allErrors field.ErrorList

	for i := range s.Spec.Nodes {
		allErrors = append(allErrors, s.ValidateNode(i)...)
	}

	allErrors = append(allErrors, s.ValidateNodeNameUniqeness()...)

	return allErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (s *Swarm) ValidateCreate() error {
	var allErrors field.ErrorList

	swarmlog.Info("validate create", "name", s.Name)

	allErrors = append(allErrors, s.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, s.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (s *Swarm) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	swarmlog.Info("validate update", "name", s.Name)

	allErrors = append(allErrors, s.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, s.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (s *Swarm) ValidateDelete() error {
	swarmlog.Info("validate delete", "name", s.Name)

	return nil
}
