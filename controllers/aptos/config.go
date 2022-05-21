package controllers

import (
	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	"gopkg.in/yaml.v2"
)

type Waypoint struct {
	FromConfig string `yaml:"from_config"`
}

type Base struct {
	Role     string   `yaml:"role"`
	Waypoint Waypoint `yaml:"waypoint"`
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
			Waypoint: Waypoint{
				FromConfig: node.Spec.Waypoint,
			},
		},
	}

	data, err := yaml.Marshal(&c)
	if err != nil {
		return
	}

	config = string(data)
	return
}
