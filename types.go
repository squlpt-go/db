package db

import "reflect"

type typeWrapper[T any] struct {
	_type T
}

func typeOf[T any]() reflect.Type {
	tid := typeWrapper[T]{}
	iType, _ := reflect.TypeOf(tid).FieldByName("_type")
	return iType.Type
}

type typeId string

func tId[T any]() typeId {
	return typeId(typeOf[T]().String())
}
