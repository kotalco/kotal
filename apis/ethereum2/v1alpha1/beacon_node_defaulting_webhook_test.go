package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 beacon node defaulting", func() {

	It("Should default beacon node with missing fee recipient and p2p port and logging", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Network: "mainnet",
				Client:  TekuClient,
			},
		}
		node.Default()
		Expect(node.Spec.Image).To(Equal(DefaultTekuBeaconNodeImage))
		Expect(node.Spec.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.Spec.Logging).To(Equal(DefaultLogging))
		Expect(node.Spec.FeeRecipient).To(Equal(shared.EthereumAddress(ZeroAddress)))
	})

	It("Should default beacon node with missing node resources", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Network: "mainnet",
				Client:  LighthouseClient,
			},
		}
		node.Default()
		Expect(node.Spec.Image).To(Equal(DefaultLighthouseBeaconNodeImage))
		Expect(node.Spec.Resources.CPU).To(Equal(DefaultCPURequest))
		Expect(node.Spec.Resources.CPULimit).To(Equal(DefaultCPULimit))
		Expect(node.Spec.Resources.Memory).To(Equal(DefaultMemoryRequest))
		Expect(node.Spec.Resources.MemoryLimit).To(Equal(DefaultMemoryLimit))
		Expect(node.Spec.Resources.Storage).To(Equal(DefaultStorage))
	})

	It("Should default beacon node with missing client and rest port/host", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Network: "mainnet",
				Client:  TekuClient,
				REST:    true,
			},
		}
		node.Default()
		Expect(node.Spec.RESTPort).To(Equal(DefaultRestPort))
		Expect(node.Spec.CORSDomains).To(ConsistOf(DefaultOrigins))
		Expect(node.Spec.Hosts).To(ConsistOf(DefaultOrigins))
	})

	It("Should default beacon node with missing rpc port and host", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Network: "mainnet",
				Client:  NimbusClient,
				RPC:     true,
			},
		}
		node.Default()
		Expect(node.Spec.Image).To(Equal(DefaultNimbusBeaconNodeImage))
		Expect(node.Spec.RPCPort).To(Equal(DefaultRPCPort))
		Expect(node.Spec.CORSDomains).To(ConsistOf(DefaultOrigins))
		Expect(node.Spec.Hosts).To(ConsistOf(DefaultOrigins))
	})

	It("Should default beacon node with missing grpc port", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Client:  PrysmClient,
				Network: "mainnet",
				GRPC:    true,
			},
		}
		node.Default()
		Expect(node.Spec.Image).To(Equal(DefaultPrysmBeaconNodeImage))
		Expect(node.Spec.GRPCPort).To(Equal(DefaultGRPCPort))
		Expect(node.Spec.CORSDomains).To(ConsistOf(DefaultOrigins))
		Expect(node.Spec.Hosts).To(ConsistOf(DefaultOrigins))
	})

	It("Should default beacon node with missing grpc host", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Client:  PrysmClient,
				Network: "mainnet",
				GRPC:    true,
			},
		}
		node.Default()
		Expect(node.Spec.CORSDomains).To(ConsistOf(DefaultOrigins))
		Expect(node.Spec.Hosts).To(ConsistOf(DefaultOrigins))
	})

	It("Should default beacon node with missing cors domains", func() {
		node := BeaconNode{
			Spec: BeaconNodeSpec{
				Client: TekuClient,
				REST:   true,
			},
		}
		node.Default()
		Expect(node.Spec.CORSDomains).To(ConsistOf(DefaultOrigins))
		Expect(node.Spec.Hosts).To(ConsistOf(DefaultOrigins))
	})

})
