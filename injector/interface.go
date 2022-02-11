package injector

import "reflect"

type (
	Injector interface {
		AddConfig(configs ...interface{})
		RegisterFactory(name string, factoryFunc interface{}, strategy Strategy) error

		ObjectByTypeName(typ reflect.Type, name string) (interface{}, error)
		ObjectByType(typ reflect.Type) (interface{}, error)
	}

	ObjectProvider interface {
		Get(cfgProvider CfgProvider) (*reflect.Value, error)
		Type() reflect.Type
	}

	CfgProvider func(cfgType reflect.Type) *reflect.Value
)
