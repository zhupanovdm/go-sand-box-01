package strategy

import (
	"fmt"

	"sandBox01/injector"
	"sandBox01/injector/spawner"
)

const (
	SingletonType Strategy = "singleton"
	FactoryType   Strategy = "factory"
)

type Strategy string

func (s Strategy) String() string {
	return string(s)
}

func (s Strategy) Validate() error {
	if s != SingletonType && s != FactoryType {
		return fmt.Errorf("strategy: unknown strategy type: %v", s)
	}
	return nil
}

func (s Strategy) ApplyTo(spawner spawner.Spawner) injector.Provider {
	switch s {
	case SingletonType:
		return Singleton(spawner)
	case FactoryType:
		return Factory(spawner)
	}
	return nil
}
