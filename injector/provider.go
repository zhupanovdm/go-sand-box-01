package injector

import "reflect"

type (
	Provider interface {
		Create(argProvider ArgProvider) (*reflect.Value, error)
		Type() reflect.Type
	}

	ArgProvider func(argType reflect.Type) (reflect.Value, bool)
)
