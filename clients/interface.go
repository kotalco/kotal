package clients

import corev1 "k8s.io/api/core/v1"

// Interface is client interface
type Interface interface {
	Args() []string
	Command() []string
	Env() []corev1.EnvVar
	HomeDir() string
}
