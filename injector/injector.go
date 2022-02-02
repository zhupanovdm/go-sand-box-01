package injector

import (
	"fmt"
	"reflect"
)

type (
	UnitDescriptor struct {
		name       string
		factory    Factory
		objectType reflect.Type
		self       interface{}
	}

	UnitDescriptorList []*UnitDescriptor
)

type Injector struct {
	registry    UnitDescriptorList
	indexByName map[string]*UnitDescriptor
	indexByType map[reflect.Type]UnitDescriptorList
	argsByType  map[reflect.Type]reflect.Value
}

func (r *Injector) SetArgs(args ...interface{}) {
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		r.argsByType[v.Type()] = v
	}
}

func (r *Injector) AddFactory(constructor interface{}, name string) error {
	v := reflect.ValueOf(constructor)

	switch v.Type().Kind() {
	case reflect.Func:
		f := Factory(v)
		if err := f.Validate(); err != nil {
			return fmt.Errorf("factory is not valid: %w", err)
		}

		d, err := r.register(name, f.returnsType())
		if err != nil {
			return fmt.Errorf("failed to register factory: %w", err)
		}
		d.factory = f

		r.registry = append(r.registry, d)
	case reflect.Ptr:
		return r.AddFactory(v.Elem().Interface(), name)
	default:
		return fmt.Errorf("incorrect factory type: %v", v.Type())
	}
	return nil
}

func (r *Injector) GetByName(name string) (interface{}, error) {
	u, exists := r.indexByName[name]
	if !exists {
		return nil, fmt.Errorf("not found by name: %s", name)
	}
	if u.self != nil {
		return u.self, nil
	}

	return nil, nil
}

func (r *Injector) args() []reflect.Value {
	return nil
}

func (r *Injector) register(name string, typ reflect.Type) (*UnitDescriptor, error) {
	if _, exists := r.indexByName[name]; exists {
		return nil, fmt.Errorf("allready registered: %s", name)
	}
	if _, exists := r.indexByType[typ]; !exists {
		r.indexByType[typ] = make(UnitDescriptorList, 0, 1)
	}

	u := &UnitDescriptor{name: name, objectType: typ}
	r.indexByName[name] = u
	r.indexByType[typ] = append(r.indexByType[typ], u)
	return u, nil
}

func New() Injector {
	return Injector{
		registry:    make(UnitDescriptorList, 0),
		indexByName: make(map[string]*UnitDescriptor),
		indexByType: make(map[reflect.Type]UnitDescriptorList),
		argsByType:  make(map[reflect.Type]reflect.Value),
	}
}
