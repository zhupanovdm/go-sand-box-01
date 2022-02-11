package injector

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	InjectTagKey = "inject"

	DefaultTagFlag = "default"
	RequireTagFlag = "require"
)

var _ Injector = (*registry)(nil)

type (
	registry struct {
		sync.RWMutex
		unitByName map[string]*unit
		unitByType map[reflect.Type]unitList

		configsByType map[reflect.Type]*reflect.Value
	}

	unit struct {
		sync.Mutex
		ObjectProvider
		name string
	}

	unitList []*unit
)

func (u *unit) String() string {
	if u == nil {
		return "<nil>"
	}
	return u.name
}

func (l unitList) String() string {
	var b strings.Builder
	for i, u := range l {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(u.String())
	}
	return b.String()
}

func (r *registry) AddConfig(configs ...interface{}) {
	for _, cfg := range configs {
		value := reflect.ValueOf(cfg)
		r.configsByType[value.Type()] = &value
	}
}

func (r *registry) RegisterFactory(name string, factoryFunc interface{}, strategy Strategy) error {
	if err := strategy.Validate(); err != nil {
		return fmt.Errorf("injector: unknown injecting strategy %v", strategy)
	}
	spawner, err := SpawnerFromFactoryFunc(factoryFunc)
	if err != nil {
		return fmt.Errorf("injector: can't convert function to spawner: %w", err)
	}
	if err := r.register(name, strategy.ApplyTo(spawner)); err != nil {
		return fmt.Errorf("injector: failed to register factory: %w", err)
	}
	return nil
}

func (r *registry) ObjectByTypeName(t reflect.Type, name string) (interface{}, error) {
	if u := r.unitByType[t].FindFirstByName(name); u != nil {
		return r.object(u, make(unitList, 0))
	}
	return nil, nil
}

func (r *registry) ObjectByType(t reflect.Type) (interface{}, error) {
	if list, exists := r.unitByType[t]; exists {
		if len(list) > 1 {
			return nil, fmt.Errorf("registry: ambigous request by type %v: %d items found", t, len(list))
		}
		return r.object(list[0], make(unitList, 0))
	}
	return nil, nil
}

func (r *registry) config(typ reflect.Type) *reflect.Value {
	return r.configsByType[typ]
}

func (r *registry) object(u *unit, cyclicCheck unitList) (interface{}, error) {
	for i, dependent := range cyclicCheck {
		if u == dependent {
			path := cyclicCheck[:i+1]
			return nil, fmt.Errorf("injector: cyclic dependency detected: %v", path)
		}
	}
	cyclicCheck = append(cyclicCheck, u)

	object, err := u.Get(r.config)
	if err != nil {
		return nil, err
	}
	if object == nil {
		return nil, nil
	}

	e := object.Elem()
	t := e.Type()
	for i := 0; i < t.NumField(); i++ {
		name := DefaultTagFlag
		tagFlags := make(map[string]bool)

		field := t.Field(i)
		if tagValue, ok := field.Tag.Lookup(InjectTagKey); ok {
			tagValues := strings.Split(tagValue, ",")
			for j, tag := range tagValues {
				switch tag {
				case DefaultTagFlag:
					if j == 0 {
						name = DefaultTagFlag
						tagFlags[DefaultTagFlag] = true
					}
				case RequireTagFlag:
					tagFlags[RequireTagFlag] = true
				default:
					name = tag
				}
			}
		}

		var dep interface{}
		if tagFlags[DefaultTagFlag] {
			if dep, err = r.ObjectByType(field.Type); err != nil {
				return nil, fmt.Errorf("registry: error while constructing dependency %v: %w", field.Type, err)
			}
		} else {
			if dep, err = r.ObjectByTypeName(field.Type, name); err != nil {
				return nil, fmt.Errorf("registry: error while constructing dependency %v&%s: %w", field.Type, name, err)
			}
		}
		if dep == nil && tagFlags[RequireTagFlag] {
			return nil, fmt.Errorf("registry: missed required dependency %v&%s", field.Type, name)
		}
		if dep != nil {
			if !object.Field(i).CanAddr() {
				return nil, fmt.Errorf("registry: cannot set field value: %s", field.Name)
			}
			object.Field(i).Set(reflect.ValueOf(dep).Elem().Addr())
		}
	}
	return object.Interface(), nil
}

func (r *registry) byName(name string) *unit {
	r.RLock()
	defer r.RUnlock()
	return r.unitByName[name]
}

func (r *registry) register(name string, provider ObjectProvider) error {
	r.Lock()
	defer r.Unlock()

	t := provider.Type()
	if _, exists := r.unitByName[name]; exists {
		return fmt.Errorf("injector: already registered by name: %s", name)
	}
	if _, exists := r.unitByType[t]; !exists {
		r.unitByType[t] = make(unitList, 0, 1)
	}
	u := &unit{name: name, ObjectProvider: provider}
	r.unitByName[name] = u
	r.unitByType[t] = append(r.unitByType[t], u)
	return nil
}

func (l unitList) FindFirstByName(name string) *unit {
	for _, u := range l {
		if u.name == name {
			return u
		}
	}
	return nil
}

func New() Injector {
	return &registry{
		unitByName:    make(map[string]*unit),
		unitByType:    make(map[reflect.Type]unitList),
		configsByType: make(map[reflect.Type]*reflect.Value),
	}
}
