package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 validator client defaulting", func() {

	It("Should default validator client with missing client and graffiti", func() {
		node := Validator{
			Spec: ValidatorSpec{
				Network: "mainnet",
			},
		}
		node.Default()
		Expect(node.Spec.Client).To(Equal(DefaultClient))
		Expect(node.Spec.Graffiti).To(Equal(DefaultGraffiti))
	})

})
