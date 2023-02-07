package filecoin

import (
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
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

	It("Should get correct args", func() {
		Expect(client.Args()).To(BeNil())
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
		Expect(client.Command()).To(ContainElements(
			"lotus",
			"daemon",
		))
	})

	It("Should get image home directory", func() {
		Expect(client.HomeDir()).To(Equal(LotusHomeDir))
	})
})
