package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-filecoin-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=filecoin.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-filecoin-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (n *Node) Default() {
	nodelog.Info("default", "name", n.Name)

	mainnet := n.Spec.Network == MainNetwork
	calibration := n.Spec.Network == CalibrationNetwork

	if n.Spec.Image == "" {
		var image string

		switch n.Spec.Network {
		case MainNetwork:
			image = DefaultLotusImage
		case CalibrationNetwork:
			image = DefaultLotusCalibrationImage
		}

		n.Spec.Image = image
	}

	if n.Spec.Logging == "" {
		n.Spec.Logging = DefaultLogging
	}

	if n.Spec.APIPort == 0 {
		n.Spec.APIPort = DefaultAPIPort
	}
	if n.Spec.APIRequestTimeout == 0 {
		n.Spec.APIRequestTimeout = DefaultAPIRequestTimeout
	}

	if n.Spec.P2PPort == 0 {
		n.Spec.P2PPort = DefaultP2PPort
	}

	if n.Spec.Resources.CPU == "" {
		if mainnet {
			n.Spec.CPU = DefaultMainnetNodeCPURequest
		}
		if calibration {
			n.Spec.CPU = DefaultCalibrationNodeCPURequest
		}
	}

	if n.Spec.CPULimit == "" {
		if mainnet {
			n.Spec.CPULimit = DefaultMainnetNodeCPULimit
		}
		if calibration {
			n.Spec.CPULimit = DefaultCalibrationNodeCPULimit
		}
	}

	if n.Spec.Memory == "" {
		if mainnet {
			n.Spec.Memory = DefaultMainnetNodeMemoryRequest
		}
		if calibration {
			n.Spec.Memory = DefaultCalibrationNodeMemoryRequest
		}
	}

	if n.Spec.MemoryLimit == "" {
		if mainnet {
			n.Spec.MemoryLimit = DefaultMainnetNodeMemoryLimit
		}
		if calibration {
			n.Spec.MemoryLimit = DefaultCalibrationNodeMemoryLimit
		}
	}

	if n.Spec.Storage == "" {
		if mainnet {
			n.Spec.Storage = DefaultMainnetNodeStorageRequest
		}
		if calibration {
			n.Spec.Storage = DefaultCalibrationNodeStorageRequest
		}
	}

}
