package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Stacks node validation", func() {
	createCases := []struct {
		Title  string
		Node   *Node
		Errors field.ErrorList
	}{
		{
			Title: "missing seedPrivateKeySecretName",
			Node: &Node{
				Spec: NodeSpec{
					Network: Mainnet,
					Miner:   true,
				},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.seedPrivateKeySecretName",
					BadValue: "",
					Detail:   "seedPrivateKeySecretName is required if node is miner",
				},
			},
		},
		{
			Title: "seedPrivateKeySecretName is given for non miner node",
			Node: &Node{
				Spec: NodeSpec{
					Network:                  Mainnet,
					SeedPrivateKeySecretName: "seed-private-key",
				},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.miner",
					BadValue: false,
					Detail:   "node must be a miner if seedPrivateKeySecretName is given",
				},
			},
		},
		{
			Title: "mineMicroblocks is given for non miner node",
			Node: &Node{
				Spec: NodeSpec{
					Network:         Mainnet,
					MineMicroblocks: true,
				},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.miner",
					BadValue: false,
					Detail:   "node must be a miner if mineMicroblocks is true",
				},
			},
		},
	}

	updateCases := []struct {
		Title   string
		OldNode *Node
		NewNode *Node
		Errors  field.ErrorList
	}{
		{
			Title: "updated network",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: "mainnet",
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: "testnet",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: "testnet",
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating node", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Node.Default()
					_, err := cc.Node.ValidateCreate()

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

	Context("While updating node", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.OldNode.Default()
					cc.NewNode.Default()
					_, err := cc.NewNode.ValidateUpdate(cc.OldNode)

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

})
