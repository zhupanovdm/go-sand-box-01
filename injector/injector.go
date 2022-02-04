package injector

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	InjectTagKey = "inject"

	DefaultFlag = "default"
	RequireFlag = "require"
)

type (
	UnitDescriptor struct {
		name       string
		factory    Factory
		objectType reflect.Type
		self       interface{}
	}

	UnitDescriptorList []*UnitDescriptor

	Injector struct {
		registry    UnitDescriptorList
		indexByType map[reflect.Type]UnitDescriptorList
		argsByType  map[reflect.Type]reflect.Value
	}
)

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
			return fmt.Errorf("injector: factory is not valid: %w", err)
		}

		d, err := r.register(name, f.ReturnsType())
		if err != nil {
			return fmt.Errorf("injector: failed to register factory: %w", err)
		}
		d.factory = f

		r.registry = append(r.registry, d)
	case reflect.Ptr:
		return r.AddFactory(v.Elem().Interface(), name)
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

func (r *Injector) object(u *UnitDescriptor) (interface{}, error) {
	if u.self != nil {
		return u.self, nil
	}
	object, err := u.factory.Spawn(func(argType reflect.Type) (reflect.Value, bool) {
		arg, ok := r.argsByType[argType]
		return arg, ok
	})
	if err != nil {
		return nil, err
	}
	if object != nil {
		v := reflect.ValueOf(object).Elem()
		t := v.Type()
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
		u.self = object
	}
	return object, nil
}

func (r *Injector) args() []reflect.Value {
	return nil
}

func (r *Injector) register(name string, typ reflect.Type) (*UnitDescriptor, error) {
	if r.indexByType[typ].FindFirstByName(name) != nil {
		return nil, fmt.Errorf("injector: allready registered by name: %s", name)
	}
	if _, exists := r.indexByType[typ]; !exists {
		r.indexByType[typ] = make(UnitDescriptorList, 0, 1)
	}
	u := &UnitDescriptor{name: name, objectType: typ}
	r.indexByType[typ] = append(r.indexByType[typ], u)
	return u, nil
}

func (l UnitDescriptorList) FindFirstByName(name string) *UnitDescriptor {
	for _, u := range l {
		if u.name == name {
			return u
		}
	}
	return nil
}

func New() Injector {
	return Injector{
		registry:    make(UnitDescriptorList, 0),
		indexByType: make(map[reflect.Type]UnitDescriptorList),
		argsByType:  make(map[reflect.Type]reflect.Value),
	}
}
