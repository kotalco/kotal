package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filecoin node defaulting", func() {
	It("Should default Filecoin mainnet node", func() {
		node := Node{
			Spec: NodeSpec{
				Network: MainNetwork,
			},
		}

		node.Default()

		Expect(node.Spec.Resources.CPU).To((Equal(DefaultMainnetNodeCPURequest)))
		Expect(node.Spec.Resources.CPULimit).To((Equal(DefaultMainnetNodeCPULimit)))
		Expect(node.Spec.Resources.Memory).To((Equal(DefaultMainnetNodeMemoryRequest)))
		Expect(node.Spec.Resources.MemoryLimit).To((Equal(DefaultMainnetNodeMemoryLimit)))
		Expect(node.Spec.Resources.Storage).To((Equal(DefaultMainnetNodeStorageRequest)))

	})

	It("Should default Filecoin nerpa node", func() {
		node := Node{
			Spec: NodeSpec{
				Network: NerpaNetwork,
			},
		}

		node.Default()

		Expect(node.Spec.Resources.CPU).To((Equal(DefaultNerpaNodeCPURequest)))
		Expect(node.Spec.Resources.CPULimit).To((Equal(DefaultNerpaNodeCPULimit)))
		Expect(node.Spec.Resources.Memory).To((Equal(DefaultNerpaNodeMemoryRequest)))
		Expect(node.Spec.Resources.MemoryLimit).To((Equal(DefaultNerpaNodeMemoryLimit)))
		Expect(node.Spec.Resources.Storage).To((Equal(DefaultNerpaNodeStorageRequest)))

	})

	It("Should default Filecoin butterfly node", func() {
		node := Node{
			Spec: NodeSpec{
				Network: ButterflyNetwork,
			},
		}

		node.Default()

		Expect(node.Spec.Resources.CPU).To((Equal(DefaultButterflyNodeCPURequest)))
		Expect(node.Spec.Resources.CPULimit).To((Equal(DefaultButterflyNodeCPULimit)))
		Expect(node.Spec.Resources.Memory).To((Equal(DefaultButterflyNodeMemoryRequest)))
		Expect(node.Spec.Resources.MemoryLimit).To((Equal(DefaultButterflyNodeMemoryLimit)))
		Expect(node.Spec.Resources.Storage).To((Equal(DefaultButterflyNodeStorageRequest)))

	})

	It("Should default Filecoin calibration node", func() {
		node := Node{
			Spec: NodeSpec{
				Network: CalibrationNetwork,
			},
		}

		node.Default()

		Expect(node.Spec.Resources.CPU).To((Equal(DefaultCalibrationNodeCPURequest)))
		Expect(node.Spec.Resources.CPULimit).To((Equal(DefaultCalibrationNodeCPULimit)))
		Expect(node.Spec.Resources.Memory).To((Equal(DefaultCalibrationNodeMemoryRequest)))
		Expect(node.Spec.Resources.MemoryLimit).To((Equal(DefaultCalibrationNodeMemoryLimit)))
		Expect(node.Spec.Resources.Storage).To((Equal(DefaultCalibrationNodeStorageRequest)))

	})

})
