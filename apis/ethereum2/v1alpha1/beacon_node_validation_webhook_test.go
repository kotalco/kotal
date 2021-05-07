package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo"
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
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
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
		{
			Title: "Node #3",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Join:   "mainnet",
					Client: TekuClient,
					RPC:    true,
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
					Join:   "mainnet",
					Client: LighthouseClient,
					RPC:    true,
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
					Join:   "mainnet",
					Client: PrysmClient,
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
					Join:   "mainnet",
					Client: TekuClient,
					GRPC:   true,
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
			Title: "Node #7",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Join:   "pyrmont",
					Client: PrysmClient,
					RPC:    true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.eth1Endpoints",
					BadValue: "",
					Detail:   "required by prysm client if network is not mainnet",
				},
			},
		},
		{
			Title: "Node #8",
			Node: &BeaconNode{
				Spec: BeaconNodeSpec{
					Join:   "mainnet",
					Client: NimbusClient,
					Eth1Endpoints: []string{
						"http://localhost:8545",
						"http://localhost:8546",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.eth1Endpoints",
					BadValue: "http://localhost:8545, http://localhost:8546",
					Detail:   "multiple Ethereum 1 endpoints not supported by nimbus client",
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
					Join: "mainnet",
				},
			},
			NewNode: &BeaconNode{
				Spec: BeaconNodeSpec{
					Join: "pyrmont",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.join",
					BadValue: "pyrmont",
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
