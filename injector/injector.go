package injector

import (
	"errors"
	"fmt"
	"reflect"
)

type DependencyList []*DependencyInfo

type DependencyInfo struct {
	name         string
	fqn          string
	args         []reflect.Type
	factory      reflect.Value
	returns      reflect.Type
	returnsError bool
	object       interface{}
}

type Injector struct {
	reg       DependencyList
	nameIndex map[string]*DependencyInfo
	retIndex  map[reflect.Type]DependencyList
}

func (r *Injector) SetArgs(args ...interface{}) {

}

func (r *Injector) AddFactory(constructor interface{}, name string) error {
	t := reflect.TypeOf(constructor)

	d := &DependencyInfo{name: name}

	switch t.Kind() {
	case reflect.Func:
		if t.NumOut() > 2 {
			return errors.New("factory must return 1 or 2 values")
		}

		rt := t.Out(0)
		if rt.Kind() != reflect.Interface {
			return fmt.Errorf("first returned value must be interface, got: %v", t)
		}
		d.returns = rt
		d.fqn = rt.String()
		d.returnsError = t.NumOut() > 1

		d.args = make([]reflect.Type, 0, t.NumIn())
		for i := 0; i < t.NumIn(); i++ {
			d.args = append(d.args, t.In(i))
		}

		r.reg = append(r.reg, d)
	case reflect.Ptr:
		return r.AddFactory(reflect.ValueOf(constructor).Elem().Interface(), name)
	default:
		return fmt.Errorf("incorrect factory type: %v", t)
	}
	return nil
}

func (r *Injector) Get(name string) interface{} {

	return nil
}

func New() Injector {
	return Injector{
		reg: make(DependencyList, 0),
	}
}
