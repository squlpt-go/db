package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
)

var (
	ErrRequired = errors.New("required field")
	ErrMismatch = errors.New("incorrect format")
)

func Require[T IEntity](entity *T, fields ...string) error {
	efs := entityToMap(entity, false, false)
	for key, field := range efs {
		if slices.Contains(fields, key) {
			if field == nil || reflect.ValueOf(field).IsZero() {
				return fmt.Errorf("%w: %s", ErrRequired, key)
			}
		}
	}
	return nil
}

func Match[T IEntity](entity *T, pattern string, fields ...string) error {
	re := regexp.MustCompile(pattern)
	efs := entityToMap(entity, false, false)
	for key, field := range efs {
		if slices.Contains(fields, key) {
			if field != nil && !reflect.ValueOf(field).IsZero() {
				var stringValue string
				if f, ok := field.(driver.Valuer); ok {
					var err error
					field, err = f.Value()
					if err != nil {
						return err
					}
				}
				err := convertAssign(&stringValue, field)
				if err != nil {
					return err
				}
				if !re.MatchString(stringValue) {
					return ErrMismatch
				}
			}
		}
	}
	return nil
}

func doValidateInsert[T IEntity](entity *T) error {
	return doValidate(entity, Insert)
}

func doValidateUpdate[T IEntity](entity *T) error {
	return doValidate(entity, Update)
}

func doValidateChildren[T IEntity](entity *T) error {
	return doValidate(entity, queryTypeChildren)
}

func doFilterInsert[T IEntity](entity *T) (map[string]any, error) {
	return doFilter[T](entity, Insert)
}

func doFilterUpdate[T IEntity](entity *T) (map[string]any, error) {
	return doFilter[T](entity, Update)
}

func doFilterChildren[T IEntity](entity *T) (map[string]any, error) {
	return doFilter[T](entity, queryTypeChildren)
}

const queryTypeChildren = QueryType("CHILDREN")

type IValidate interface {
	Validate() error
}

type IValidateInsert interface {
	ValidateInsert() error
}

type IValidateUpdate interface {
	ValidateUpdate() error
}

type IValidateChildren interface {
	ValidateChildren() error
}

type IFilter interface {
	Filter(map[string]any) error
}

type IFilterInsert interface {
	FilterInsert(map[string]any) error
}

type IFilterUpdate interface {
	FilterUpdate(map[string]any) error
}

type IFilterChildren interface {
	FilterChildren(map[string]any) error
}

func doValidate[T IEntity](entity *T, queryType QueryType) error {
	var err error
	if e, ok := any(entity).(IValidate); ok {
		err = e.Validate()
		if err != nil {
			return err
		}
	}

	switch queryType {
	case queryTypeChildren:
		if e, ok := any(entity).(IValidateChildren); ok {
			err = e.ValidateChildren()
			if err != nil {
				return err
			}
		}
	case Insert:
		if e, ok := any(entity).(IValidateInsert); ok {
			err = e.ValidateInsert()
			if err != nil {
				return err
			}
		}
	case Update:
		if e, ok := any(entity).(IValidateUpdate); ok {
			err = e.ValidateUpdate()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func doFilter[T IEntity](entity *T, queryType QueryType) (map[string]any, error) {
	var flat map[string]any

	switch queryType {
	case queryTypeChildren:
		flat = entityToMap(entity, true, false)
		var err error
		if e, ok := any(entity).(IFilter); ok {
			err = e.Filter(flat)
			if err != nil {
				return flat, err
			}
		}
		if e, ok := any(entity).(IFilterChildren); ok {
			err = e.FilterChildren(flat)
			if err != nil {
				return flat, err
			}
		}
	case Insert:
		flat = entityToMap(entity, false, false)
		var err error
		if e, ok := any(entity).(IFilter); ok {
			err = e.Filter(flat)
			if err != nil {
				return flat, err
			}
		}
		if e, ok := any(entity).(IFilterInsert); ok {
			err = e.FilterInsert(flat)
			if err != nil {
				return flat, err
			}
		}
	case Update:
		flat = entityToMap(entity, true, false)
		var err error
		if e, ok := any(entity).(IFilter); ok {
			err = e.Filter(flat)
			if err != nil {
				return flat, err
			}
		}
		if e, ok := any(entity).(IFilterUpdate); ok {
			err = e.FilterUpdate(flat)
			if err != nil {
				return flat, err
			}
		}
	}

	return flat, nil
}
