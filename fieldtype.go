package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
)

type FieldType interface {
	sql.Scanner
	driver.Valuer
}

type Nullable[T any] struct {
	Wrapped T
	Valid   bool // Valid is true if Wrapped is not NULL
}

func (ns *Nullable[T]) Scan(value any) error {
	if value == nil {
		ns.Valid = false
		ns.Wrapped = *new(T) // zero value of T
		return nil
	}
	ns.Valid = true
	return convertAssign(&ns.Wrapped, value)
}

func (ns Nullable[T]) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Wrapped, nil
}

func NewNullable[T any](v T) Nullable[T] {
	return Nullable[T]{
		Valid:   true,
		Wrapped: v,
	}
}

type Json[T any] struct {
	Encoded T
}

func (j *Json[T]) Scan(value any) error {
	if value == nil {
		var v T
		j.Encoded = v
		return nil
	}

	if s, ok := value.([]byte); ok {
		err := json.Unmarshal(s, &j.Encoded)

		if err != nil {
			return err
		}
	}

	if s, ok := value.(T); ok {
		j.Encoded = s
	}

	return nil
}

func (j Json[T]) Value() (driver.Value, error) {
	if reflect.ValueOf(j.Encoded).IsNil() {
		return nil, nil
	}

	return json.Marshal(j.Encoded)
}

func (j Json[T]) Unserialize() any {
	return j.Encoded
}

type Unserializeable interface {
	Unserialize() any
}

func NewJson[T any](v T) Json[T] {
	return Json[T]{Encoded: v}
}

type Regexp struct {
	*regexp.Regexp
}

func (j *Regexp) Scan(value any) error {
	if value == nil {
		j.Regexp = nil
		return nil
	}

	var err error
	var pattern *regexp.Regexp

	switch s := value.(type) {
	case []byte:
		pattern, err = regexp.Compile(string(s))
	case string:
		pattern, err = regexp.Compile(string(s))
	default:
		return errors.New("invalid regex scan type")
	}

	if err != nil {
		return err
	}

	j.Regexp = pattern
	return nil
}

func (j Regexp) Value() (driver.Value, error) {
	if reflect.ValueOf(j.Regexp).IsNil() {
		return nil, nil
	}

	return j.Regexp, nil
}

func NewRegexp(pattern string) Regexp {
	return Regexp{Regexp: regexp.MustCompile(pattern)}
}
