package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 validator client defaulting", func() {

	It("Should default validator client with missing client, graffiti, and resources", func() {
		node := Validator{
			Spec: ValidatorSpec{
				Network: "mainnet",
				Client:  TekuClient,
			},
		}
		node.Default()
		Expect(node.Spec.Image).To(Equal(DefaultTekuValidatorImage))
		Expect(node.Spec.Graffiti).To(Equal(DefaultGraffiti))
		Expect(node.Spec.FeeRecipient).To(Equal(shared.EthereumAddress(ZeroAddress)))
		Expect(node.Spec.Logging).To(Equal(DefaultLogging))
		Expect(node.Spec.Resources.CPU).To(Equal(DefaultCPURequest))
		Expect(node.Spec.Resources.CPULimit).To(Equal(DefaultCPULimit))
		Expect(node.Spec.Resources.Memory).To(Equal(DefaultMemoryRequest))
		Expect(node.Spec.Resources.MemoryLimit).To(Equal(DefaultMemoryLimit))
		Expect(node.Spec.Resources.Storage).To(Equal(DefaultStorage))
	})

})
