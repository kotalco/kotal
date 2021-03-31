package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var nodelog = logf.Log.WithName("node-resource")

// SetupWebhookWithManager sets up the webook with a given controller manager
func (n *Node) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(n).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-filecoin-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=filecoin.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-filecoin-v1alpha1-node.kb.io

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (n *Node) Default() {
	nodelog.Info("default", "name", n.Name)

	nerpa := n.Spec.Network == NerpaNetwork
	mainnet := n.Spec.Network == MainNetwork
	calibration := n.Spec.Network == CalibrationNetwork
	butterfly := n.Spec.Network == ButterflyNetwork

	if n.Spec.Resources.CPU == "" {
		if nerpa {
			n.Spec.CPU = DefaultNerpaNodeCPURequest
		}
		if mainnet {
			n.Spec.CPU = DefaultMainnetNodeCPURequest
		}
		if calibration {
			n.Spec.CPU = DefaultCalibrationNodeCPURequest
		}
		if butterfly {
			n.Spec.CPU = DefaultButterflyNodeCPURequest
		}
	}

	if n.Spec.CPULimit == "" {
		if nerpa {
			n.Spec.CPULimit = DefaultNerpaNodeCPULimit
		}
		if mainnet {
			n.Spec.CPULimit = DefaultMainnetNodeCPULimit
		}
		if calibration {
			n.Spec.CPULimit = DefaultCalibrationNodeCPULimit
		}
		if butterfly {
			n.Spec.CPULimit = DefaultButterflyNodeCPULimit
		}
	}

	if n.Spec.Memory == "" {
		if nerpa {
			n.Spec.Memory = DefaultNerpaNodeMemoryRequest
		}
		if mainnet {
			n.Spec.Memory = DefaultMainnetNodeMemoryRequest
		}
		if calibration {
			n.Spec.Memory = DefaultCalibrationNodeMemoryRequest
		}
		if butterfly {
			n.Spec.Memory = DefaultButterflyNodeMemoryRequest
		}
	}

	if n.Spec.MemoryLimit == "" {
		if nerpa {
			n.Spec.MemoryLimit = DefaultNerpaNodeMemoryLimit
		}
		if mainnet {
			n.Spec.MemoryLimit = DefaultMainnetNodeMemoryLimit
		}
		if calibration {
			n.Spec.MemoryLimit = DefaultCalibrationNodeMemoryLimit
		}
		if butterfly {
			n.Spec.MemoryLimit = DefaultButterflyNodeMemoryLimit
		}
	}

	if n.Spec.Storage == "" {
		if nerpa {
			n.Spec.Storage = DefaultNerpaNodeStorageRequest
		}
		if mainnet {
			n.Spec.Storage = DefaultMainnetNodeStorageRequest
		}
		if calibration {
			n.Spec.Storage = DefaultCalibrationNodeStorageRequest
		}
		if butterfly {
			n.Spec.Storage = DefaultButterflyNodeStorageRequest
		}
	}

}

// +kubebuilder:webhook:verbs=create;update,path=/validate-filecoin-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=filecoin.kotal.io,resources=nodes,versions=v1alpha1,name=validate-filecoin-v1alpha1-node.kb.io

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateCreate() error {
	nodelog.Info("validate create", "name", n.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateUpdate(old runtime.Object) error {
	nodelog.Info("validate update", "name", n.Name)

	var allErrors field.ErrorList

	oldNode := old.(*Node)

	// validate network is immutable
	if oldNode.Spec.Network != n.Spec.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), n.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)

}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", n.Name)

	return nil
}
