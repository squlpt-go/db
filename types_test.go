package db

import (
	"fmt"
	"testing"
)

type i interface{}

func TestType(t *testing.T) {
	_ = typeOf[i]()
}

func TestErr(_ *testing.T) {
	fmt.Println(typeOf[error]())
}
