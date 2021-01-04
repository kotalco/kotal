package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Ethereum 2.0 node validation", func() {

	createCases := []struct {
		Title  string
		Node   *Node
		Errors field.ErrorList
	}{
		{
			Title: "Node #1",
			Node: &Node{
				Spec: NodeSpec{
					Join:   "mainnet",
					Client: PrysmClient,
					REST:   true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rest",
					BadValue: true,
					Detail:   "not supported by prysm client",
				},
			},
		},
		{
			Title: "Node #2",
			Node: &Node{
				Spec: NodeSpec{
					Join:   "mainnet",
					Client: NimbusClient,
					REST:   true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rest",
					BadValue: true,
					Detail:   "not supported by nimbus client",
				},
			},
		},
	}

	updateCases := []struct {
		Title   string
		OldNode *Node
		NewNode *Node
		Errors  field.ErrorList
	}{}

	Context("While creating network", func() {
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

	Context("While updating network", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
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
