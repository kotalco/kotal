package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ipfs-kotal-io-v1alpha1-clusterpeer,mutating=true,failurePolicy=fail,groups=ipfs.kotal.io,resources=clusterpeers,verbs=create;update,versions=v1alpha1,name=mutate-ipfs-v1alpha1-clusterpeer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterPeer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterPeer) Default() {
	clusterpeerlog.Info("default", "name", r.Name)
}
