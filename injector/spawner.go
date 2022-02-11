package injector

import (
	"errors"
	"fmt"
	"reflect"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

var _ ObjectProvider = (*Spawner)(nil)

type Spawner reflect.Value

func (s *Spawner) Validate() error {
	t := s.t()
	if t.Kind() != reflect.Func {
		return fmt.Errorf("spawner: must be a function, got: %T", t)
	}
	cnt := t.NumOut()
	if cnt == 0 || cnt > 2 {
		return errors.New("spawner: function must return 1 or 2 values")
	}
	r0 := s.Type()
	if r0.Kind() != reflect.Interface {
		return fmt.Errorf("spawner: 1st returned value must be an interface, got: %T", r0)
	}
	if cnt > 1 {
		r1 := t.Out(1)
		if !r1.Implements(errorInterface) {
			return fmt.Errorf("spawner: 2nd returned value must be error, got: %T", r1)
		}
	}
	return nil
}

func (s *Spawner) Type() reflect.Type {
	return s.t().Out(0)
}

func (s *Spawner) ReturnsErr() bool {
	return s.t().NumOut() == 2
}

func (s *Spawner) Get(cfgProvider CfgProvider) (*reflect.Value, error) {
	config, err := s.cfg(cfgProvider)
	if err != nil {
		return nil, err
	}
	result := reflect.Value(*s).Call(config)
	if s.ReturnsErr() && !result[1].IsNil() {
		return nil, result[1].Interface().(error)
	}
	object := result[0]
	if object.IsNil() {
		return nil, nil
	}
	return &object, nil
}

func (s *Spawner) cfg(cfgProvider CfgProvider) ([]reflect.Value, error) {
	t := s.t()
	values := make([]reflect.Value, 0, t.NumIn())
	for i := 0; i < cap(values); i++ {
		argType := t.In(i)
		if arg := cfgProvider(argType); arg != nil {
			values = append(values, *arg)
			continue
		}
		return nil, fmt.Errorf("spawner: config is not set: %v", argType)
	}
	return values, nil
}

func (s *Spawner) t() reflect.Type {
	return reflect.Value(*s).Type()
}

func SpawnerFromFactoryFunc(factoryFunc interface{}) (*Spawner, error) {
	value := reflect.ValueOf(factoryFunc)
	switch value.Type().Kind() {
	case reflect.Func:
		s := Spawner(value)
		if err := s.Validate(); err != nil {
			return nil, err
		}
		return &s, nil
	case reflect.Ptr:
		return SpawnerFromFactoryFunc(value.Elem().Interface())
	}
	return nil, fmt.Errorf("spawner: incorrect factory function: %v", value.Type())
}
