package test

import (
	"golang-demo/internal/handler"
	"testing"
)

func TestA(t *testing.T) {
	p := handler.Personal{}
	p.Run()
}
