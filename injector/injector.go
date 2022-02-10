package injector

import (
	"fmt"
	"reflect"
	"strings"

	"sandBox01/injector/spawner"
	"sandBox01/injector/strategy"
)

const (
	InjectTagKey = "inject"

	DefaultFlag = "default"
	RequireFlag = "require"
)

type (
	Unit struct {
		Provider
		name string
	}

	UnitList []*Unit

	Injector struct {
		indexByType map[reflect.Type]UnitList
		argsByType  map[reflect.Type]reflect.Value
	}
)

func (r *Injector) SetArgs(args ...interface{}) {
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		r.argsByType[v.Type()] = v
	}
}

func (r *Injector) AddFactory(strategy strategy.Strategy, constructor interface{}, name string) error {
	if err := strategy.Validate(); err != nil {
		return fmt.Errorf("injector: unknown injecting strategy %v", strategy)
	}

	v := reflect.ValueOf(constructor)
	switch v.Type().Kind() {
	case reflect.Func:
		spwnr := spawner.Spawner(v)
		if err := spwnr.Validate(); err != nil {
			return fmt.Errorf("injector: factory is not valid: %w", err)
		}

		prvdr := strategy.ApplyTo(spwnr)
		d, err := r.register(name, prvdr)
		if err != nil {
			return fmt.Errorf("injector: failed to register factory: %w", err)
		}
		d.Provider = prvdr

	case reflect.Ptr:
		return r.AddFactory(strategy, v.Elem().Interface(), name)

	default:
		return fmt.Errorf("injector: incorrect factory type: %v", v.Type())

	}
	return nil
}

func (r *Injector) ObjectByTypeName(t reflect.Type, name string) (interface{}, error) {
	if u := r.indexByType[t].FindFirstByName(name); u != nil {
		return r.object(u)
	}
	return nil, nil
}

func (r *Injector) ObjectByType(t reflect.Type) (interface{}, error) {
	if list, exists := r.indexByType[t]; exists {
		if len(list) > 1 {
			return nil, fmt.Errorf("injector: ambigous request by type %v: %d items found", t, len(list))
		}
		return r.object(list[0])
	}
	return nil, nil
}

func (r *Injector) object(u *Unit) (interface{}, error) {
	object, err := u.Create(func(argType reflect.Type) (arg reflect.Value, ok bool) { arg, ok = r.argsByType[argType]; return })
	if err != nil {
		return nil, err
	}
	if object == nil {
		return nil, nil
	}

	v := reflect.ValueOf(*object)
	e := v.Elem()
	t := e.Type()
	for i := 0; i < t.NumField(); i++ {
		name := DefaultFlag
		tagFlags := make(map[string]bool)

		field := t.Field(i)
		if tagValue, ok := field.Tag.Lookup(InjectTagKey); ok {
			tagValues := strings.Split(tagValue, ",")
			for j, tag := range tagValues {
				switch tag {
				case DefaultFlag:
					if j == 0 {
						name = DefaultFlag
						tagFlags[DefaultFlag] = true
					}
				case RequireFlag:
					tagFlags[RequireFlag] = true
				default:
					name = tag
				}
			}
		}

		var dep interface{}
		if tagFlags[DefaultFlag] {
			if dep, err = r.ObjectByType(field.Type); err != nil {
				return nil, fmt.Errorf("injector: error while constructing dependency %v: %w", field.Type, err)
			}
		} else {
			if dep, err = r.ObjectByTypeName(field.Type, name); err != nil {
				return nil, fmt.Errorf("injector: error while constructing dependency %v&%s: %w", field.Type, name, err)
			}
		}
		if dep == nil && tagFlags[RequireFlag] {
			return nil, fmt.Errorf("injector: missed required dependency %v&%s", field.Type, name)
		}
		if dep != nil {
			if !v.Field(i).CanAddr() {
				return nil, fmt.Errorf("injector: cannot set field value: %s", field.Name)
			}
			v.Field(i).Set(reflect.ValueOf(dep).Elem().Addr())
		}
	}
	return v.Interface(), nil
}

func (r *Injector) args() []reflect.Value {
	return nil
}

func (r *Injector) register(name string, provider Provider) (*Unit, error) {
	t := provider.Type()
	if r.indexByType[t].FindFirstByName(name) != nil {
		return nil, fmt.Errorf("injector: allready registered by name: %s", name)
	}
	if _, exists := r.indexByType[t]; !exists {
		r.indexByType[t] = make(UnitList, 0, 1)
	}
	u := &Unit{name: name, Provider: provider}
	r.indexByType[t] = append(r.indexByType[t], u)
	return u, nil
}

func (l UnitList) FindFirstByName(name string) *Unit {
	for _, u := range l {
		if u.name == name {
			return u
		}
	}
	return nil
}

func New() Injector {
	return Injector{
		indexByType: make(map[reflect.Type]UnitList),
		argsByType:  make(map[reflect.Type]reflect.Value),
	}
}
