package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Polkadot node validation", func() {
	t := true
	createCases := []struct {
		Title  string
		Node   *Node
		Errors field.ErrorList
	}{
		{
			Title: "validator node with rpc enabled",
			Node: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
					RPC:       true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: true,
					Detail:   "must be false if node is validator",
				},
			},
		},
		{
			Title: "validator node with ws enabled",
			Node: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
					WS:        true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.ws",
					BadValue: true,
					Detail:   "must be false if node is validator",
				},
			},
		},
		{
			Title: "validator node with pruning enabled",
			Node: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
					Pruning:   &t,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.pruning",
					BadValue: true,
					Detail:   "must be false if node is validator",
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
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: "kusama",
				},
			},
			NewNode: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: "polkadot",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: "polkadot",
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "enabling rpc for validator node",
			OldNode: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
				},
			},
			NewNode: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
					RPC:       true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: true,
					Detail:   "must be false if node is validator",
				},
			},
		},
		{
			Title: "enabling ws for validator node",
			OldNode: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
				},
			},
			NewNode: &Node{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:   "kusama",
					Validator: true,
					WS:        true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.ws",
					BadValue: true,
					Detail:   "must be false if node is validator",
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
					err := cc.Node.ValidateCreate()

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
					err := cc.NewNode.ValidateUpdate(cc.OldNode)

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

})
