package filecoin

import (
	"os"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Lotus Filecoin Client", func() {
	node := filecoinv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "calibration-node",
			Namespace: "filecoin",
		},
		Spec: filecoinv1alpha1.NodeSpec{
			Network: filecoinv1alpha1.CalibrationNetwork,
		},
	}

	client := NewClient(&node)

	It("Should get correct image", func() {
		Expect(client.Image()).To(Equal(DefaultLotusCalibrationImage))
		node.Spec.Network = filecoinv1alpha1.CalibrationNetwork
		Expect(client.Image()).To(Equal(DefaultLotusCalibrationImage))
		node.Spec.Network = filecoinv1alpha1.MainNetwork
		Expect(client.Image()).To(Equal(DefaultLotusImage))
		testImage := "kotalco/lotus:test"
		os.Setenv(EnvLotusImage, testImage)
		Expect(client.Image()).To(Equal(testImage))
	})

	It("Should get correct args", func() {
		Expect(client.Args()).To(ContainElements(
			"lotus",
			"daemon",
		))
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(ContainElements(
			corev1.EnvVar{
				Name:  EnvLotusPath,
				Value: shared.PathData(client.HomeDir()),
			},
			corev1.EnvVar{
				Name:  EnvLogLevel,
				Value: string(node.Spec.Logging),
			},
		))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(BeNil())
	})

	It("Should get image home directory", func() {
		Expect(client.HomeDir()).To(Equal(LotusHomeDir))
	})
})
