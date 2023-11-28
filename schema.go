package db

import (
	"database/sql"
	"reflect"
)

type Relation interface {
	getChildrenQuery(id any) *QueryBuilder
	joinParentsQuery() *QueryBuilder
	assignChildren(db *sql.DB, parentId string, childPk string, childIds []string, subtractive bool) error
	setChildren(db *sql.DB, parentId string, childPk string, childEntities []map[string]any, subtractive bool) error
	from() string
	to() string
}

type ManyToOneDef struct {
	FromTable string
	FromField string
	ToTable   string
	ToKey     string
}

func (r ManyToOneDef) getChildrenQuery(_ any) *QueryBuilder {
	panic("ManyToOne " + r.child() + " -> " + r.parent() + " relation does not have children")
}

func (r ManyToOneDef) joinParentsQuery() *QueryBuilder {
	return NewQuery().
		AddField(
			TableField(r.parent(), "*"),
		).
		LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		)
}

func (r ManyToOneDef) assignChildren(_ *sql.DB, _ string, _ string, _ []string, _ bool) error {
	panic("ManyToOne " + r.child() + " -> " + r.parent() + " relation does not have children")
}

func (r ManyToOneDef) setChildren(_ *sql.DB, _ string, _ string, _ []map[string]any, _ bool) error {
	panic("ManyToOne " + r.child() + " -> " + r.parent() + " relation does not have children")
}

func (r ManyToOneDef) parent() string {
	return r.ToTable
}

func (r ManyToOneDef) child() string {
	return r.FromTable
}

func (r ManyToOneDef) parentKey() string {
	return r.ToKey
}

func (r ManyToOneDef) childKey() string {
	return r.FromField
}

func (r ManyToOneDef) from() string {
	return r.FromTable
}

func (r ManyToOneDef) to() string {
	return r.ToTable
}

type OneToManyDef struct {
	FromTable string
	FromField string
	ToTable   string
	ToField   string
}

func (r OneToManyDef) getChildrenQuery(id any) *QueryBuilder {
	return NewQuery().
		AddField(
			TableField(r.parent(), "*"),
		).
		LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		).
		WhereEq(
			Ident(r.childKey()),
			id,
		)
}

func (r OneToManyDef) joinParentsQuery() *QueryBuilder {
	panic("OneToMany " + r.child() + " -> " + r.parent() + " relation does not have parents")
}

func (r OneToManyDef) assignChildren(db *sql.DB, parentId string, childPk string, childIds []string, subtractive bool) error {
	cpk := childPk

	q1 := NewQuery().
		Select(cpk).
		From(r.child()).
		LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		).
		WhereEq(r.parentKey(), parentId)
	r1 := queryStd(db, q1)

	existingIds := make([]string, 0)
	for r1.Next() {
		var id string
		err := r1.Scan(&id)
		if err != nil {
			panic(err)
		}
		existingIds = append(existingIds, id)
	}

	newIds := idDiff(stringIds(childIds), existingIds)

	q2 := NewQuery().
		Update(r.child()).
		LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		).
		Set(map[string]any{
			r.childKey(): parentId,
		}).
		WhereIn(cpk, newIds)
	r2 := Exec(db, q2)
	c2, err := r2.RowsAffected()

	if err != nil {
		panic(err)
	}

	if c2 != int64(len(newIds)) {
		panic("unexpected number of rows affected, you may have passed in an invalid ID")
	}

	if subtractive {
		rmIds := idDiff(existingIds, stringIds(childIds))

		q3 := NewQuery().
			Update(r.child()).
			Set(map[string]any{
				r.childKey(): nil,
			}).
			WhereEq(r.childKey(), parentId).
			WhereNotIn(cpk, childIds)
		r3 := Exec(db, q3)
		c3, err := r3.RowsAffected()

		if err != nil {
			panic(err)
		}

		if c3 != int64(len(rmIds)) {
			panic("unexpected number of rows affected")
		}
	}

	return nil
}

func (r OneToManyDef) setChildren(db *sql.DB, parentId string, childPk string, childEntities []map[string]any, subtractive bool) error {
	cpk := childPk
	ppk := r.parentKey()

	q1 := NewQuery().
		Select(cpk).
		From(r.child()).
		LeftJoinEq(
			r.parent(),
			Ident(r.parentKey()),
			Ident(r.childKey()),
		).
		WhereEq(r.parentKey(), parentId)
	r1 := queryStd(db, q1)

	existingIds := make([]string, 0)
	for r1.Next() {
		var id string
		err := r1.Scan(&id)
		if err != nil {
			panic(err)
		}
		existingIds = append(existingIds, id)
	}

	childIds := make([]string, 0)
	for i := 0; i < len(childEntities); i++ {
		c := childEntities[i]
		c[r.childKey()] = parentId
		_, has := c[cpk]
		v := reflect.ValueOf(c[cpk])
		exists := false

		if has && !v.IsZero() {
			iq := NewQuery().
				Select(cpk).
				From(r.child()).
				WhereEq(cpk, c[cpk])
			row := queryStd(db, iq)
			if row != nil && row.Err() == nil && row.Next() {
				exists = true
			}
		}

		if exists {
			q2 := NewQuery().
				Update(r.child()).
				Set(c).
				WhereEq(cpk, c[cpk])
			r2 := Exec(db, q2)
			_, err := r2.RowsAffected()

			if err != nil {
				panic(err)
			}

			childIds = append(childIds, asString(c[cpk]))
		} else {
			q2 := NewQuery().
				InsertInto(r.child()).
				Set(c)
			r2 := Exec(db, q2)
			c2, err := r2.RowsAffected()

			if err != nil {
				panic(err)
			}

			if c2 != 1 {
				panic("unexpected number of rows affected")
			}

			if _, ok := c[cpk]; ok {
				childIds = append(childIds, asString(c[cpk]))
			} else {
				id, err := r2.LastInsertId()
				if err != nil {
					panic(err)
				}
				childIds = append(childIds, asString(id))
			}
		}
	}

	if subtractive {
		rmIds := idDiff(existingIds, stringIds(childIds))

		q2 := NewQuery().
			DeleteFrom(r.child()).
			AddField(TableField(r.child(), "*")).
			LeftJoinEq(
				r.parent(),
				Ident(r.parentKey()),
				Ident(r.childKey()),
			).
			WhereEq(ppk, parentId).
			WhereNotIn(cpk, childIds)
		r2 := Exec(db, q2)
		c2, err := r2.RowsAffected()

		if err != nil {
			panic(err)
		}

		if c2 != int64(len(rmIds)) {
			panic("unexpected number of rows deleted")
		}
	}

	return nil
}

func (r OneToManyDef) parent() string {
	return r.FromTable
}

func (r OneToManyDef) child() string {
	return r.ToTable
}

func (r OneToManyDef) parentKey() string {
	return r.FromField
}

func (r OneToManyDef) childKey() string {
	return r.ToField
}

func (r OneToManyDef) from() string {
	return r.FromTable
}

func (r OneToManyDef) to() string {
	return r.ToTable
}

type ManyToManyDef struct {
	FromTable      string
	FromKey        string
	ToTable        string
	ToKey          string
	ThroughTable   string
	ThroughFromKey string
	ThroughToKey   string
}

func (r ManyToManyDef) getChildrenQuery(id any) *QueryBuilder {
	return NewQuery().
		Select(
			TableField(r.ThroughTable, "*"),
			TableField(r.parent(), "*"),
		).
		LeftJoinEq(
			r.ThroughTable,
			Ident(r.ThroughToKey),
			Ident(r.childKey()),
		).
		LeftJoinEq(
			r.parent(),
			Ident(r.ThroughFromKey),
			r.parentKey(),
		).
		WhereEq(
			Ident(r.ThroughFromKey),
			id,
		)
}

func (r ManyToManyDef) joinParentsQuery() *QueryBuilder {
	panic("ManyToMany " + r.parent() + " <-> " + r.child() + " relation does not have a single Parent")
}

func (r ManyToManyDef) assignChildren(db *sql.DB, parentId string, childPk string, childIds []string, subtractive bool) error {
	children := make([]map[string]any, 0)
	for i := 0; i < len(childIds); i++ {
		children = append(children, map[string]any{
			r.ThroughFromKey: parentId,
			childPk:          childIds[i],
		})
	}
	return r.setAssignChildren(db, parentId, childPk, children, subtractive, false)
}

func (r ManyToManyDef) setChildren(db *sql.DB, parentId string, childPk string, childEntities []map[string]any, subtractive bool) error {
	return r.setAssignChildren(db, parentId, childPk, childEntities, subtractive, true)
}

func (r ManyToManyDef) setAssignChildren(db *sql.DB, parentId string, childPk string, childEntities []map[string]any, subtractive bool, set bool) error {
	cpk := childPk

	q1 := NewQuery().
		Select(r.ThroughToKey).
		From(r.ThroughTable).
		WhereEq(r.ThroughFromKey, parentId)
	r1 := queryStd(db, q1)

	existingIds := make([]string, 0)
	for r1.Next() {
		var id string
		err := r1.Scan(&id)
		if err != nil {
			panic(err)
		}
		existingIds = append(existingIds, id)
	}

	childIds := make([]string, 0)

	for i := 0; i < len(childEntities); i++ {
		c := childEntities[i]

		if set {
			_, has := c[cpk]
			v := reflect.ValueOf(c[cpk])
			exists := false

			if has && !v.IsZero() {
				iq := NewQuery().
					Select(cpk).
					From(r.child()).
					WhereEq(cpk, c[cpk])
				r := queryStd(db, iq)
				if r != nil && r.Next() {
					exists = true
				}
			}

			if exists {
				q2 := NewQuery().
					Update(r.child()).
					Set(c).
					WhereEq(cpk, c[cpk])
				r2 := Exec(db, q2)
				_, err := r2.RowsAffected()

				if err != nil {
					panic(err)
				}

				childIds = append(childIds, asString(c[cpk]))
			} else {
				q2 := NewQuery().
					InsertInto(r.child()).
					Set(c)
				r2 := Exec(db, q2)
				c2, err := r2.RowsAffected()

				if err != nil {
					panic(err)
				}

				if c2 != 1 {
					panic("unexpected number of rows affected")
				}

				if _, ok := c[cpk]; ok {
					childIds = append(childIds, asString(c[cpk]))
				} else {
					id, err := r2.LastInsertId()
					if err != nil {
						panic(err)
					}
					childIds = append(childIds, asString(id))
				}
			}
		} else {
			childIds = append(childIds, asString(c[cpk]))
		}
	}

	newIds := idDiff(stringIds(childIds), existingIds)

	for i := 0; i < len(newIds); i++ {
		iq := NewQuery().
			InsertInto(r.ThroughTable).
			Set(map[string]any{
				r.ThroughFromKey: parentId,
				r.ThroughToKey:   newIds[i],
			})
		ir := Exec(db, iq)
		ic, err := ir.RowsAffected()
		if err != nil {
			panic(err)
		}
		if ic != 1 {
			panic("unexpected number of rows affected")
		}
	}

	if subtractive {
		rmIds := idDiff(existingIds, stringIds(childIds))

		q2 := NewQuery().
			DeleteFrom(r.ThroughTable).
			WhereEq(r.ThroughFromKey, parentId).
			WhereNotIn(r.ThroughToKey, childIds)
		r2 := Exec(db, q2)
		c2, err := r2.RowsAffected()

		if err != nil {
			panic(err)
		}

		if c2 != int64(len(rmIds)) {
			panic("unexpected number of rows deleted")
		}
	}

	return nil
}

func (r ManyToManyDef) parent() string {
	return r.FromTable
}

func (r ManyToManyDef) child() string {
	return r.ToTable
}

func (r ManyToManyDef) parentKey() string {
	return r.FromKey
}

func (r ManyToManyDef) childKey() string {
	return r.ToKey
}

func (r ManyToManyDef) from() string {
	return r.FromTable
}

func (r ManyToManyDef) to() string {
	return r.ToTable
}

var schema = make(map[*sql.DB][]Relation)

func DefRelation(db *sql.DB, def Relation) {
	schema[db] = append(schema[db], def)
}

func ManyToOne[Child any, Parent any](db *sql.DB) {
	var child Child
	var parent Parent

	childTable := mustGetTable(&child)
	parentTable := mustGetTable(&parent)
	parentPk := mustGetPrimaryKeyFieldName(&parent)

	DefRelation(db, ManyToOneDef{
		childTable,
		string(TableField(childTable, parentPk)),
		parentTable,
		string(TableField(parentTable, parentPk)),
	})
}

func OneToMany[Parent any, Child any](db *sql.DB) {
	var child Child
	var parent Parent

	childTable := mustGetTable(&child)
	parentTable := mustGetTable(&parent)
	parentPk := mustGetPrimaryKeyFieldName(&parent)

	DefRelation(db, OneToManyDef{
		parentTable,
		string(TableField(parentTable, parentPk)),
		childTable,
		string(TableField(childTable, parentPk)),
	})
}

func ManyToMany[Parent any, Child any](db *sql.DB, throughTable string) {
	var child Child
	var parent Parent

	childTable := mustGetTable(&child)
	parentTable := mustGetTable(&parent)
	childPk := mustGetPrimaryKeyFieldName(&child)
	parentPk := mustGetPrimaryKeyFieldName(&parent)

	DefRelation(db, ManyToManyDef{
		parentTable,
		string(TableField(parentTable, parentPk)),
		childTable,
		string(TableField(childTable, childPk)),
		throughTable,
		string(TableField(throughTable, parentPk)),
		string(TableField(throughTable, childPk)),
	})
}

func getRelation(db *sql.DB, from string, to string) Relation {
	relations, ok := schema[db]

	if !ok {
		return nil
	}

	for i := 0; i < len(relations); i++ {
		relation := relations[i]

		if relation.from() == from && relation.to() == to {
			return relation
		}
	}

	return nil
}

func getManyToOneRelations(db *sql.DB, from string) []ManyToOneDef {
	relations, ok := schema[db]
	out := make([]ManyToOneDef, 0)

	if !ok {
		return out
	}

	for i := 0; i < len(relations); i++ {
		relation := relations[i]
		if r, ok := relation.(ManyToOneDef); ok && relation.from() == from {
			out = append(out, r)
		}
	}

	return out
}

func stringIds[I IDType](ids []I) []string {
	out := make([]string, 0)

	for i := 0; i < len(ids); i++ {
		out = append(out, asString(ids[i]))
	}

	return out
}

func idDiff(newIds, oldIds []string) []string {
	existing := make(map[string]bool)
	for _, val := range oldIds {
		existing[val] = true
	}

	result := make([]string, 0)

	for _, val := range newIds {
		if !existing[val] {
			result = append(result, val)
		}
	}

	return result
}
