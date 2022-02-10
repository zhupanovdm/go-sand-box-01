package spawner

import (
	"errors"
	"fmt"
	"reflect"

	"sandBox01/injector"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

var _ injector.Provider = (*Spawner)(nil)

type Spawner reflect.Value

func (f Spawner) Validate() error {
	t := f.t()
	if t.Kind() != reflect.Func {
		return fmt.Errorf("factory: must be a function, got: %T", t)
	}
	cnt := t.NumOut()
	if cnt == 0 || cnt > 2 {
		return errors.New("factory: function must return 1 or 2 values")
	}
	r0 := f.Type()
	if r0.Kind() != reflect.Interface {
		return fmt.Errorf("factory: 1st returned value must be an interface, got: %T", r0)
	}
	if cnt > 1 {
		r1 := t.Out(1)
		if !r1.Implements(errorInterface) {
			return fmt.Errorf("factory: 2nd returned value must be error, got: %T", r1)
		}
	}
	return nil
}

func (f Spawner) Type() reflect.Type {
	return f.t().Out(0)
}

func (f Spawner) ReturnsErr() bool {
	return f.t().NumOut() == 2
}

func (f Spawner) args(argProvider injector.ArgProvider) ([]reflect.Value, error) {
	t := f.t()
	values := make([]reflect.Value, 0, t.NumIn())
	for i := 0; i < cap(values); i++ {
		argType := t.In(i)
		if arg, ok := argProvider(argType); ok {
			values = append(values, arg)
			continue
		}
		return nil, fmt.Errorf("arg value for type is not set: %v", argType)
	}
	return values, nil
}

func (f Spawner) Create(argProvider injector.ArgProvider) (*reflect.Value, error) {
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

	object := returned[0]
	if f.ReturnsErr() {
		if returned[1].IsNil() {
			return &object, nil
		}
		return nil, returned[1].Interface().(error)
	}

	return &object, nil
}

func (f Spawner) t() reflect.Type {
	return reflect.Value(f).Type()
}
