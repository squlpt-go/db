package db

import (
	"database/sql"
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

func Validate(db *sql.DB, rules ...RuleDefT) {
	validate(db, validationRules, rules...)
}

func Rule[T IEntity](callback func(*T) error) RuleDef[T] {
	return RuleDef[T]{
		rule: func(entity any) error {
			return callback(entity.(*T))
		},
	}
}

func Filter[T IEntity](callback func(map[string]any)) RuleDef[T] {
	return RuleDef[T]{
		filter: func(fields map[string]any) {
			callback(fields)
		},
	}
}

func Required[T IEntity](fields ...string) RuleDef[T] {
	return Rule[T](func(entity *T) error {
		efs := entityToMap(entity, false, false)
		for key, field := range efs {
			if slices.Contains(fields, key) {
				if field == nil || reflect.ValueOf(field).IsZero() {
					return fmt.Errorf("%w: %s", ErrRequired, key)
				}
			}
		}
		return nil
	})
}

func Matches[T IEntity](pattern string, fields ...string) RuleDef[T] {
	re := regexp.MustCompile(pattern)

	return Rule[T](func(entity *T) error {
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
	})
}

func DoValidateInsert[T IEntity](db *sql.DB, entity *T) error {
	return doValidate(db, entity, Insert)
}

func DoValidateUpdate[T IEntity](db *sql.DB, entity *T) error {
	return doValidate(db, entity, Update)
}

func DoValidateChildren[T IEntity](db *sql.DB, entity *T) error {
	return doValidate(db, entity, queryTypeChildren)
}

func DoFilterInsert[T IEntity](db *sql.DB, fields map[string]any) {
	doFilter[T](db, fields, Insert)
}

func DoFilterUpdate[T IEntity](db *sql.DB, fields map[string]any) {
	doFilter[T](db, fields, Update)
}

func DoFilterChildren[T IEntity](db *sql.DB, fields map[string]any) {
	doFilter[T](db, fields, queryTypeChildren)
}

type RuleDefT interface {
	Validate(any) error
	Filter(map[string]any)
	Error() error
	WithErr(error) RuleDefT
	typeId() typeId
	QueryType() QueryType
	ForInsert() RuleDefT
	ForUpdate() RuleDefT
	ForChildren() RuleDefT
}

type RuleDef[T IEntity] struct {
	rule      func(any) error
	filter    func(map[string]any)
	err       error
	queryType QueryType
}

func (rd RuleDef[T]) Validate(entity any) error {
	if rd.rule != nil {
		return rd.rule(entity)
	}

	return nil
}

func (rd RuleDef[T]) Filter(flat map[string]any) {
	if rd.filter != nil {
		rd.filter(flat)
	}
}

func (rd RuleDef[T]) Error() error {
	return rd.err
}

func (rd RuleDef[T]) WithErr(err error) RuleDefT {
	return RuleDef[T]{
		rule: rd.rule,
		err:  err,
	}
}

func (rd RuleDef[T]) typeId() typeId {
	return tId[T]()
}

func (rd RuleDef[T]) QueryType() QueryType {
	return rd.queryType
}

func (rd RuleDef[T]) ForInsert() RuleDefT {
	return RuleDef[T]{
		rule:      rd.rule,
		err:       rd.err,
		queryType: Insert,
	}
}

func (rd RuleDef[T]) ForUpdate() RuleDefT {
	return RuleDef[T]{
		rule:      rd.rule,
		err:       rd.err,
		queryType: Update,
	}
}

func (rd RuleDef[T]) ForChildren() RuleDefT {
	return RuleDef[T]{
		rule:      rd.rule,
		err:       rd.err,
		queryType: queryTypeChildren,
	}
}

type ruleSet map[*sql.DB]map[typeId][]RuleDefT

var validationRules = make(ruleSet)

func validate(db *sql.DB, m ruleSet, rules ...RuleDefT) {
	if _, ok := m[db]; !ok {
		m[db] = make(map[typeId][]RuleDefT)
	}

	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		tid := rule.typeId()
		m[db][tid] = append(m[db][tid], rule)
	}
}

const queryTypeAll = QueryType("")
const queryTypeChildren = QueryType("CHILDREN")

func doValidate[T IEntity](db *sql.DB, entity *T, queryType QueryType) error {
	tid := tId[T]()

	if rules, ok := validationRules[db][tid]; ok {
		for i := 0; i < len(rules); i++ {
			rule := rules[i]
			if rule.QueryType() == queryType || rule.QueryType() == queryTypeAll {
				err := rule.Validate(entity)

				if err != nil {
					if rule.Error() != nil {
						return rule.Error()
					} else {
						return err
					}
				}
			}
		}
	}

	return nil
}

func doFilter[T IEntity](db *sql.DB, flat map[string]any, queryType QueryType) error {
	tid := tId[T]()

	if rules, ok := validationRules[db][tid]; ok {
		for i := 0; i < len(rules); i++ {
			rule := rules[i]
			if rule.QueryType() == queryType || rule.QueryType() == queryTypeAll {
				rule.Filter(flat)
			}
		}
	}

	return nil
}
