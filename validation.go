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

func Rule[T IEntity](callback func(*T) error) ruleDef[T] {
	return ruleDef[T]{
		rule: func(entity any) error {
			return callback(entity.(*T))
		},
	}
}

func Required[T IEntity](fields ...string) ruleDef[T] {
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

func Matches[T IEntity](pattern string, fields ...string) ruleDef[T] {
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

type RuleDefT interface {
	Validate(any) error
	Error() error
	WithErr(error) RuleDefT
	typeId() typeId
	QueryType() QueryType
	ForInsert() RuleDefT
	ForUpdate() RuleDefT
	ForChildren() RuleDefT
}

type ruleDef[T IEntity] struct {
	rule      func(any) error
	err       error
	queryType QueryType
}

func (rd ruleDef[T]) Validate(entity any) error {
	return rd.rule(entity)
}

func (rd ruleDef[T]) Error() error {
	return rd.err
}

func (rd ruleDef[T]) WithErr(err error) RuleDefT {
	return ruleDef[T]{
		rule: rd.rule,
		err:  err,
	}
}

func (rd ruleDef[T]) typeId() typeId {
	return tId[T]()
}

func (rd ruleDef[T]) QueryType() QueryType {
	return rd.queryType
}

func (rd ruleDef[T]) ForInsert() RuleDefT {
	return ruleDef[T]{
		rule:      rd.rule,
		err:       rd.err,
		queryType: Insert,
	}
}

func (rd ruleDef[T]) ForUpdate() RuleDefT {
	return ruleDef[T]{
		rule:      rd.rule,
		err:       rd.err,
		queryType: Update,
	}
}

func (rd ruleDef[T]) ForChildren() RuleDefT {
	return ruleDef[T]{
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
