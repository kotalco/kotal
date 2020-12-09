package v1alpha1

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum defaulting", func() {
	testCases := []struct {
		Name     string
		Node     Node
		Defaults map[string]string
	}{
		{
			Name: "missing client",
			Node: Node{
				Spec: NodeSpec{
					Join: "mainnet",
				},
			},
			Defaults: map[string]string{
				"Client": "teku",
			},
		},
	}

	for i, testCase := range testCases {
		func() {
			tc := testCase
			It(fmt.Sprintf("Should default node #%d with %s", i, tc.Name), func() {
				tc.Node.Default()
				r := reflect.ValueOf(tc.Node.Spec)
				for k, v := range tc.Defaults {
					Expect(r.FieldByName(k).String()).To(Equal(v))
				}
			})
		}()
	}
})
