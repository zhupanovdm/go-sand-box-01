package injector

import (
	"errors"
	"fmt"
	"reflect"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

type (
	Factory     reflect.Value
	ArgProvider func(argType reflect.Type) (reflect.Value, bool)
)

func (f Factory) Validate() error {
	t := f.t()
	if t.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function, got: %v", t)
	}
	if t.NumOut() < 1 || t.NumOut() > 2 {
		return errors.New("factory must return 1 or 2 values")
	}
	if t.Out(0).Kind() != reflect.Interface {
		return fmt.Errorf("first of returned value must be interface, got: %v", t.Out(0))
	}
	if t.NumOut() == 2 && !t.Out(1).Implements(errorInterface) {
		return fmt.Errorf("second of returned value must be error, got: %v", t.Out(1))
	}
	return nil
}

func (f Factory) ReturnsType() reflect.Type {
	return f.t().Out(0)
}

func (f Factory) Spawn(argProvider ArgProvider) (interface{}, error) {
	factory := reflect.Value(f)
	t := factory.Type()
	num := t.NumIn()
	args := make([]reflect.Value, 0, num)
	for i := 0; i < num; i++ {
		argType := t.In(i)
		if arg, ok := argProvider(argType); ok {
			args = append(args, arg)
			continue
		}
		return nil, fmt.Errorf("arg value for type is not set: %v", argType)
	}

	returned := factory.Call(args)

	object := returned[0].Interface()
	if t.NumOut() == 2 {
		if returned[1].IsNil() {
			return object, nil
		}
		return nil, returned[1].Interface().(error)
	}

	return object, nil
}

func (f Factory) t() reflect.Type {
	return reflect.Value(f).Type()
}
