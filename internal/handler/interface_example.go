package handler

import "fmt"

type Animal interface {
	Run()
	Eat()
}
type Person1 interface {
	GetAge() uint
}
type Personal struct {
}

func (p *Personal) Run() {
	fmt.Printf("跑")
}

func (p *Personal) Eat() {
	fmt.Printf("吃")
}

func (p *Personal) GetAge() uint {
	return 18
}
