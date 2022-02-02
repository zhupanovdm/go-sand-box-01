package main

import (
	"log"
	"sandBox01/injector"
)

type Config struct{}

var _ ObjectIFace = (*Object1)(nil)
var _ ObjectIFace = (*Object2)(nil)

type ObjectIFace interface {
}

type Object1 struct {
	dep2 ObjectIFace
}

type Object2 struct {
}

func NewObject1(cfg *Config) ObjectIFace {
	return Object1{}
}

func main() {
	i := injector.New()
	i.SetArgs(&Config{})
	if err := i.AddFactory(NewObject1, "object1"); err != nil {
		log.Fatal(err)
	}
	i.GetByName("dep")
}
