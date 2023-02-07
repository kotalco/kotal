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

var _ = Describe("Aptos node validation", func() {
	createCases := []struct {
		Title  string
		Node   *Node
		Errors field.ErrorList
	}{
		{
			Title: "missing peerId",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:                  Devnet,
					NodePrivateKeySecretName: "my-private-key",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.peerId",
					BadValue: "",
					Detail:   "must provide peerId if nodePrivateKeySecretName is provided",
				},
			},
		},
		{
			Title: "missing nodePrivateKeySecretName",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: Devnet,
					PeerId:  "76c8ca8bb75d1abd853fc17b70cc72cb78a63425fa85be96743825d93cc57d6f",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodePrivateKeySecretName",
					BadValue: "",
					Detail:   "must provide nodePrivateKeySecretName if peerId is provided",
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
					Network: Devnet,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: Testnet,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: Testnet,
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "missing peerId",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:                  Devnet,
					NodePrivateKeySecretName: "my-private-key",
					PeerId:                   "76c8ca8bb75d1abd853fc17b70cc72cb78a63425fa85be96743825d93cc57d6f",
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:                  Devnet,
					NodePrivateKeySecretName: "my-private-key",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.peerId",
					BadValue: "",
					Detail:   "must provide peerId if nodePrivateKeySecretName is provided",
				},
			},
		},
		{
			Title: "missing nodePrivateKeySecretName",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network:                  Devnet,
					NodePrivateKeySecretName: "my-private-key",
					PeerId:                   "76c8ca8bb75d1abd853fc17b70cc72cb78a63425fa85be96743825d93cc57d6f",
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-node",
				},
				Spec: NodeSpec{
					Network: Devnet,
					PeerId:  "76c8ca8bb75d1abd853fc17b70cc72cb78a63425fa85be96743825d93cc57d6f",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodePrivateKeySecretName",
					BadValue: "",
					Detail:   "must provide nodePrivateKeySecretName if peerId is provided",
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
