package shared

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetSecret returns k8s secret stored at key
func GetSecret(ctx context.Context, client client.Client, name types.NamespacedName, key string) (value string, err error) {
	secret := &corev1.Secret{}

	if err = client.Get(ctx, name, secret); err != nil {
		return
	}

	value = string(secret.Data[key])

	return
}
