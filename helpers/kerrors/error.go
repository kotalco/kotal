package kerrors

import "k8s.io/apimachinery/pkg/util/validation/field"

type KErrors struct {
	FieldErr   field.Error
	ChildField string
	CustomMsg  string
}

func New(err field.Error) *KErrors {
	return &KErrors{
		FieldErr: err,
	}
}
