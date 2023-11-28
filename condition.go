package db

type Conj string
type Neg bool

const (
	ConjAnd = Conj("AND")
	ConjOr  = Conj("OR")
)

type Eq struct {
	Left  Value
	Right Value
	Not   Neg
}

type Gt struct {
	Left  Value
	Right Value
	Not   Neg
}

type GtEq struct {
	Left  Value
	Right Value
	Not   Neg
}

type Lt struct {
	Left  Value
	Right Value
	Not   Neg
}

type LtEq struct {
	Left  Value
	Right Value
	Not   Neg
}

type In struct {
	Left  Value
	Right Value
	Not   Neg
}

type IsNull struct {
	Value Value
	Not   Neg
}

type IsTrue struct {
	Value Value
	Not   Neg
}

type IsFalse struct {
	Value Value
	Not   Neg
}

type Like struct {
	Left  Value
	Right Value
	Not   Neg
}

type ConditionSet struct {
	Not        Neg
	Conj       Conj
	Conditions []any
}

func Condition() *ConditionSet {
	return &ConditionSet{
		Not:        false,
		Conj:       ConjAnd,
		Conditions: make([]any, 0),
	}
}

func Or() *ConditionSet {
	return &ConditionSet{
		Not:        false,
		Conj:       ConjOr,
		Conditions: make([]any, 0),
	}
}

func (c *ConditionSet) Like(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		Like{Left: LValue(left), Right: RValue(right)},
	)
	return c
}

func (c *ConditionSet) Eq(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		Eq{Left: LValue(left), Right: RValue(right)},
	)

	return c
}

func (c *ConditionSet) NotEq(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		Eq{Left: LValue(left), Right: RValue(right), Not: true},
	)

	return c
}

func (c *ConditionSet) Lt(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		Lt{Left: LValue(left), Right: RValue(right)},
	)

	return c
}

func (c *ConditionSet) LtEq(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		LtEq{Left: LValue(left), Right: RValue(right)},
	)

	return c
}

func (c *ConditionSet) Gt(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		Gt{Left: LValue(left), Right: RValue(right)},
	)

	return c
}

func (c *ConditionSet) GtEq(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		GtEq{Left: LValue(left), Right: RValue(right)},
	)

	return c
}

func (c *ConditionSet) In(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		In{Left: LValue(left), Right: InValue(right)},
	)

	return c
}

func (c *ConditionSet) NotIn(left any, right any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		In{Left: LValue(left), Right: InValue(right), Not: true},
	)

	return c
}

func (c *ConditionSet) IsNull(value any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		IsNull{Value: LValue(value)},
	)

	return c
}

func (c *ConditionSet) IsNotNull(value any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		IsNull{Value: LValue(value), Not: true},
	)

	return c
}

func (c *ConditionSet) IsFalse(value any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		IsFalse{Value: LValue(value)},
	)

	return c
}

func (c *ConditionSet) IsNotFalse(value any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		IsFalse{Value: LValue(value), Not: true},
	)

	return c
}

func (c *ConditionSet) IsTrue(value any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		IsTrue{Value: LValue(value)},
	)

	return c
}

func (c *ConditionSet) IsNotTrue(value any) *ConditionSet {
	c.Conditions = append(
		c.Conditions,
		IsTrue{Value: LValue(value), Not: true},
	)

	return c
}

func (c *ConditionSet) Condition(sub *ConditionSet) *ConditionSet {
	c.Conditions = append(c.Conditions, sub)
	return c
}
