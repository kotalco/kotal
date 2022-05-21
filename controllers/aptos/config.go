package controllers

import (
	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	"gopkg.in/yaml.v2"
)

type Base struct {
	Role string `yaml:"role"`
}
type Config struct {
	Base Base `yaml:"base"`
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *aptosv1alpha1.Node) (config string, err error) {
	var role string
	if node.Spec.Validator {
		role = "validator"
	} else {
		role = "full_node"
	}

	c := Config{
		Base: Base{
			Role: role,
		},
	}

	data, err := yaml.Marshal(&c)
	if err != nil {
		return
	}

	config = string(data)
	return
}
