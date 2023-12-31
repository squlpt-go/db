package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
)

func FromRows[T any](rows *sql.Rows) (T, error) {
	e := new(T)

	entityField := reflect.ValueOf(e).Elem().FieldByName("Entity")

	if entityField.IsValid() {
		if entityField.Type() != reflect.TypeOf(&Entity{}) {
			return *e, errors.New("Invalid embedded Entity type in " + reflect.ValueOf(e).Elem().Type().String() + ": must be an embedded *Entity")
		}

		ref, err := entityFromRows(rows)

		if err != nil {
			return *e, err
		}

		entityField.Set(reflect.ValueOf(ref))
	}

	columnTypes, err := rows.ColumnTypes()
	var scan = make([]any, 0)

	if err != nil {
		return *e, err
	}

	for _, columnType := range columnTypes {
		fieldRef, found := getFieldRefByColumnType(e, columnType, true, true)

		if found {
			scan = append(scan, fieldRef)
		} else {
			scan = append(scan, new(any))
		}
	}

	err = rows.Scan(scan...)

	if err != nil {
		return *e, fmt.Errorf("failed to execute FromRows: %w", err)
	}

	if h, ok := any(e).(Hydratable); ok {
		m := make(map[string]any)
		for i, columnType := range columnTypes {
			m[columnType.Name()] = reflect.ValueOf(scan[i]).Elem().Interface()
		}
		i, err := h.Hydrate(m)
		if err != nil {
			return i.(T), err
		}
		*e = i.(T)
	}

	return *e, nil
}

func FromMap[T any, M ~map[string]any](m M) (T, error) {
	e := new(T)
	entityField := reflect.ValueOf(e).Elem().FieldByName("Entity")

	if entityField.IsValid() {
		if entityField.Type() != reflect.TypeOf(&Entity{}) {
			return *e, errors.New("Invalid embedded Entity type in " + reflect.ValueOf(e).Elem().Type().String() + ": must be an embedded *Entity")
		}

		ref, err := entityFromMap(m)

		if err != nil {
			return *e, err
		}

		entityField.Set(reflect.ValueOf(ref))
	}

	for k, v := range m {
		fieldRef, found := getFieldRef(e, k, true, true)

		if found && v != nil {
			if reflect.TypeOf(fieldRef).Elem().Kind() == reflect.Slice {
				panic("Does not currently handle slices")
			} else if reflect.TypeOf(fieldRef).Implements(typeOf[Unserializeable]()) {
				err := convertAssign(fieldRef, v)
				if err != nil {
					return *e, err
				}
			} else if reflect.TypeOf(v).Kind() == reflect.Slice {
				if reflect.ValueOf(v).Len() > 0 {
					err := convertAssign(fieldRef, reflect.ValueOf(v).Index(0).Interface())
					if err != nil {
						return *e, err
					}
				}
			} else {
				err := convertAssign(fieldRef, v)
				if err != nil {
					return *e, err
				}
			}
		}
	}

	if h, ok := any(e).(Hydratable); ok {
		i, err := h.Hydrate(m)
		if i, ok := i.(T); ok {
			if err != nil {
				return i, err
			}
			*e = i
		} else {
			return *e, errors.New(typeOf[T]().String() + "'s Hydrate() method must return an entity of the same type")
		}
	}

	return *e, nil
}

func ToMap[T any](entity T) map[string]any {
	return entityToMap(&entity, true, true, true)
}

func Flatten[T any](entities []T) []map[string]any {
	flattened := make([]map[string]any, 0)
	for i := 0; i < len(entities); i++ {
		flattened = append(flattened, ToMap(entities[i]))
	}
	return flattened
}

func flattenForSave[T IEntity](db *sql.DB, entities []T) []map[string]any {
	flattened := make([]map[string]any, 0)
	for i := 0; i < len(entities); i++ {
		pk := mustGetPrimaryKeyField(entities[i])
		table := getPrimaryKeyTable(pk)
		fields, err := doFilterChildren[T](&entities[i])
		if err != nil {
			return nil
		}
		fields = filterTableFields(db, table, fields)
		flattened = append(flattened, fields)
	}
	return flattened
}

func Inflate[T any, M ~map[string]any](mapSlice []M) ([]T, error) {
	inflated := make([]T, 0)
	for i := 0; i < len(mapSlice); i++ {
		e, err := FromMap[T](mapSlice[i])
		if err != nil {
			return nil, err
		}
		inflated = append(inflated, e)
	}
	return inflated, nil
}

func MapField[T any](fields map[string]any, key string) (T, bool) {
	if value, ok := fields[key]; ok {
		if v, ok := value.(T); ok {
			return v, true
		}
	}

	var t T
	return t, false
}

type Hydratable interface {
	Hydrate(map[string]any) (IEntity, error)
}

func NewEntity() *Entity {
	return &Entity{
		fields: make(map[string]field),
	}
}

type field struct {
	Type  *sql.ColumnType
	Value *any
}

type IEntity interface {
	entityFields() map[string]field
}

type Entity struct {
	fields map[string]field
}

func (e *Entity) entityFields() map[string]field {
	if e != nil && e.fields != nil {
		return e.fields
	}

	return nil
}

func entityFromRows(rows *sql.Rows) (*Entity, error) {
	e := NewEntity()

	columnTypes, err := rows.ColumnTypes()
	var scan = make([]any, 0)

	if err != nil {
		return nil, err
	}

	for _, column := range columnTypes {
		e.fields[column.Name()] = field{
			Type:  column,
			Value: new(any),
		}

		scan = append(scan, e.fields[column.Name()].Value)
	}

	err = rows.Scan(scan...)

	if err != nil {
		return nil, err
	}

	return e, nil
}

func entityFromMap(m map[string]any) (*Entity, error) {
	e := NewEntity()

	for k, v := range m {
		e.fields[k] = field{
			Value: new(any),
		}

		err := convertAssign(e.fields[k].Value, v)

		if err != nil {
			return nil, err
		}
	}

	return e, nil
}

func entityToMap[T any](entity *T, onlyUpdated bool, unserialize bool, recurse bool) map[string]any {
	return entityValueToMap(reflect.ValueOf(entity), onlyUpdated, unserialize, recurse)
}

func entityValueToMap(v reflect.Value, onlyUpdated bool, unserialize bool, recurse bool) map[string]any {
	t := v.Type()

	if t.Elem().Kind() == reflect.Pointer {
		panic("entity should not be passed as a pointer")
	}

	values := make(map[string]any)

	if e, ok := v.Interface().(IEntity); ok {
		ef := e.entityFields()

		if ef != nil {
			for key, field := range ef {
				if _, hasValue := values[key]; !hasValue {
					values[key] = *field.Value
				}
			}
		}

		if !recurse {
			_, has := getValuePrimaryKeyField(v)
			if has {
				values = filterStructFields(v, values, false, false)
			}
		}
	}

	for i := 0; i < v.Elem().NumField(); i++ {
		field := v.Elem().Field(i)
		fieldType := t.Elem().Field(i)
		fieldName := fieldType.Tag.Get("field")

		if fieldName != "" {
			value := field.Interface()

			if onlyUpdated && field.IsZero() {
				if v, has := values[fieldName]; has {
					if v != nil && reflect.ValueOf(v).IsZero() {
						continue
					}
				} else {
					continue
				}
			}

			if field.Kind() == reflect.Struct {
				if v, ok := value.(Unserializeable); ok && unserialize {
					values[fieldName] = v.Unserialize()
				} else if v, ok := value.(driver.Valuer); ok {
					val, err := v.Value()
					if err == nil {
						values[fieldName] = val
					}
				} else if _, ok := value.(IEntity); ok {
					relatedPk, hasPk := getPrimaryKeyField(value)
					if hasPk {
						if recurse {
							inner := entityValueToMap(field.Addr(), onlyUpdated, unserialize, recurse)
							for k, v := range inner {
								values[k] = v
							}
						} else {
							pkField, has := getField(field, relatedPk.Tag.Get("field"), false, false)

							if has && !pkField.Elem().IsZero() {
								values[fieldName] = pkField.Elem().Interface()
							}
						}
					}
				}
			} else {
				values[fieldName] = value
			}
		}
	}

	return values
}
