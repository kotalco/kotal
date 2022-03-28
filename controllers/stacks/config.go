package controllers

import (
	"bytes"

	"github.com/BurntSushi/toml"
	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	stacksClients "github.com/kotalco/kotal/clients/stacks"
	"github.com/kotalco/kotal/controllers/shared"
)

type BurnChain struct {
	Chain string `toml:"chain"`
	Mode  string `toml:"mode"`
}

type Node struct {
	WorkingDir string `toml:"working_dir"`
}

type Config struct {
	Node      Node      `toml:"node"`
	BurnChain BurnChain `toml:"burnchain"`
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *stacksv1alpha1.Node) (config string, err error) {
	c := &Config{}

	c.Node = Node{
		WorkingDir: shared.PathData(stacksClients.StacksNodeHomeDir),
	}

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
