package db

import (
	"database/sql"
	"math"
)

type JoinType string

const (
	LeftJoin  = JoinType("LEFT JOIN")
	RightJoin = JoinType("RIGHT JOIN")
	InnerJoin = JoinType("INNER JOIN")
)

type Join struct {
	JoinType  JoinType
	Table     Value
	Condition *ConditionSet
}

type Ord string

const (
	Asc  = Ord("ASC")
	Desc = Ord("DESC")

	OffsetStart = 0
	Unlimited   = math.MaxUint32
)

type Order struct {
	Field Value
	Ord   Ord
}

type QueryType string

const (
	Select       = QueryType("SELECT")
	Insert       = QueryType("INSERT")
	InsertIgnore = QueryType("INSERT IGNORE")
	InsertUpdate = QueryType("INSERT UPDATE")
	Update       = QueryType("UPDATE")
	Delete       = QueryType("DELETE")
)

type Offset struct {
	Start uint
	Limit uint
}

type UnionType string

const (
	UnionDefault = UnionType("UNION")
	UnionAll     = UnionType("UNION ALL")
)

type Union struct {
	Query     *QueryBuilder
	UnionType UnionType
}

type QueryBuilder struct {
	Type            QueryType
	Fields          List
	FieldsCleared   bool
	Values          map[string]any
	PrimaryTable    Value
	Alias           Ident
	Joins           []Join
	WhereCondition  *ConditionSet
	GroupBys        List
	HavingCondition *ConditionSet
	OrderBys        []Order
	OrderBysCleared bool
	Offset          Offset
	Unions          []Union
}

func NewQuery() *QueryBuilder {
	return &QueryBuilder{
		Fields:          make([]Value, 0),
		PrimaryTable:    nil,
		Values:          make(map[string]any),
		WhereCondition:  Condition(),
		GroupBys:        make([]Value, 0),
		HavingCondition: Condition(),
		OrderBys:        make([]Order, 0),
		Unions:          make([]Union, 0),
	}
}

func (q *QueryBuilder) ComposeWith(queries ...*QueryBuilder) *QueryBuilder {
	for _, query := range queries {
		q.compose(query)
	}
	return q
}

func (q *QueryBuilder) compose(query *QueryBuilder) {
	if query.FieldsCleared {
		q.Fields = make([]Value, 0)
		q.FieldsCleared = true
	}
	q.Fields = append(q.Fields, query.Fields...)

	for k, v := range query.Values {
		q.Values[k] = v
	}

	if query.PrimaryTable != nil {
		q.PrimaryTable = query.PrimaryTable
	}

	if query.Alias != "" {
		q.Alias = query.Alias
	}

	q.Joins = append(q.Joins, query.Joins...)
	q.WhereCondition.Conditions = append(q.WhereCondition.Conditions, query.WhereCondition.Conditions...)
	q.GroupBys = append(q.GroupBys, query.GroupBys...)
	q.HavingCondition.Conditions = append(q.HavingCondition.Conditions, query.HavingCondition.Conditions...)

	if query.OrderBysCleared {
		q.OrderBys = make([]Order, 0)
	}
	q.OrderBys = append(q.OrderBys, query.OrderBys...)

	if query.Offset.Start != OffsetStart || query.Offset.Limit != Unlimited {
		q.Offset = query.Offset
	}

	q.Unions = append(q.Unions, query.Unions...)
}

func (q *QueryBuilder) Select(fields ...any) *QueryBuilder {
	q.Type = Select

	for _, f := range fields {
		q.AddField(f)
	}

	return q
}

func (q *QueryBuilder) InsertInto(tableName string) *QueryBuilder {
	q.Type = Insert
	q.PrimaryTable = Ident(tableName)
	return q
}

func (q *QueryBuilder) InsertIgnoreInto(tableName string) *QueryBuilder {
	q.Type = InsertIgnore
	q.PrimaryTable = Ident(tableName)
	return q
}

func (q *QueryBuilder) InsertUpdateInto(tableName string) *QueryBuilder {
	q.Type = InsertUpdate
	q.PrimaryTable = Ident(tableName)
	return q
}

func (q *QueryBuilder) Update(tableName string) *QueryBuilder {
	q.Type = Update
	q.PrimaryTable = Ident(tableName)
	return q
}

func (q *QueryBuilder) DeleteFrom(tableName string) *QueryBuilder {
	q.Type = Delete
	q.PrimaryTable = Ident(tableName)
	return q
}

func (q *QueryBuilder) Delete(fields ...any) *QueryBuilder {
	q.Type = Delete

	for _, f := range fields {
		q.AddField(f)
	}

	return q
}

func (q *QueryBuilder) From(table any) *QueryBuilder {
	q.PrimaryTable = LValue(table)
	return q
}

func (q *QueryBuilder) As(alias string) *QueryBuilder {
	q.Alias = Ident(alias)
	return q
}

func (q *QueryBuilder) AddField(f any) *QueryBuilder {
	q.Fields = append(q.Fields, LValue(f))
	return q
}

func (q *QueryBuilder) ClearFields() *QueryBuilder {
	q.Fields = make([]Value, 0)
	q.FieldsCleared = true
	return q
}

func (q *QueryBuilder) Set(values any) *QueryBuilder {
	var vs map[string]any
	var ok bool

	if vs, ok = values.(map[string]any); !ok {
		panic("Invalid query type for Set()")
	}

	for k, v := range vs {
		q.Values[k] = v
	}

	return q
}

func (q *QueryBuilder) LeftJoin(table any, condition *ConditionSet) *QueryBuilder {
	q.Joins = append(
		q.Joins,
		Join{LeftJoin, LValue(table), condition},
	)
	return q
}

func (q *QueryBuilder) LeftJoinEq(table any, left any, right any) *QueryBuilder {
	q.Joins = append(
		q.Joins,
		Join{
			JoinType:  LeftJoin,
			Table:     LValue(table),
			Condition: Condition().Eq(LValue(left), LValue(right)),
		},
	)
	return q
}

func (q *QueryBuilder) InnerJoin(table any, condition *ConditionSet) *QueryBuilder {
	q.Joins = append(
		q.Joins,
		Join{InnerJoin, LValue(table), condition},
	)
	return q
}

func (q *QueryBuilder) InnerJoinEq(table any, left any, right any) *QueryBuilder {
	q.Joins = append(
		q.Joins,
		Join{
			JoinType:  InnerJoin,
			Table:     LValue(table),
			Condition: Condition().Eq(LValue(left), LValue(right)),
		},
	)
	return q
}

func (q *QueryBuilder) RightJoin(table any, condition *ConditionSet) *QueryBuilder {
	q.Joins = append(
		q.Joins,
		Join{RightJoin, LValue(table), condition},
	)
	return q
}

func (q *QueryBuilder) RightJoinEq(table any, left any, right any) *QueryBuilder {
	q.Joins = append(
		q.Joins,
		Join{
			JoinType:  RightJoin,
			Table:     LValue(table),
			Condition: Condition().Eq(LValue(left), LValue(right)),
		},
	)
	return q
}

func (q *QueryBuilder) Where(c *ConditionSet) *QueryBuilder {
	q.WhereCondition.Condition(c)
	return q
}

func (q *QueryBuilder) WhereLike(left any, right any) *QueryBuilder {
	q.WhereCondition.Like(left, right)
	return q
}

func (q *QueryBuilder) WhereNot(c *ConditionSet) *QueryBuilder {
	c.Not = !c.Not
	q.WhereCondition.Condition(c)
	return q
}

func (q *QueryBuilder) WhereEq(left any, right any) *QueryBuilder {
	q.WhereCondition.Eq(left, right)
	return q
}

func (q *QueryBuilder) WhereNotEq(left any, right any) *QueryBuilder {
	q.WhereCondition.NotEq(left, right)
	return q
}

func (q *QueryBuilder) WhereLt(left any, right any) *QueryBuilder {
	q.WhereCondition.Lt(left, right)
	return q
}

func (q *QueryBuilder) WhereLtEq(left any, right any) *QueryBuilder {
	q.WhereCondition.LtEq(left, right)
	return q
}

func (q *QueryBuilder) WhereGt(left any, right any) *QueryBuilder {
	q.WhereCondition.Gt(left, right)
	return q
}

func (q *QueryBuilder) WhereGtEq(left any, right any) *QueryBuilder {
	q.WhereCondition.GtEq(left, right)
	return q
}

func (q *QueryBuilder) WhereIn(left any, right any) *QueryBuilder {
	q.WhereCondition.In(left, right)
	return q
}

func (q *QueryBuilder) WhereNotIn(left any, right any) *QueryBuilder {
	q.WhereCondition.NotIn(left, right)
	return q
}

func (q *QueryBuilder) WhereIsNull(value any) *QueryBuilder {
	q.WhereCondition.IsNull(value)
	return q
}

func (q *QueryBuilder) WhereIsNotNull(value any) *QueryBuilder {
	q.WhereCondition.IsNotNull(value)
	return q
}

func (q *QueryBuilder) WhereIsTrue(value any) *QueryBuilder {
	q.WhereCondition.IsTrue(value)
	return q
}

func (q *QueryBuilder) WhereIsNotTrue(value any) *QueryBuilder {
	q.WhereCondition.IsNotTrue(value)
	return q
}

func (q *QueryBuilder) WhereIsFalse(value any) *QueryBuilder {
	q.WhereCondition.IsFalse(value)
	return q
}

func (q *QueryBuilder) WhereIsNotFalse(value any) *QueryBuilder {
	q.WhereCondition.IsNotFalse(value)
	return q
}

func (q *QueryBuilder) GroupBy(field any) *QueryBuilder {
	q.GroupBys = append(q.GroupBys, LValue(field))
	return q
}

func (q *QueryBuilder) Having(c *ConditionSet) *QueryBuilder {
	q.HavingCondition.Condition(c)
	return q
}

func (q *QueryBuilder) HavingNot(c *ConditionSet) *QueryBuilder {
	c.Not = !c.Not
	q.HavingCondition.Condition(c)
	return q
}

func (q *QueryBuilder) HavingEq(left any, right any) *QueryBuilder {
	q.HavingCondition.Eq(left, right)
	return q
}

func (q *QueryBuilder) HavingNotEq(left any, right any) *QueryBuilder {
	q.HavingCondition.NotEq(left, right)
	return q
}

func (q *QueryBuilder) HavingLt(left any, right any) *QueryBuilder {
	q.HavingCondition.Lt(left, right)
	return q
}

func (q *QueryBuilder) HavingLtEq(left any, right any) *QueryBuilder {
	q.HavingCondition.LtEq(left, right)
	return q
}

func (q *QueryBuilder) HavingGt(left any, right any) *QueryBuilder {
	q.HavingCondition.Gt(left, right)
	return q
}

func (q *QueryBuilder) HavingGtEq(left any, right any) *QueryBuilder {
	q.HavingCondition.GtEq(left, right)
	return q
}

func (q *QueryBuilder) HavingIn(left any, right any) *QueryBuilder {
	q.HavingCondition.In(left, right)
	return q
}

func (q *QueryBuilder) HavingNotIn(left any, right any) *QueryBuilder {
	q.HavingCondition.NotIn(left, right)
	return q
}

func (q *QueryBuilder) HavingIsNull(value any) *QueryBuilder {
	q.HavingCondition.IsNull(value)
	return q
}

func (q *QueryBuilder) HavingIsNotNull(value any) *QueryBuilder {
	q.HavingCondition.IsNotNull(value)
	return q
}

func (q *QueryBuilder) HavingIsTrue(value any) *QueryBuilder {
	q.HavingCondition.IsTrue(value)
	return q
}

func (q *QueryBuilder) HavingIsNotTrue(value any) *QueryBuilder {
	q.HavingCondition.IsNotTrue(value)
	return q
}

func (q *QueryBuilder) HavingIsFalse(value any) *QueryBuilder {
	q.HavingCondition.IsFalse(value)
	return q
}

func (q *QueryBuilder) HavingIsNotFalse(value any) *QueryBuilder {
	q.HavingCondition.IsNotFalse(value)
	return q
}

func (q *QueryBuilder) OrderBy(field any, ord Ord) *QueryBuilder {
	q.OrderBys = append(q.OrderBys, Order{Field: LValue(field), Ord: ord})
	return q
}

func (q *QueryBuilder) ClearOrderBys() *QueryBuilder {
	q.OrderBysCleared = true
	q.OrderBys = make([]Order, 0)
	return q
}

func (q *QueryBuilder) Limit(start uint, limit uint) *QueryBuilder {
	q.Offset = Offset{start, limit}
	return q
}

func (q *QueryBuilder) Union(query *QueryBuilder) *QueryBuilder {
	q.Unions = append(q.Unions, Union{Query: query, UnionType: UnionDefault})
	return q
}

func (q *QueryBuilder) UnionAll(query *QueryBuilder) *QueryBuilder {
	q.Unions = append(q.Unions, Union{Query: query, UnionType: UnionAll})
	return q
}

func (q *QueryBuilder) ValidateSQL() error {
	return nil
}

func (q *QueryBuilder) Transcribe(db *sql.DB) (string, []any, error) {
	transcriber := getTranscriber(db.Driver())
	return transcriber.Transcribe(q)
}
