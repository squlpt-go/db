package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
)

type IDType interface {
	~string | ~int | ~int64 | ~uint64
}

type EntitySet[T any] interface {
	[]T
}

type EntityOrMapSet[T any] interface {
	EntitySet[T] | []map[string]any
}

func GetRows[T IEntity](db *sql.DB, qs ...*QueryBuilder) *Rows[T] {
	s := new(T)
	pk := mustGetPrimaryKeyField(s)
	table := getPrimaryKeyTable(pk)

	q := NewQuery().
		Select("*").
		From(table)

	for _, r := range getManyToOneRelations(db, table) {
		q.LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		).AddField(TableField(r.parent(), "*"))
	}

	q.ComposeWith(qs...)

	return Query[T](db, q)
}

func GetCount[T IEntity](db *sql.DB, qs ...*QueryBuilder) int64 {
	s := new(T)
	pk := mustGetPrimaryKeyField(s)
	table := getPrimaryKeyTable(pk)

	query := NewQuery().
		Select(Raw("COUNT(*) AS count")).
		From(table).
		ComposeWith(qs...)

	transcriber := getTranscriber(db.Driver())
	q, args, err := transcriber.Transcribe(query)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		m := fmt.Sprintf("%s: \"%s\" args: %v", err, q, args)
		panic(m)
	}
	defer rows.Close()

	var count int64
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		panic(err)
	}

	return count
}

func GetRowById[T IEntity, I IDType](db *sql.DB, id I, qs ...*QueryBuilder) (T, bool) {
	s := new(T)
	pk := mustGetPrimaryKeyField(s)
	table := getPrimaryKeyTable(pk)
	pkField := TableField(table, pk.Tag.Get("field"))

	r := getTableRowsByValue(db, table, string(pkField), id, qs...)
	return As[T](r).Row()
}

func InsertRow[T IEntity](db *sql.DB, entity T) (*Result, error) {
	if reflect.ValueOf(entity).IsZero() {
		panic("Cannot insert zero entity " + reflect.TypeOf(entity).String())
	}

	pk := mustGetPrimaryKeyField(entity)
	table := getPrimaryKeyTable(pk)

	fields, err := doFilterInsert[T](&entity)
	if err != nil {
		return nil, err
	}
	fields = filterTableFields(db, table, fields)

	if len(fields) == 0 {
		panic("no fields to insert")
	}

	q := NewQuery().
		InsertInto(table).
		Set(fields)

	return Exec(db, q), nil
}

func UpdateRow[T IEntity](db *sql.DB, entity T) (*Result, error) {
	if reflect.ValueOf(entity).IsZero() {
		panic("Cannot insert zero entity " + reflect.TypeOf(entity).String())
	}

	pk := mustGetPrimaryKeyField(entity)
	table := getPrimaryKeyTable(pk)

	fields, err := doFilterUpdate[T](&entity)
	if err != nil {
		return nil, err
	}
	fields = filterTableFields(db, table, fields)

	if len(fields) == 0 {
		panic("no fields to update")
	}

	pkFieldName := pk.Tag.Get("field")
	pkFieldValue, hasPkField := fields[pkFieldName]
	if !hasPkField {
		panic("No primary key value " + pkFieldName + "provided")
	}

	q := NewQuery().
		Update(table).
		Set(fields).
		WhereEq(pkFieldName, pkFieldValue)

	return Exec(db, q), nil
}

func DeleteRow[T IEntity](db *sql.DB, entity T) (*Result, error) {
	if reflect.ValueOf(entity).IsZero() {
		panic("Cannot delete zero entity " + reflect.TypeOf(entity).String())
	}

	pk := mustGetPrimaryKeyField(entity)
	table := getPrimaryKeyTable(pk)
	fields := entityToMap(&entity, false, false)

	pkFieldName := pk.Tag.Get("field")
	pkFieldValue, hasPkField := fields[pkFieldName]

	if !hasPkField {
		panic("No primary key value " + pkFieldName + "provided")
	}

	q := NewQuery().
		DeleteFrom(table).
		WhereEq(pkFieldName, pkFieldValue)

	return Exec(db, q), nil
}

func GetChildren[Parent IEntity, Children IEntity, I IDType](db *sql.DB, id I, queries ...*QueryBuilder) *Rows[Children] {
	var p Parent
	var c Children
	pt := mustGetTable(&p)
	ct := mustGetTable(&c)

	relation := getRelation(db, pt, ct)
	if relation == nil {
		panic("Invalid relation " + ct + " -> " + pt)
	}

	q := NewQuery().
		Select(TableField(ct, "*")).
		From(ct).
		ComposeWith(relation.getChildrenQuery(id))

	for _, r := range getManyToOneRelations(db, ct) {
		if r.parent() != pt {
			q.LeftJoinEq(
				r.parent(),
				Ident(r.parentKey()),
				Ident(r.childKey()),
			).AddField(TableField(r.parent(), "*"))
		}
	}

	q.ComposeWith(queries...)

	return Query[Children](db, q)
}

func AssignChildren[Parent IEntity, Children IEntity, I IDType](db *sql.DB, parentId I, childIds []I, subtractive bool) {
	var p Parent
	var c Children
	pt := mustGetTable(&p)
	ct := mustGetTable(&c)
	cpk := mustGetPrimaryKeyFieldName(&c)

	relation := getRelation(db, pt, ct)
	if relation == nil {
		panic("Invalid relation " + ct + " -> " + pt)
	}

	err := relation.assignChildren(db, asString(parentId), cpk, stringIds(childIds), subtractive)
	if err != nil {
		panic(err)
	}
}

func SetChildren[Parent IEntity, Children IEntity, S EntitySet[Children]](db *sql.DB, parentId any, childEntities S, subtractive bool) error {
	var p Parent
	var c Children

	pt := mustGetTable(&p)
	ct := mustGetTable(&c)
	cpk := mustGetPrimaryKeyFieldName(&c)

	relation := getRelation(db, pt, ct)
	if relation == nil {
		panic("Invalid relation " + ct + " -> " + pt)
	}

	var err error

	switch es := any(childEntities).(type) {
	case []map[string]any:
		err = relation.setChildren(db, asString(parentId), cpk, es, subtractive)
	case []Children:
		err = relation.setChildren(db, asString(parentId), cpk, flattenForSave[Children](db, es), subtractive)
	case []IEntity:
		err = relation.setChildren(db, asString(parentId), cpk, flattenForSave[IEntity](db, es), subtractive)
	default:
		panic("invalid entity set type")
	}

	if err != nil {
		panic(err)
	}

	return nil
}

func GetTableFields(db *sql.DB, table string) []string {
	if fs, ok := _tableFields[table]; ok {
		return fs
	}

	query := NewQuery().Select("*").From(table).WhereEq(1, 0)
	transcriber := getTranscriber(db.Driver())
	q, args, err := transcriber.Transcribe(query)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		m := fmt.Sprintf("Failed fetching fields from %s: %s: \"%s\" args: %v", table, err, q, args)
		panic(m)
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	_tableFields[table] = columns

	return columns
}

var _tableFields = make(map[string][]string)

func tableHasField(db *sql.DB, table string, field string) bool {
	fields := GetTableFields(db, table)
	return slices.Contains(fields, field)
}

func filterTableFields(db *sql.DB, table string, fields map[string]any) map[string]any {
	f := make(map[string]any)

	for k, v := range fields {
		if tableHasField(db, table, k) {
			f[k] = v
		}
	}

	return f
}

func getField(v reflect.Value, name string, readOnly bool) (reflect.Value, bool) {
	t := v.Type()

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		if (t.Field(i).Tag.Get("field") == name && t.Field(i).Tag.Get("foreign") == "") ||
			(readOnly && t.Field(i).Tag.Get("computed") == name) {
			return v.Field(i).Addr(), true
		}
	}

	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct {
			if fieldRef, found := getField(v.Field(i).Addr(), name, readOnly); found {
				return fieldRef, true
			}
		}
	}

	return reflect.Zero(t), false
}

func getFieldRef(entity any, name string, readOnly bool) (any, bool) {
	f, ok := getField(reflect.ValueOf(entity), name, readOnly)
	return f.Interface(), ok
}

func getFieldRefByColumnType(entity any, columnType *sql.ColumnType, readOnly bool) (any, bool) {
	return getFieldRef(entity, columnType.Name(), readOnly)
}

func getPrimaryKeyField(entity any) (reflect.StructField, bool) {
	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if getPrimaryKeyTable(field) != "" {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

func mustGetPrimaryKeyField(entity any) reflect.StructField {
	f, has := getPrimaryKeyField(entity)
	t := reflect.TypeOf(entity)

	if !has {
		panic("No primary key defined in " + t.String())
	}

	return f
}

func mustGetPrimaryKeyFieldName(entity any) string {
	return mustGetPrimaryKeyField(entity).Tag.Get("field")
}

func getPrimaryKeyTable(field reflect.StructField) string {
	return field.Tag.Get("primary")
}

func getTable(entity any) (string, bool) {
	pk, has := getPrimaryKeyField(entity)

	if !has {
		return "", false
	}

	return getPrimaryKeyTable(pk), true
}

func mustGetTable(entity any) string {
	table, has := getTable(entity)

	if !has {
		panic(reflect.TypeOf(entity).String() + " has not table defined")
	}

	return table
}

func getTableRowsByValue(db *sql.DB, table string, field string, value any, qs ...*QueryBuilder) *sql.Rows {
	q := NewQuery().
		From(table).
		WhereEq(field, value).
		ComposeWith(qs...)

	for _, r := range getManyToOneRelations(db, table) {
		q.LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		).AddField(TableField(r.parent(), "*"))
	}

	q.Select(TableField(table, "*"))

	return queryStd(db, q)
}
