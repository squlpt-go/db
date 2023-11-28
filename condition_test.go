package db

import (
	"reflect"
	"testing"
)

func TestConditionSet_Condition(t *testing.T) {
	c := ConditionSet{}
	sub := ConditionSet{}

	c.Condition(&sub)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.Condition() did not add the sub condition")
	}

	if !reflect.DeepEqual(c.Conditions[0], &sub) {
		t.Errorf("ConditionSet.Condition() did not add the correct sub condition")
	}
}

func TestConditionSet_Eq(t *testing.T) {
	c := ConditionSet{}
	left := Ident("name")
	right := String("John")

	c.Eq(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.Eq() did not add the condition")
	}

	eq := c.Conditions[0].(Eq)

	if !reflect.DeepEqual(eq.Left, left) {
		t.Errorf("ConditionSet.Eq() did not set the correct left value")
	}

	if !reflect.DeepEqual(eq.Right, right) {
		t.Errorf("ConditionSet.Eq() did not set the correct right value")
	}

	if eq.Not {
		t.Errorf("ConditionSet.Eq() should not set the Not flag")
	}
}

func TestConditionSet_NotEq(t *testing.T) {
	c := ConditionSet{}
	left := Ident("name")
	right := String("John")

	c.NotEq(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.NotEq() did not add the condition")
	}

	eq := c.Conditions[0].(Eq)

	if !reflect.DeepEqual(eq.Left, left) {
		t.Errorf("ConditionSet.NotEq() did not set the correct left value")
	}

	if !reflect.DeepEqual(eq.Right, right) {
		t.Errorf("ConditionSet.NotEq() did not set the correct right value")
	}

	if !eq.Not {
		t.Errorf("ConditionSet.NotEq() did not set the Not flag")
	}
}

func TestConditionSet_Lt(t *testing.T) {
	c := ConditionSet{}
	left := Ident("age")
	right := Int(18)

	c.Lt(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.Lt() did not add the condition")
	}

	lt := c.Conditions[0].(Lt)

	if !reflect.DeepEqual(lt.Left, left) {
		t.Errorf("ConditionSet.Lt() did not set the correct left value")
	}

	if !reflect.DeepEqual(lt.Right, right) {
		t.Errorf("ConditionSet.Lt() did not set the correct right value")
	}

	if lt.Not {
		t.Errorf("ConditionSet.Lt() should not set the Not flag")
	}
}

func TestConditionSet_LtEq(t *testing.T) {
	c := ConditionSet{}
	left := Ident("age")
	right := Int(18)

	c.LtEq(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.LtEq() did not add the condition")
	}

	ltEq := c.Conditions[0].(LtEq)

	if !reflect.DeepEqual(ltEq.Left, left) {
		t.Errorf("ConditionSet.LtEq() did not set the correct left value")
	}

	if !reflect.DeepEqual(ltEq.Right, right) {
		t.Errorf("ConditionSet.LtEq() did not set the correct right value")
	}

	if ltEq.Not {
		t.Errorf("ConditionSet.LtEq() should not set the Not flag")
	}
}

func TestConditionSet_Gt(t *testing.T) {
	c := ConditionSet{}
	left := Ident("age")
	right := Int(18)

	c.Gt(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.Gt() did not add the condition")
	}

	gt := c.Conditions[0].(Gt)

	if !reflect.DeepEqual(gt.Left, left) {
		t.Errorf("ConditionSet.Gt() did not set the correct left value")
	}

	if !reflect.DeepEqual(gt.Right, right) {
		t.Errorf("ConditionSet.Gt() did not set the correct right value")
	}

	if gt.Not {
		t.Errorf("ConditionSet.Gt() should not set the Not flag")
	}
}

func TestConditionSet_GtEq(t *testing.T) {
	c := ConditionSet{}
	left := Ident("age")
	right := Int(18)

	c.GtEq(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.GtEq() did not add the condition")
	}

	gtEq := c.Conditions[0].(GtEq)

	if !reflect.DeepEqual(gtEq.Left, left) {
		t.Errorf("ConditionSet.GtEq() did not set the correct left value")
	}

	if !reflect.DeepEqual(gtEq.Right, right) {
		t.Errorf("ConditionSet.GtEq() did not set the correct right value")
	}

	if gtEq.Not {
		t.Errorf("ConditionSet.GtEq() should not set the Not flag")
	}
}

func TestConditionSet_In(t *testing.T) {
	c := ConditionSet{}
	left := Ident("category")
	right := []string{"A", "B", "C"}
	expected := List{String("A"), String("B"), String("C")}

	c.In(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.In() did not add the condition")
	}

	_in := c.Conditions[0].(In)

	if !reflect.DeepEqual(_in.Left, left) {
		t.Errorf("ConditionSet.In() did not set the correct left value")
	}

	if !reflect.DeepEqual(_in.Right, expected) {
		t.Errorf("ConditionSet.In() did not set the correct right value")
	}

	if _in.Not {
		t.Errorf("ConditionSet.In() should not set the Not flag")
	}
}

func TestConditionSet_NotIn(t *testing.T) {
	c := ConditionSet{}
	left := Ident("category")
	right := []string{"A", "B", "C"}
	expected := List{String("A"), String("B"), String("C")}

	c.NotIn(left, right)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.NotIn() did not add the condition")
	}

	_in := c.Conditions[0].(In)

	if !reflect.DeepEqual(_in.Left, left) {
		t.Errorf("ConditionSet.NotIn() did not set the correct left value")
	}

	if !reflect.DeepEqual(_in.Right, expected) {
		t.Errorf("ConditionSet.NotIn() did not set the correct right value")
	}

	if !_in.Not {
		t.Errorf("ConditionSet.NotIn() did not set the Not flag")
	}
}

func TestConditionSet_IsNull(t *testing.T) {
	c := ConditionSet{}
	value := Ident("name")

	c.IsNull(value)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.IsNull() did not add the condition")
	}

	isNull := c.Conditions[0].(IsNull)

	if !reflect.DeepEqual(isNull.Value, value) {
		t.Errorf("ConditionSet.IsNull() did not set the correct value")
	}

	if isNull.Not {
		t.Errorf("ConditionSet.IsNull() should not set the Not flag")
	}
}

func TestConditionSet_IsNotNull(t *testing.T) {
	c := ConditionSet{}
	value := Ident("name")

	c.IsNotNull(value)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.IsNotNull() did not add the condition")
	}

	isNull := c.Conditions[0].(IsNull)

	if !reflect.DeepEqual(isNull.Value, value) {
		t.Errorf("ConditionSet.IsNotNull() did not set the correct value")
	}

	if !isNull.Not {
		t.Errorf("ConditionSet.IsNotNull() did not set the Not flag")
	}
}

func TestConditionSet_IsFalse(t *testing.T) {
	c := ConditionSet{}
	value := Ident("is_active")

	c.IsFalse(value)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.IsFalse() did not add the condition")
	}

	isFalse := c.Conditions[0].(IsFalse)

	if !reflect.DeepEqual(isFalse.Value, value) {
		t.Errorf("ConditionSet.IsFalse() did not set the correct value")
	}

	if isFalse.Not {
		t.Errorf("ConditionSet.IsFalse() should not set the Not flag")
	}
}

func TestConditionSet_IsNotFalse(t *testing.T) {
	c := ConditionSet{}
	value := Ident("is_active")

	c.IsNotFalse(value)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.IsNotFalse() did not add the condition")
	}

	isFalse := c.Conditions[0].(IsFalse)

	if !reflect.DeepEqual(isFalse.Value, value) {
		t.Errorf("ConditionSet.IsNotFalse() did not set the correct value")
	}

	if !isFalse.Not {
		t.Errorf("ConditionSet.IsNotFalse() did not set the Not flag")
	}
}

func TestConditionSet_IsTrue(t *testing.T) {
	c := ConditionSet{}
	value := Ident("is_active")

	c.IsTrue(value)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.IsTrue() did not add the condition")
	}

	isTrue := c.Conditions[0].(IsTrue)

	if !reflect.DeepEqual(isTrue.Value, value) {
		t.Errorf("ConditionSet.IsTrue() did not set the correct value")
	}

	if isTrue.Not {
		t.Errorf("ConditionSet.IsTrue() should not set the Not flag")
	}
}

func TestConditionSet_IsNotTrue(t *testing.T) {
	c := ConditionSet{}
	value := Ident("is_active")

	c.IsNotTrue(value)

	if len(c.Conditions) != 1 {
		t.Errorf("ConditionSet.IsNotTrue() did not add the condition")
	}

	isTrue := c.Conditions[0].(IsTrue)

	if !reflect.DeepEqual(isTrue.Value, value) {
		t.Errorf("ConditionSet.IsNotTrue() did not set the correct value")
	}

	if !isTrue.Not {
		t.Errorf("ConditionSet.IsNotTrue() did not set the Not flag")
	}
}
