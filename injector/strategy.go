package injector

import (
	"fmt"
	"reflect"
	"sync"
)

const (
	Singleton Strategy = "singleton"
	Factory   Strategy = "factory"
)

type Strategy string

func (s Strategy) String() string {
	return string(s)
}

func (s Strategy) Validate() error {
	if s != Singleton && s != Factory {
		return fmt.Errorf("strategy: unknown strategy type: %v", s)
	}
	return nil
}

func (s Strategy) ApplyTo(provider ObjectProvider) ObjectProvider {
	switch s {
	case Singleton:
		return SingletonProvider(provider)
	case Factory:
		return provider
	}
	return nil
}

var _ ObjectProvider = (*singletonProvider)(nil)

type singletonProvider struct {
	sync.Once
	ObjectProvider

	object *reflect.Value
	err    error
}

func (s *singletonProvider) Get(argProvider CfgProvider) (*reflect.Value, error) {
	s.Do(func() { s.object, s.err = s.ObjectProvider.Get(argProvider) })
	return s.object, s.err
}

func SingletonProvider(provider ObjectProvider) ObjectProvider {
	return &singletonProvider{ObjectProvider: provider}
}
