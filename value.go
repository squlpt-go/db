package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Stringer interface {
	String() string
}

type Value interface {
	ValidateSQL() error
}

type RawQuery struct {
	Query string
	Args  []any
}

func Raw(query string, args ...any) RawQuery {
	return RawQuery{query, args}
}

func (rq RawQuery) ValidateSQL() error {
	if rq.Query == "" {
		return errors.New("empty Raw SQL values are not allowed")
	}

	return nil
}

// RawQuery implements Transcriber, so it can be used as a query

func (rq RawQuery) Transcribe(_ *sql.DB) (string, []any, error) {
	err := rq.ValidateSQL()
	if err != nil {
		return "", nil, err
	}

	return rq.Query, rq.Args, nil
}

type Ident string

func (i Ident) ValidateSQL() error {
	if i == "" {
		return errors.New("empty SQL Idents are not allowed")
	}

	return nil
}

type String string

func (s String) ValidateSQL() error {
	return nil
}

type Int int

func (i Int) ValidateSQL() error {
	return nil
}

type Bool bool

func (b Bool) ValidateSQL() error {
	return nil
}

type Float float64

func (f Float) ValidateSQL() error {
	return nil
}

type List []Value

func (l List) ValidateSQL() error {
	return nil
}

type Null struct{}

func (n Null) ValidateSQL() error {
	return nil
}

type Time time.Time

func (n Time) ValidateSQL() error {
	return nil
}

func LValue(value any) Value {
	switch val := value.(type) {
	case string:
		return Ident(val)
	case *string:
		return Ident(*val)
	}

	return RValue(value)
}

func RValue(value any) Value {
	switch val := value.(type) {
	case string:
		return String(val)
	case *string:
		return String(*val)
	case []byte:
		return String(val)
	case int:
		return Int(val)
	case *int:
		return Int(*val)
	case int64:
		return Int(val)
	case *int64:
		return Int(*val)
	case int32:
		return Int(val)
	case *int32:
		return Int(*val)
	case int16:
		return Int(val)
	case *int16:
		return Int(*val)
	case int8:
		return Int(val)
	case *int8:
		return Int(*val)
	case uint:
		return Int(val)
	case *uint:
		return Int(*val)
	case uint64:
		return Int(val)
	case *uint64:
		return Int(*val)
	case uint32:
		return Int(val)
	case *uint32:
		return Int(*val)
	case uint16:
		return Int(val)
	case *uint16:
		return Int(*val)
	case uint8:
		return Int(val)
	case *uint8:
		return Int(*val)
	case float64:
		return Float(val)
	case *float64:
		return Float(*val)
	case float32:
		return Float(val)
	case *float32:
		return Float(*val)
	case bool:
		return Bool(val)
	case *bool:
		return Bool(*val)
	case nil:
		return Null{}
	case time.Time:
		return Time(val)
	case Null:
		return val
	case RawQuery:
		return val
	case Ident:
		return val
	case String:
		return val
	case Int:
		return val
	case Float:
		return val
	case List:
		return val
	case Time:
		return val
	case *QueryBuilder:
		return val
	case driver.Valuer:
		v, err := val.Value()
		if err != nil {
			panic(err)
		}
		return RValue(v)
	case *any:
		v := value.(*any)
		return RValue(*v)
	case Stringer:
		return String(val.String())
	default:
		switch reflect.TypeOf(value).Kind() {
		case reflect.String:
			s := fmt.Sprintf("%s", value)
			return String(s)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			s := fmt.Sprintf("%d", value)
			i, _ := strconv.Atoi(s)
			return Int(i)

		case reflect.Float32, reflect.Float64:
			s := fmt.Sprintf("%f", value)
			f, _ := strconv.ParseFloat(s, 64)
			return Float(f)
		}
	}

	panic("Disallowed SQL value type: " + reflect.TypeOf(value).String())
}

func InValue(value any) Value {
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	if t == typeOf[*QueryBuilder]() {
		return value.(*QueryBuilder)
	}

	if t.Kind() == reflect.Slice {
		l := make(List, 0)

		for i := 0; i < v.Len(); i++ {
			l = append(l, RValue(v.Index(i).Interface()))
		}

		return l
	}

	panic("Invalid IN condition type " + t.String())
}

func TableField(table string, field string) Ident {
	return Ident(table + "." + field)
}
