package main

import (
	"errors"
	"service/supervisor"

	"github.com/brimstone/logger"
)

var log = logger.New()

func main() {

	s := supervisor.New()

	if err := s.Add("foo", &foo{}); err != nil {
		log.Println("foo failed and this is ok")
	}
	if err := s.Add("bar", &bar{}); err != nil {
		panic(err)
	}

	s.Run()

}

type foo struct {
}

func (f *foo) Init() error {
	log.Println("foo init")
	return errors.New("Failed")
}

func (f *foo) Run() error {
	log.Println("foo Run")
	return nil
}

type bar struct {
}

func (b *bar) Init() error {
	log.Println("bar Init()")
	return nil
}

func (b *bar) Run() error {
	log.Println("bar Run()")
	return errors.New("bar fails to run")
}
