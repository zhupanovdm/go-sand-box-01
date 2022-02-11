package main

import (
	"fmt"
	"log"
	"reflect"
	"sandBox01/injector"
)

type Config struct {
	attr string
}

var _ ObjectIFace = (*Object1)(nil)
var _ ObjectIFace = (*Object2)(nil)

type ObjectIFace interface {
}

type Object1 struct {
	Dep2 ObjectIFace `inject:"object2,require"`
	attr string
}

type Object2 struct {
}

func NewObject1(cfg *Config) (ObjectIFace, error) {
	return &Object1{attr: cfg.attr}, nil
}

func NewObject2() (ObjectIFace, error) {
	return &Object2{}, nil
}

func main() {
	i := injector.New()
	i.AddConfig(&Config{"fddgg"})
	if err := i.RegisterFactory("object1", NewObject1, injector.Singleton); err != nil {
		log.Fatal(err)
	}
	if err := i.RegisterFactory("object2", NewObject2, injector.Singleton); err != nil {
		log.Fatal(err)
	}

	var t = reflect.TypeOf((*ObjectIFace)(nil)).Elem()
	o1, err := i.ObjectByTypeName(t, "object1")
	fmt.Println(o1, err)
}
