package controllers

import (
	"bytes"

	"github.com/BurntSushi/toml"
	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
)

type BurnChain struct {
	Chain string `toml:"chain"`
	Mode  string `toml:"mode"`
}

type Config struct {
	BurnChain BurnChain `toml:"burnchain"`
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *stacksv1alpha1.Node) (config string, err error) {
	c := &Config{}

	c.BurnChain = BurnChain{
		Chain: "bitcoin",
		Mode:  string(node.Spec.Network),
	}

	var buff bytes.Buffer
	enc := toml.NewEncoder(&buff)
	err = enc.Encode(c)
	if err != nil {
		return
	}

	config = buff.String()

	return
}
