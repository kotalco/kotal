package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filecoin node defaulting", func() {
	It("Should default Filecoin mainnet node", func() {
		node := Node{
			Spec: NodeSpec{
				Network: MainNetwork,
				API:     true,
			},
		}

		node.Default()

		Expect(node.Spec.Image).To((Equal(DefaultLotusImage)))
		Expect(node.Spec.Resources.CPU).To((Equal(DefaultMainnetNodeCPURequest)))
		Expect(node.Spec.Resources.CPULimit).To((Equal(DefaultMainnetNodeCPULimit)))
		Expect(node.Spec.Resources.Memory).To((Equal(DefaultMainnetNodeMemoryRequest)))
		Expect(node.Spec.Resources.MemoryLimit).To((Equal(DefaultMainnetNodeMemoryLimit)))
		Expect(node.Spec.Resources.Storage).To((Equal(DefaultMainnetNodeStorageRequest)))
		Expect(node.Spec.Logging).To(Equal(DefaultLogging))
		Expect(node.Spec.APIPort).To(Equal(DefaultAPIPort))
		Expect(node.Spec.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.Spec.APIRequestTimeout).To(Equal(DefaultAPIRequestTimeout))

	})

	It("Should default Filecoin calibration node", func() {
		node := Node{
			Spec: NodeSpec{
				Network: CalibrationNetwork,
			},
		}

		node.Default()

		Expect(node.Spec.Image).To((Equal(DefaultLotusCalibrationImage)))
		Expect(node.Spec.Resources.CPU).To((Equal(DefaultCalibrationNodeCPURequest)))
		Expect(node.Spec.Resources.CPULimit).To((Equal(DefaultCalibrationNodeCPULimit)))
		Expect(node.Spec.Resources.Memory).To((Equal(DefaultCalibrationNodeMemoryRequest)))
		Expect(node.Spec.Resources.MemoryLimit).To((Equal(DefaultCalibrationNodeMemoryLimit)))
		Expect(node.Spec.Resources.Storage).To((Equal(DefaultCalibrationNodeStorageRequest)))

	})

})
