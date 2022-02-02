package injector

import (
	"errors"
	"fmt"
	"reflect"
)

type Factory reflect.Value

func (f Factory) Validate() error {
	t := f.t()
	if t.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function, got: %v", t)
	}
	if t.NumOut() < 1 || t.NumOut() > 2 {
		return errors.New("factory must return 1 or 2 values")
	}

	if t.Out(0).Kind() != reflect.Interface {
		return fmt.Errorf("first of returned value must be interface, got: %v", t)
	}

	return nil
}

func (f Factory) returnsError() bool {
	return f.t().NumOut() == 2
}

func (f Factory) returnsType() reflect.Type {
	return f.t().Out(0)
}

func (f Factory) t() reflect.Type {
	return reflect.Value(f).Type()
}
