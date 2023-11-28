package db

import (
	"reflect"
	"testing"
)

func TestOrder_ValidateSQL(t *testing.T) {
	o := Order{
		Field: Ident("name"),
		Ord:   Asc,
	}

	err := o.Field.ValidateSQL()
	if err != nil {
		t.Errorf("Order.ValidateSQL() returned an error for valid field")
	}
}

func TestQuery_ComposeWith(t *testing.T) {
	q1 := NewQuery().Select("name").From("users").WhereEq("age", 18)
	q2 := NewQuery().Select("email").From("users").WhereEq("active", true)

	q1.ComposeWith(q2)

	if q1.Type != Select {
		t.Error("ComposeWith() did not set the correct query type")
	}

	if len(q1.Fields) != 2 {
		t.Error("ComposeWith() did not add the fields from the composed query")
	}

	if q1.PrimaryTable.(Ident) != "users" {
		t.Error("ComposeWith() did not set the correct primary table")
	}

	if len(q1.Joins) != 0 {
		t.Error("ComposeWith() added joins when it should not")
	}

	if len(q1.WhereCondition.Conditions) != 2 {
		t.Error("ComposeWith() did not add the conditions from the composed query")
	}
}

func TestQuery_Select(t *testing.T) {
	q := NewQuery().Select("name", "email")

	if q.Type != Select {
		t.Error("Select() did not set the correct query type")
	}

	if len(q.Fields) != 2 {
		t.Error("Select() did not add the fields")
	}
}

func TestQuery_InsertInto(t *testing.T) {
	q := NewQuery().InsertInto("users")

	if q.Type != Insert {
		t.Error("InsertInto() did not set the correct query type")
	}

	if q.PrimaryTable.(Ident) != "users" {
		t.Error("InsertInto() did not set the correct primary table")
	}
}

func TestQuery_InsertIgnoreInto(t *testing.T) {
	q := NewQuery().InsertIgnoreInto("users")

	if q.Type != InsertIgnore {
		t.Error("InsertIgnoreInto() did not set the correct query type")
	}

	if q.PrimaryTable.(Ident) != "users" {
		t.Error("InsertIgnoreInto() did not set the correct primary table")
	}
}

func TestQuery_InsertUpdateInto(t *testing.T) {
	q := NewQuery().InsertUpdateInto("users")

	if q.Type != InsertUpdate {
		t.Error("InsertUpdateInto() did not set the correct query type")
	}

	if q.PrimaryTable.(Ident) != "users" {
		t.Error("InsertUpdateInto() did not set the correct primary table")
	}
}

func TestQuery_UpdateTable(t *testing.T) {
	q := NewQuery().Update("users")

	if q.Type != Update {
		t.Error("Update() did not set the correct query type")
	}

	if q.PrimaryTable.(Ident) != "users" {
		t.Error("Update() did not set the correct primary table")
	}
}

func TestQuery_DeleteFrom(t *testing.T) {
	q := NewQuery().DeleteFrom("users")

	if q.Type != Delete {
		t.Error("DeleteFrom() did not set the correct query type")
	}

	if q.PrimaryTable.(Ident) != "users" {
		t.Error("DeleteFrom() did not set the correct primary table")
	}
}

func TestQuery_As(t *testing.T) {
	q := NewQuery().Select("name").From("users").As("u")

	if q.Alias != "u" {
		t.Error("As() did not set the correct alias")
	}
}

func TestQuery_AddField(t *testing.T) {
	q := NewQuery().Select("name")

	q.AddField("email")

	if len(q.Fields) != 2 {
		t.Error("AddField() did not add the field")
	}
}

func TestQuery_ClearFields(t *testing.T) {
	q := NewQuery().Select("name")

	q.ClearFields()

	if len(q.Fields) != 0 {
		t.Error("ClearFields() did not clear the fields")
	}

	if !q.FieldsCleared {
		t.Error("ClearFields() did not set the FieldsCleared flag")
	}
}

func TestQuery_Set(t *testing.T) {
	q := NewQuery().InsertInto("users")

	values := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	q.Set(values)

	if len(q.Values) != 2 {
		t.Error("Set() did not set the values")
	}
}

func TestQuery_LeftJoin(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	joinCondition := ConditionSet{}

	q.LeftJoin("roles", &joinCondition)

	if len(q.Joins) != 1 {
		t.Error("LeftJoin() did not add the join")
	}

	leftJoin := q.Joins[0]

	if leftJoin.JoinType != LeftJoin {
		t.Errorf("LeftJoin() did not set the correct table %s", leftJoin.Table)
	}

	if leftJoin.Table != Ident("roles") {
		t.Errorf("LeftJoin() did not set the correct table %s", leftJoin.Table)
	}

	if !reflect.DeepEqual(leftJoin.Condition, &joinCondition) {
		t.Error("LeftJoin() did not set the correct join condition")
	}
}

func TestQuery_LeftJoinEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.LeftJoinEq("roles", "role_id", "id")

	if len(q.Joins) != 1 {
		t.Error("LeftJoinEq() did not add the join")
	}

	leftJoin := q.Joins[0]

	expectedCondition := Condition()
	expectedCondition.Eq("role_id", LValue("id"))

	if leftJoin.JoinType != LeftJoin {
		t.Errorf("LeftJoin() did not set the correct table %s", leftJoin.Table)
	}

	if !reflect.DeepEqual(leftJoin.Condition, expectedCondition) {
		t.Error("LeftJoinEq() did not set the correct join condition")
	}
}

func TestQuery_InnerJoin(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	joinCondition := Condition()

	q.InnerJoin("roles", joinCondition)

	if len(q.Joins) != 1 {
		t.Error("InnerJoin() did not add the join")
	}

	innerJoin := q.Joins[0]

	if innerJoin.JoinType != InnerJoin {
		t.Errorf("InnerJoin() did not set the correct table %s", innerJoin.Table)
	}

	if innerJoin.Table != Ident("roles") {
		t.Error("InnerJoin() did not set the correct table")
	}

	if !reflect.DeepEqual(innerJoin.Condition, joinCondition) {
		t.Error("InnerJoin() did not set the correct join condition")
	}
}

func TestQuery_InnerJoinEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.InnerJoinEq("roles", "role_id", "id")

	if len(q.Joins) != 1 {
		t.Error("InnerJoinEq() did not add the join")
	}

	innerJoin := q.Joins[0]

	expectedCondition := Condition()
	expectedCondition.Eq("role_id", LValue("id"))

	if innerJoin.JoinType != InnerJoin {
		t.Errorf("InnerJoin() did not set the correct table %s", innerJoin.Table)
	}

	if !reflect.DeepEqual(innerJoin.Condition, expectedCondition) {
		t.Error("InnerJoinEq() did not set the correct join condition")
	}
}

func TestQuery_RightJoin(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	joinCondition := Condition()

	q.RightJoin("roles", joinCondition)

	if len(q.Joins) != 1 {
		t.Error("RightJoin() did not add the join")
	}

	rightJoin := q.Joins[0]

	if rightJoin.JoinType != RightJoin {
		t.Errorf("RightJoin() did not set the correct table %s", rightJoin.Table)
	}

	if rightJoin.Table != Ident("roles") {
		t.Error("RightJoin() did not set the correct table")
	}

	if !reflect.DeepEqual(rightJoin.Condition, joinCondition) {
		t.Error("RightJoin() did not set the correct join condition")
	}
}

func TestQuery_RightJoinEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.RightJoinEq("roles", "role_id", "id")

	if len(q.Joins) != 1 {
		t.Error("RightJoinEq() did not add the join")
	}

	rightJoin := q.Joins[0]

	expectedCondition := Condition()
	expectedCondition.Eq("role_id", LValue("id"))

	if !reflect.DeepEqual(rightJoin.Condition, expectedCondition) {
		t.Error("RightJoinEq() did not set the correct join condition")
	}
}

func TestQuery_Where(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	condition := Condition()

	q.Where(condition)

	if !reflect.DeepEqual(q.WhereCondition.Conditions[0], condition) {
		t.Error("Where() did not set the correct condition")
	}
}

func TestQuery_WhereNot(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	condition := Condition()
	q.WhereNot(condition)

	if !reflect.DeepEqual(q.WhereCondition.Conditions[0], condition) {
		t.Error("WhereNot() did not set the correct condition")
	}

	if q.WhereCondition.Not {
		t.Error("WhereNot() did not set the Not flag")
	}
}

func TestQuery_WhereEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereEq("age", 18)

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereEq() did not add the condition")
	}

	eq := q.WhereCondition.Conditions[0].(Eq)

	if eq.Left != LValue("age") {
		t.Error("WhereEq() did not set the correct left value")
	}

	if eq.Right != RValue(18) {
		t.Error("WhereEq() did not set the correct right value")
	}

	if eq.Not {
		t.Error("WhereEq() should not set the Not flag")
	}
}

func TestQuery_WhereNotEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereNotEq("age", 18)

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereNotEq() did not add the condition")
	}

	eq := q.WhereCondition.Conditions[0].(Eq)

	if eq.Left != LValue("age") {
		t.Error("WhereNotEq() did not set the correct left value")
	}

	if eq.Right != RValue(18) {
		t.Error("WhereNotEq() did not set the correct right value")
	}

	if !eq.Not {
		t.Error("WhereNotEq() did not set the Not flag")
	}
}

func TestQuery_WhereLt(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereLt("age", 18)

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereLt() did not add the condition")
	}

	lt := q.WhereCondition.Conditions[0].(Lt)

	if lt.Left != LValue("age") {
		t.Error("WhereLt() did not set the correct left value")
	}

	if lt.Right != RValue(18) {
		t.Error("WhereLt() did not set the correct right value")
	}

	if lt.Not {
		t.Error("WhereLt() should not set the Not flag")
	}
}

func TestQuery_WhereLtEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereLtEq("age", 18)

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereLtEq() did not add the condition")
	}

	ltEq := q.WhereCondition.Conditions[0].(LtEq)

	if ltEq.Left != LValue("age") {
		t.Error("WhereLtEq() did not set the correct left value")
	}

	if ltEq.Right != RValue(18) {
		t.Error("WhereLtEq() did not set the correct right value")
	}

	if ltEq.Not {
		t.Error("WhereLtEq() should not set the Not flag")
	}
}

func TestQuery_WhereGt(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereGt("age", 18)

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereGt() did not add the condition")
	}

	gt := q.WhereCondition.Conditions[0].(Gt)

	if gt.Left != LValue("age") {
		t.Error("WhereGt() did not set the correct left value")
	}

	if gt.Right != RValue(18) {
		t.Error("WhereGt() did not set the correct right value")
	}

	if gt.Not {
		t.Error("WhereGt() should not set the Not flag")
	}
}

func TestQuery_WhereGtEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereGtEq("age", 18)

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereGtEq() did not add the condition")
	}

	gtEq := q.WhereCondition.Conditions[0].(GtEq)

	if gtEq.Left != LValue("age") {
		t.Error("WhereGtEq() did not set the correct left value")
	}

	if gtEq.Right != RValue(18) {
		t.Error("WhereGtEq() did not set the correct right value")
	}

	if gtEq.Not {
		t.Error("WhereGtEq() should not set the Not flag")
	}
}

func TestQuery_WhereIn(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIn("category", []string{"A", "B", "C"})

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIn() did not add the condition")
	}

	_in := q.WhereCondition.Conditions[0].(In)

	if _in.Left != LValue("category") {
		t.Error("WhereIn() did not set the correct left value")
	}

	if !reflect.DeepEqual(_in.Right, List{String("A"), String("B"), String("C")}) {
		t.Error("WhereIn() did not set the correct right value")
	}

	if _in.Not {
		t.Error("WhereIn() should not set the Not flag")
	}
}

func TestQuery_WhereNotIn(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereNotIn("category", []string{"A", "B", "C"})

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereNotIn() did not add the condition")
	}

	_in := q.WhereCondition.Conditions[0].(In)

	if _in.Left != LValue("category") {
		t.Error("WhereNotIn() did not set the correct left value")
	}

	if !reflect.DeepEqual(_in.Right, List{String("A"), String("B"), String("C")}) {
		t.Error("WhereNotIn() did not set the correct right value")
	}

	if !_in.Not {
		t.Error("WhereNotIn() did not set the Not flag")
	}
}

func TestQuery_WhereIsNull(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIsNull("name")

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIsNull() did not add the condition")
	}

	isNull := q.WhereCondition.Conditions[0].(IsNull)

	if isNull.Value != LValue("name") {
		t.Error("WhereIsNull() did not set the correct value")
	}

	if isNull.Not {
		t.Error("WhereIsNull() should not set the Not flag")
	}
}

func TestQuery_WhereIsNotNull(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIsNotNull("name")

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIsNotNull() did not add the condition")
	}

	isNull := q.WhereCondition.Conditions[0].(IsNull)

	if isNull.Value != LValue("name") {
		t.Error("WhereIsNotNull() did not set the correct value")
	}

	if !isNull.Not {
		t.Error("WhereIsNotNull() did not set the Not flag")
	}
}

func TestQuery_WhereIsFalse(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIsFalse("is_active")

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIsFalse() did not add the condition")
	}

	isFalse := q.WhereCondition.Conditions[0].(IsFalse)

	if isFalse.Value != LValue("is_active") {
		t.Error("WhereIsFalse() did not set the correct value")
	}

	if isFalse.Not {
		t.Error("WhereIsFalse() should not set the Not flag")
	}
}

func TestQuery_WhereIsNotFalse(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIsNotFalse("is_active")

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIsNotFalse() did not add the condition")
	}

	isFalse := q.WhereCondition.Conditions[0].(IsFalse)

	if isFalse.Value != LValue("is_active") {
		t.Error("WhereIsNotFalse() did not set the correct value")
	}

	if !isFalse.Not {
		t.Error("WhereIsNotFalse() did not set the Not flag")
	}
}

func TestQuery_WhereIsTrue(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIsTrue("is_active")

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIsTrue() did not add the condition")
	}

	isTrue := q.WhereCondition.Conditions[0].(IsTrue)

	if isTrue.Value != LValue("is_active") {
		t.Error("WhereIsTrue() did not set the correct value")
	}

	if isTrue.Not {
		t.Error("WhereIsTrue() should not set the Not flag")
	}
}

func TestQuery_WhereIsNotTrue(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.WhereIsNotTrue("is_active")

	if len(q.WhereCondition.Conditions) != 1 {
		t.Error("WhereIsNotTrue() did not add the condition")
	}

	isTrue := q.WhereCondition.Conditions[0].(IsTrue)

	if isTrue.Value != LValue("is_active") {
		t.Error("WhereIsNotTrue() did not set the correct value")
	}

	if !isTrue.Not {
		t.Error("WhereIsNotTrue() did not set the Not flag")
	}
}

func TestQuery_GroupBy(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.GroupBy("name")

	if len(q.GroupBys) != 1 {
		t.Error("GroupBy() did not add the field to the group by clause")
	}
}

func TestQuery_Having(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	condition := ConditionSet{}

	q.Having(&condition)

	if !reflect.DeepEqual(q.HavingCondition.Conditions[0], &condition) {
		t.Error("Having() did not set the correct condition")
	}
}

func TestQuery_HavingNot(t *testing.T) {
	q := NewQuery().Select("name").From("users")
	condition := ConditionSet{}

	q.HavingNot(&condition)

	if !reflect.DeepEqual(q.HavingCondition.Conditions[0], &condition) {
		t.Error("HavingNot() did not set the correct condition")
	}

	if q.HavingCondition.Not {
		t.Error("HavingNot() did not set the Not flag")
	}
}

func TestQuery_HavingEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingEq("age", 18)

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingEq() did not add the condition")
	}

	eq := q.HavingCondition.Conditions[0].(Eq)

	if eq.Left != LValue("age") {
		t.Error("HavingEq() did not set the correct left value")
	}

	if eq.Right != RValue(18) {
		t.Error("HavingEq() did not set the correct right value")
	}

	if eq.Not {
		t.Error("HavingEq() should not set the Not flag")
	}
}

func TestQuery_HavingNotEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingNotEq("age", 18)

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingNotEq() did not add the condition")
	}

	eq := q.HavingCondition.Conditions[0].(Eq)

	if eq.Left != LValue("age") {
		t.Error("HavingNotEq() did not set the correct left value")
	}

	if eq.Right != RValue(18) {
		t.Error("HavingNotEq() did not set the correct right value")
	}

	if !eq.Not {
		t.Error("HavingNotEq() did not set the Not flag")
	}
}

func TestQuery_HavingLt(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingLt("age", 18)

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingLt() did not add the condition")
	}

	lt := q.HavingCondition.Conditions[0].(Lt)

	if lt.Left != LValue("age") {
		t.Error("HavingLt() did not set the correct left value")
	}

	if lt.Right != RValue(18) {
		t.Error("HavingLt() did not set the correct right value")
	}

	if lt.Not {
		t.Error("HavingLt() should not set the Not flag")
	}
}

func TestQuery_HavingLtEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingLtEq("age", 18)

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingLtEq() did not add the condition")
	}

	ltEq := q.HavingCondition.Conditions[0].(LtEq)

	if ltEq.Left != LValue("age") {
		t.Error("HavingLtEq() did not set the correct left value")
	}

	if ltEq.Right != RValue(18) {
		t.Error("HavingLtEq() did not set the correct right value")
	}

	if ltEq.Not {
		t.Error("HavingLtEq() should not set the Not flag")
	}
}

func TestQuery_HavingGt(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingGt("age", 18)

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingGt() did not add the condition")
	}

	gt := q.HavingCondition.Conditions[0].(Gt)

	if gt.Left != LValue("age") {
		t.Error("HavingGt() did not set the correct left value")
	}

	if gt.Right != RValue(18) {
		t.Error("HavingGt() did not set the correct right value")
	}

	if gt.Not {
		t.Error("HavingGt() should not set the Not flag")
	}
}

func TestQuery_HavingGtEq(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingGtEq("age", 18)

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingGtEq() did not add the condition")
	}

	gtEq := q.HavingCondition.Conditions[0].(GtEq)

	if gtEq.Left != LValue("age") {
		t.Error("HavingGtEq() did not set the correct left value")
	}

	if gtEq.Right != RValue(18) {
		t.Error("HavingGtEq() did not set the correct right value")
	}

	if gtEq.Not {
		t.Error("HavingGtEq() should not set the Not flag")
	}
}

func TestQuery_HavingIn(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIn("category", []string{"A", "B", "C"})

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIn() did not add the condition")
	}

	_in := q.HavingCondition.Conditions[0].(In)

	if _in.Left != LValue("category") {
		t.Error("HavingIn() did not set the correct left value")
	}

	if !reflect.DeepEqual(_in.Right, List{String("A"), String("B"), String("C")}) {
		t.Error("HavingIn() did not set the correct right value")
	}

	if _in.Not {
		t.Error("HavingIn() should not set the Not flag")
	}
}

func TestQuery_HavingNotIn(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingNotIn("category", []string{"A", "B", "C"})

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingNotIn() did not add the condition")
	}

	_in := q.HavingCondition.Conditions[0].(In)

	if _in.Left != LValue("category") {
		t.Error("HavingNotIn() did not set the correct left value")
	}

	if !reflect.DeepEqual(_in.Right, List{String("A"), String("B"), String("C")}) {
		t.Error("HavingNotIn() did not set the correct right value")
	}

	if !_in.Not {
		t.Error("HavingNotIn() did not set the Not flag")
	}
}

func TestQuery_HavingIsNull(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIsNull("name")

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIsNull() did not add the condition")
	}

	isNull := q.HavingCondition.Conditions[0].(IsNull)

	if isNull.Value != LValue("name") {
		t.Error("HavingIsNull() did not set the correct value")
	}

	if isNull.Not {
		t.Error("HavingIsNull() should not set the Not flag")
	}
}

func TestQuery_HavingIsNotNull(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIsNotNull("name")

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIsNotNull() did not add the condition")
	}

	isNull := q.HavingCondition.Conditions[0].(IsNull)

	if isNull.Value != LValue("name") {
		t.Error("HavingIsNotNull() did not set the correct value")
	}

	if !isNull.Not {
		t.Error("HavingIsNotNull() did not set the Not flag")
	}
}

func TestQuery_HavingIsFalse(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIsFalse("is_active")

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIsFalse() did not add the condition")
	}

	isFalse := q.HavingCondition.Conditions[0].(IsFalse)

	if isFalse.Value != LValue("is_active") {
		t.Error("HavingIsFalse() did not set the correct value")
	}

	if isFalse.Not {
		t.Error("HavingIsFalse() should not set the Not flag")
	}
}

func TestQuery_HavingIsNotFalse(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIsNotFalse("is_active")

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIsNotFalse() did not add the condition")
	}

	isFalse := q.HavingCondition.Conditions[0].(IsFalse)

	if isFalse.Value != LValue("is_active") {
		t.Error("HavingIsNotFalse() did not set the correct value")
	}

	if !isFalse.Not {
		t.Error("HavingIsNotFalse() did not set the Not flag")
	}
}

func TestQuery_HavingIsTrue(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIsTrue("is_active")

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIsTrue() did not add the condition")
	}

	isTrue := q.HavingCondition.Conditions[0].(IsTrue)

	if isTrue.Value != LValue("is_active") {
		t.Error("HavingIsTrue() did not set the correct value")
	}

	if isTrue.Not {
		t.Error("HavingIsTrue() should not set the Not flag")
	}
}

func TestQuery_HavingIsNotTrue(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.HavingIsNotTrue("is_active")

	if len(q.HavingCondition.Conditions) != 1 {
		t.Error("HavingIsNotTrue() did not add the condition")
	}

	isTrue := q.HavingCondition.Conditions[0].(IsTrue)

	if isTrue.Value != LValue("is_active") {
		t.Error("HavingIsNotTrue() did not set the correct value")
	}

	if !isTrue.Not {
		t.Error("HavingIsNotTrue() did not set the Not flag")
	}
}

func TestQuery_OrderBy(t *testing.T) {
	q := NewQuery().Select("name", "age").From("users")

	q.OrderBy("name", Asc)
	q.OrderBy("age", Desc)

	if len(q.OrderBys) != 2 {
		t.Error("OrderBy() did not add the order by clause")
	}
}

func TestQuery_ClearOrderBys(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.ClearOrderBys()

	if len(q.OrderBys) != 0 {
		t.Error("ClearOrderBys() did not clear the order by clause")
	}

	if !q.OrderBysCleared {
		t.Error("ClearOrderBys() did not set the OrderBysCleared flag")
	}
}

func TestQuery_Limit(t *testing.T) {
	q := NewQuery().Select("name").From("users")

	q.Limit(10, 10)

	if q.Offset.Start != 10 {
		t.Error("Limit() did not set the correct start offset")
	}

	if q.Offset.Limit != 10 {
		t.Error("Limit() did not set the correct limit value")
	}
}

func TestQuery_Union(t *testing.T) {
	q1 := NewQuery().Select("name").From("users")
	q2 := NewQuery().Select("name").From("customers")

	q1.Union(q2)

	if len(q1.Unions) != 1 {
		t.Error("Union() did not add the query to the unions")
	}

	union := q1.Unions[0]

	if !reflect.DeepEqual(union.Query, q2) {
		t.Error("Union() did not set the correct query in the union")
	}

	if union.UnionType != UnionDefault {
		t.Error("Union() did not set the correct union type")
	}
}

func TestQuery_UnionAll(t *testing.T) {
	q1 := NewQuery().Select("name").From("users")
	q2 := NewQuery().Select("name").From("customers")

	q1.UnionAll(q2)

	if len(q1.Unions) != 1 {
		t.Error("UnionAll() did not add the query to the unions")
	}

	union := q1.Unions[0]

	if !reflect.DeepEqual(union.Query, q2) {
		t.Error("UnionAll() did not set the correct query in the union")
	}

	if union.UnionType != UnionAll {
		t.Error("UnionAll() did not set the correct union type")
	}
}
