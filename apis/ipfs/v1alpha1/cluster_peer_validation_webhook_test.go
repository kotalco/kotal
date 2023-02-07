package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("IPFS cluster peer validation", func() {
	createCases := []struct {
		Title  string
		Peer   *ClusterPeer
		Errors field.ErrorList
	}{
		{
			Title: "Cluster Peer #1",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					ID: "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.privateKeySecretName",
					BadValue: "",
					Detail:   "must provide privateKeySecretName if id is provided",
				},
			},
		},
		{
			Title: "Cluster Peer #1",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "my-cluster-privatekey",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: "",
					Detail:   "must provide id if privateKeySecretName is provided",
				},
			},
		},
	}

	updateCases := []struct {
		Title   string
		Peer    *ClusterPeer
		NewPeer *ClusterPeer
		Errors  field.ErrorList
	}{
		{
			Title: "Cluster Peer #1",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
					PrivateKeySecretName: "my-cluster-privatekey",
				},
			},
			NewPeer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					ID: "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.privateKeySecretName",
					BadValue: "",
					Detail:   "must provide privateKeySecretName if id is provided",
				},
			},
		},
		{
			Title: "Cluster Peer #2",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "my-cluster-privatekey",
					ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
				},
			},
			NewPeer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "my-cluster-privatekey",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: "",
					Detail:   "must provide id if privateKeySecretName is provided",
				},
			},
		},
		{
			Title: "Cluster Peer #3",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					Consensus: CRDT,
				},
			},
			NewPeer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					Consensus: Raft,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: Raft,
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "Cluster Peer #4",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "my-cluster-privatekey",
					ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuiq",
				},
			},
			NewPeer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "my-cluster-privatekey",
					ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuir",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuir",
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "Cluster Peer #5",
			Peer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "my-cluster-privatekey",
					ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuir",
				},
			},
			NewPeer: &ClusterPeer{
				Spec: ClusterPeerSpec{
					PrivateKeySecretName: "your-cluster-privatekey",
					ID:                   "12D3KooWBcEtY8GH4mNkri9kM3haeWhEXtQV7mi81ErWrqLYGuir",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.privateKeySecretName",
					BadValue: "your-cluster-privatekey",
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating cluster peer", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Peer.Default()
					err := cc.Peer.ValidateCreate()

					// all test cases has validation errors
					Expect(err).NotTo(BeNil())

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

	Context("While updating cluster peer", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Peer.Default()
					cc.NewPeer.Default()
					err := cc.NewPeer.ValidateUpdate(cc.Peer)

					// all test cases has validation errors
					Expect(err).NotTo(BeNil())

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})
})
