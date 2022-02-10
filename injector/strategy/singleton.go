package strategy

import (
	"reflect"
	"sync"

	"sandBox01/injector"
	"sandBox01/injector/spawner"
)

var _ injector.Provider = (*singleton)(nil)

type singleton struct {
	sync.Once
	spawner.Spawner

	object *reflect.Value
	err    error
}

func (s *singleton) Create(argProvider injector.ArgProvider) (*reflect.Value, error) {
	s.Do(func() { s.object, s.err = s.Spawner.Create(argProvider) })
	return s.object, s.err
}

func Singleton(spawner spawner.Spawner) injector.Provider {
	return &singleton{Spawner: spawner}
}
