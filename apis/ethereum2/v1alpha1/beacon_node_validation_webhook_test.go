package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Ethereum 2.0 beacon node validation", func() {

	createCases := []struct {
		Title  string
		Node   *BeaconNode
		Errors field.ErrorList
	}{
		{
			Title: "Node #1",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  PrysmClient,
					REST:    true,
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
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  PrysmClient,
					REST:    true,
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
			Title: "Node #3",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  TekuClient,
					RPC:     true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: true,
					Detail:   "not supported by teku client",
				},
			},
		},
		{
			Title: "Node #4",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  LighthouseClient,
					RPC:     true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: true,
					Detail:   "not supported by lighthouse client",
				},
			},
		},
		{
			Title: "Node #5",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  PrysmClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: false,
					Detail:   "can't be disabled in prysm client",
				},
			},
		},
		{
			Title: "Node #6",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  TekuClient,
					GRPC:    true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.grpc",
					BadValue: true,
					Detail:   "not supported by teku client",
				},
			},
		},
		{
			Title: "Node #9",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network:        "mainnet",
					Client:         TekuClient,
					CertSecretName: "my-cert",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.certSecretName",
					BadValue: "my-cert",
					Detail:   "not supported by teku client",
				},
			},
		},
	}

	updateCases := []struct {
		Title   string
		OldNode *BeaconNode
		NewNode *BeaconNode
		Errors  field.ErrorList
	}{
		{
			Title: "Node #1",
			OldNode: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  TekuClient,
				},
			},
			NewNode: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "goerli",
					Client:  TekuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: "goerli",
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "Node #2",
			OldNode: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  TekuClient,
				},
			},
			NewNode: &BeaconNode{
				Spec: BeaconNodeSpec{
					Network: "mainnet",
					Client:  PrysmClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: PrysmClient,
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating beacon node", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Node.Default()
					err := cc.Node.ValidateCreate()
					Expect(err).ToNot(BeNil())

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

	Context("While updating beacon node", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.NewNode.Default()
					cc.OldNode.Default()
					err := cc.NewNode.ValidateUpdate(cc.OldNode)
					Expect(err).ToNot(BeNil())

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

})
