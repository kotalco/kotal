/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var networklog = logf.Log.WithName("network-resource")

// SetupWebhookWithManager sets up the webook with a given controller manager
func (r *Network) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-ethereum-kotal-io-v1alpha1-network,mutating=true,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,verbs=create;update,versions=v1alpha1,name=mnetwork.kb.io

var _ webhook.Defaulter = &Network{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Network) Default() {
	networklog.Info("default", "name", r.Name)

	for i := range r.Spec.Nodes {
		// default only p2p port, leave the rest to the client
		// TODO: change this after multi-client support
		if r.Spec.Nodes[i].P2PPort == 0 {
			r.Spec.Nodes[i].P2PPort = 30303
		}
	}

	if r.Spec.Genesis.Coinbase == "" {
		r.Spec.Genesis.Coinbase = "0x0000000000000000000000000000000000000000"
	}

	if r.Spec.Genesis.Difficulty == "" {
		r.Spec.Genesis.Difficulty = "0x1"
	}

	if r.Spec.Genesis.MixHash == "" {
		r.Spec.Genesis.MixHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
	}

	if r.Spec.Genesis.GasLimit == "" {
		r.Spec.Genesis.GasLimit = "0x47b760"
	}

	if r.Spec.Genesis.Nonce == "" {
		r.Spec.Genesis.Nonce = "0x0"
	}

	if r.Spec.Genesis.Timestamp == "" {
		r.Spec.Genesis.Timestamp = "0x0"
	}

	if r.Spec.Consensus == ProofOfAuthority {
		if r.Spec.Genesis.Clique.BlockPeriod == 0 {
			r.Spec.Genesis.Clique.BlockPeriod = 15
		}
		if r.Spec.Genesis.Clique.EpochLength == 0 {
			r.Spec.Genesis.Clique.EpochLength = 3000
		}
	}

	if r.Spec.Consensus == IstanbulBFT {
		if r.Spec.Genesis.IBFT2.BlockPeriod == 0 {
			r.Spec.Genesis.IBFT2.BlockPeriod = 15
		}
		if r.Spec.Genesis.IBFT2.EpochLength == 0 {
			r.Spec.Genesis.IBFT2.EpochLength = 3000
		}
		if r.Spec.Genesis.IBFT2.RequestTimeout == 0 {
			r.Spec.Genesis.IBFT2.RequestTimeout = 10
		}
		if r.Spec.Genesis.IBFT2.MessageQueueLimit == 0 {
			r.Spec.Genesis.IBFT2.MessageQueueLimit = 1000
		}
		if r.Spec.Genesis.IBFT2.DuplicateMesageLimit == 0 {
			r.Spec.Genesis.IBFT2.DuplicateMesageLimit = 100
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesLimit == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesLimit = 1000
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance = 10
		}

	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum-kotal-io-v1alpha1-network,mutating=false,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,versions=v1alpha1,name=vnetwork.kb.io

var _ webhook.Validator = &Network{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateCreate() error {
	networklog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateUpdate(old runtime.Object) error {
	networklog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateDelete() error {
	networklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
