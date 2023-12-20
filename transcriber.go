package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	RegisterTranscriber(&mysql.MySQLDriver{}, MySQLTranscriber{UsePlaceholders: true})
}

const clauseSeparator = " "

type Transcribeable interface {
	Transcribe(db *sql.DB) (string, []any, error)
}

type Transcriber interface {
	Transcribe(*QueryBuilder) (string, []any, error)
}

type driverID string

func getDriverID(d driver.Driver) driverID {
	return driverID(reflect.TypeOf(d).String())
}

var transcribers map[driverID]Transcriber = make(map[driverID]Transcriber)

func getTranscriber(d driver.Driver) Transcriber {
	id := getDriverID(d)
	t, ok := transcribers[id]

	if !ok {
		panic("No transcriber defined for driver ID: " + string(id))
	}

	return t
}

func RegisterTranscriber(d driver.Driver, t Transcriber) {
	transcribers[getDriverID(d)] = t
}

type MySQLTranscriber struct {
	UsePlaceholders bool
}

func (t MySQLTranscriber) Transcribe(q *QueryBuilder) (string, []any, error) {
	sql := ""
	args := make([]any, 0)

	switch q.Type {
	case Select:
		return t.processSelectQuery(q)
	case Update:
		return t.processUpdateQuery(q)
	case Insert, InsertUpdate, InsertIgnore:
		return t.processInsertQuery(q)
	case Delete:
		return t.processDeleteQuery(q)
	default:
		return sql, args, errors.New("invalid query type")
	}
}

func (t MySQLTranscriber) processSelectQuery(q *QueryBuilder) (string, []any, error) {
	lines := make([]string, 0)
	args := make([]any, 0)

	err := t.fields(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.from(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.joins(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.where(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.group(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.having(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.order(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.limit(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.union(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	return strings.Join(lines, clauseSeparator), args, nil
}

func (t MySQLTranscriber) processUpdateQuery(q *QueryBuilder) (string, []any, error) {
	lines := make([]string, 0)
	args := make([]any, 0)

	err := t.update(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.joins(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.set(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.where(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.group(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.having(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.order(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.limit(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	return strings.Join(lines, clauseSeparator), args, nil
}

func (t MySQLTranscriber) processDeleteQuery(q *QueryBuilder) (string, []any, error) {
	lines := make([]string, 0)
	args := make([]any, 0)

	err := t.delete(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.joins(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.set(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.where(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.group(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.having(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.order(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.limit(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	return strings.Join(lines, clauseSeparator), args, nil
}

func (t MySQLTranscriber) processInsertQuery(q *QueryBuilder) (string, []any, error) {
	lines := make([]string, 0)
	args := make([]any, 0)

	err := t.insert(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.joins(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.set(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	err = t.insertSuffix(q, &lines, &args)
	if err != nil {
		return "", nil, err
	}

	return strings.Join(lines, clauseSeparator), args, nil
}

func (t MySQLTranscriber) insert(q *QueryBuilder, lines *[]string, args *[]any) error {
	s, a, e := t.processValue(q.PrimaryTable)
	if e != nil {
		return e
	}
	switch q.Type {
	case Insert:
		*lines = append(*lines, "INSERT INTO "+s)
	case InsertUpdate:
		*lines = append(*lines, "INSERT INTO "+s)
	case InsertIgnore:
		*lines = append(*lines, "INSERT IGNORE INTO "+s)
	default:
		panic("Unreachable")
	}
	*args = append(*args, a...)
	return nil
}

func (t MySQLTranscriber) insertSuffix(q *QueryBuilder, lines *[]string, args *[]any) error {
	if q.Type == InsertUpdate {
		ss, sa, se := t.processSet(q.Values)
		if se != nil {
			return se
		}

		*lines = append(*lines, "ON DUPLICATE KEY UPDATE "+ss)
		*args = append(*args, sa...)
	}
	return nil
}

func (t MySQLTranscriber) update(q *QueryBuilder, lines *[]string, args *[]any) error {
	s, a, e := t.processValue(q.PrimaryTable)
	if e != nil {
		return e
	}
	*lines = append(*lines, "UPDATE "+s)
	*args = append(*args, a...)
	return nil
}

func (t MySQLTranscriber) delete(q *QueryBuilder, lines *[]string, args *[]any) error {
	ts, ta, te := t.processValue(q.PrimaryTable)
	if te != nil {
		return te
	}
	if len(q.Fields) > 0 {
		fs, fa, fe := t.processValue(q.Fields)
		if fe != nil {
			return fe
		}
		*lines = append(*lines, "DELETE "+fs+" FROM "+ts)
		*args = append(*args, fa...)
		*args = append(*args, ta...)
	} else {
		*lines = append(*lines, "DELETE FROM "+ts)
		*args = append(*args, ta...)
	}
	return nil
}

func (t MySQLTranscriber) fields(q *QueryBuilder, lines *[]string, args *[]any) error {
	s, a, err := t.processValue(q.Fields)
	if err != nil {
		return err
	}
	*lines = append(*lines, "SELECT "+s)
	*args = append(*args, a...)
	return nil
}

func (t MySQLTranscriber) from(q *QueryBuilder, lines *[]string, args *[]any) error {
	s, a, err := t.processValue(q.PrimaryTable)
	if err != nil {
		return err
	}
	*lines = append(*lines, "FROM "+s)
	*args = append(*args, a...)
	return nil
}

func (t MySQLTranscriber) joins(q *QueryBuilder, lines *[]string, args *[]any) error {
	if len(q.Joins) > 0 {
		s, a, err := t.processJoins(q.Joins)
		if err != nil {
			return err
		}
		*lines = append(*lines, s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) set(q *QueryBuilder, lines *[]string, args *[]any) error {
	if len(q.Values) > 0 {
		s, a, err := t.processSet(q.Values)
		if err != nil {
			return err
		}
		*lines = append(*lines, "SET "+s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) where(q *QueryBuilder, lines *[]string, args *[]any) error {
	if len(q.WhereCondition.Conditions) > 0 {
		s, a, err := t.processCondition(q.WhereCondition)
		if err != nil {
			return err
		}
		*lines = append(*lines, "WHERE "+s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) group(q *QueryBuilder, lines *[]string, args *[]any) error {
	if len(q.GroupBys) > 0 {
		s, a, err := t.processValue(q.GroupBys)
		if err != nil {
			return err
		}
		*lines = append(*lines, "GROUP BY "+s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) having(q *QueryBuilder, lines *[]string, args *[]any) error {
	if len(q.HavingCondition.Conditions) > 0 {
		s, a, err := t.processCondition(q.HavingCondition)
		if err != nil {
			return err
		}
		*lines = append(*lines, "HAVING "+s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) order(q *QueryBuilder, lines *[]string, args *[]any) error {
	if len(q.OrderBys) > 0 {
		s, a, err := t.processOrderBys(q.OrderBys)
		if err != nil {
			return err
		}
		*lines = append(*lines, "ORDER BY "+s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) limit(q *QueryBuilder, lines *[]string, args *[]any) error {
	s, a, err := t.processOffset(q.Offset)
	if err != nil {
		return err
	}
	if s != "" {
		*lines = append(*lines, "LIMIT "+s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) union(q *QueryBuilder, lines *[]string, args *[]any) error {
	s, a, err := t.processUnions(q.Unions)
	if err != nil {
		return err
	}
	if s != "" {
		*lines = append(*lines, s)
		*args = append(*args, a...)
	}
	return nil
}

func (t MySQLTranscriber) processCondition(condition *ConditionSet) (string, []any, error) {
	if len(condition.Conditions) == 0 {
		if condition.Not {
			return "FALSE", []any{}, nil
		} else {
			return "TRUE", []any{}, nil
		}
	}

	cs := make([]string, 0)
	as := make([]any, 0)

	for _, c := range condition.Conditions {
		switch c := c.(type) {
		case Eq:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			rs, ra, re := t.processValue(RValue(c.Right))
			if re != nil {
				return "", nil, re
			}
			if c.Not {
				cs = append(cs, ls+" != "+rs)
			} else {
				cs = append(cs, ls+" = "+rs)
			}
			as = append(as, la...)
			as = append(as, ra...)
		case Gt:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			rs, ra, re := t.processValue(RValue(c.Right))
			if re != nil {
				return "", nil, re
			}
			if c.Not {
				cs = append(cs, ls+" <= "+rs)
			} else {
				cs = append(cs, ls+" > "+rs)
			}
			as = append(as, la...)
			as = append(as, ra...)
		case GtEq:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			rs, ra, re := t.processValue(RValue(c.Right))
			if re != nil {
				return "", nil, re
			}
			if c.Not {
				cs = append(cs, ls+" < "+rs)
			} else {
				cs = append(cs, ls+" >= "+rs)
			}
			as = append(as, la...)
			as = append(as, ra...)
		case Lt:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			rs, ra, re := t.processValue(RValue(c.Right))
			if re != nil {
				return "", nil, re
			}
			if c.Not {
				cs = append(cs, ls+" >= "+rs)
			} else {
				cs = append(cs, ls+" < "+rs)
			}
			as = append(as, la...)
			as = append(as, ra...)
		case LtEq:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			rs, ra, re := t.processValue(RValue(c.Right))
			if re != nil {
				return "", nil, re
			}
			if c.Not {
				cs = append(cs, ls+" > "+rs)
			} else {
				cs = append(cs, ls+" <= "+rs)
			}
			as = append(as, la...)
			as = append(as, ra...)
		case In:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			l := c.Right.(List)
			if len(l) > 0 {
				rs, ra, re := t.processValue(RValue(c.Right))
				if re != nil {
					return "", nil, re
				}
				if c.Not {
					cs = append(cs, ls+" NOT IN("+rs+")")
				} else {
					cs = append(cs, ls+" IN("+rs+")")
				}
				as = append(as, la...)
				as = append(as, ra...)
			} else {
				if c.Not {
					cs = append(cs, "TRUE")
				} else {
					cs = append(cs, "FALSE")
				}
			}
		case IsTrue:
			vs, va, ve := t.processValue(LValue(c.Value))
			if ve != nil {
				return "", nil, ve
			}
			if c.Not {
				cs = append(cs, vs+" IS NOT TRUE")
			} else {
				cs = append(cs, vs+" IS TRUE")
			}
			as = append(as, va...)
		case IsFalse:
			vs, va, ve := t.processValue(LValue(c.Value))
			if ve != nil {
				return "", nil, ve
			}
			if c.Not {
				cs = append(cs, vs+" IS NOT FALSE")
			} else {
				cs = append(cs, vs+" IS FALSE")
			}
			as = append(as, va...)
		case IsNull:
			vs, va, ve := t.processValue(LValue(c.Value))
			if ve != nil {
				return "", nil, ve
			}
			if c.Not {
				cs = append(cs, vs+" IS NOT NULL")
			} else {
				cs = append(cs, vs+" IS NULL")
			}
			as = append(as, va...)
		case Like:
			ls, la, le := t.processValue(LValue(c.Left))
			if le != nil {
				return "", nil, le
			}
			rs, ra, re := t.processValue(RValue(c.Right))
			if re != nil {
				return "", nil, re
			}
			if c.Not {
				cs = append(cs, ls+" NOT LIKE "+rs)
			} else {
				cs = append(cs, ls+" LIKE "+rs)
			}
			as = append(as, la...)
			as = append(as, ra...)
		case *ConditionSet:
			vs, va, ve := t.processCondition(c)
			if ve != nil {
				return "", nil, ve
			}
			if c.Not {
				cs = append(cs, "NOT ("+vs+")")
			} else {
				cs = append(cs, "("+vs+")")
			}
			as = append(as, va...)
		default:
			panic("Invalid condition type " + fmt.Sprintf("%T", c))
		}
	}

	var sql string

	switch condition.Conj {
	case ConjAnd:
		sql = strings.Join(cs, " AND ")
	case ConjOr:
		sql = strings.Join(cs, " OR ")
	default:
		return "", nil, errors.New("Invalid conjunction '" + string(condition.Conj) + "'")
	}

	return sql, as, nil
}

func addSlashes(str string) string {
	var tmpRune []rune
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}

func (t MySQLTranscriber) processValue(value Value) (string, []any, error) {
	switch val := value.(type) {
	case RawQuery:
		return val.Query, val.Args, nil
	case Ident:
		return string(val), []any{}, nil
	case String:
		if t.UsePlaceholders {
			return "?", []any{string(val)}, nil
		} else {
			return "'" + addSlashes(string(val)) + "'", []any{}, nil
		}
	case Int:
		if t.UsePlaceholders {
			return "?", []any{int(val)}, nil
		} else {
			return strconv.Itoa(int(val)), []any{}, nil
		}
	case Float:
		if t.UsePlaceholders {
			return "?", []any{float64(val)}, nil
		} else {
			return fmt.Sprintf("%f", float64(val)), []any{}, nil
		}
	case Time:
		if t.UsePlaceholders {
			return "?", []any{time.Time(val)}, nil
		} else {
			return fmt.Sprintf("%s", time.Time(val).Format("2006-01-02 15:04:05")), []any{}, nil
		}
	case *QueryBuilder:
		q := value.(*QueryBuilder)
		s, a, err := t.Transcribe(q)
		if err != nil {
			return "", nil, err
		}
		s = "(" + normalizeSql(s) + ")"

		if q.Alias != "" {
			s += " AS " + string(q.Alias)
		}

		return s, a, nil
	case List:
		sqls := make([]string, 0)
		args := make([]any, 0)

		for _, v := range value.(List) {
			vSql, vArgs, err := t.processValue(v)

			if err != nil {
				return "", nil, err
			}

			sqls = append(sqls, vSql)
			args = append(args, vArgs...)
		}

		return strings.Join(sqls, ", "), args, nil
	case Null:
		return "NULL", nil, nil
	}

	panic("Unreachable")
}

func (t MySQLTranscriber) processJoins(joins []Join) (string, []any, error) {
	sqls := make([]string, 0)
	args := make([]any, 0)

	for _, j := range joins {
		ts, ta, te := t.processValue(j.Table)
		if te != nil {
			return "", nil, te
		}

		cs, ca, ce := t.processCondition(j.Condition)
		if ce != nil {
			return "", nil, ce
		}

		sqls = append(sqls, string(j.JoinType)+" "+ts+" ON "+cs)
		args = append(args, ta...)
		args = append(args, ca...)
	}

	return strings.Join(sqls, clauseSeparator), args, nil
}

func (t MySQLTranscriber) processOrderBys(orders []Order) (string, []any, error) {
	sqls := make([]string, 0)
	args := make([]any, 0)

	for _, o := range orders {
		ts, ta, te := t.processValue(o.Field)
		if te != nil {
			return "", nil, te
		}

		sqls = append(sqls, ts+" "+string(o.Ord))
		args = append(args, ta...)
	}

	return strings.Join(sqls, ", "), args, nil
}

func (t MySQLTranscriber) processOffset(offset Offset) (string, []any, error) {
	if offset.Start != 0 {
		return fmt.Sprintf("%d, %d", offset.Start, offset.Limit), []any{}, nil
	} else if offset.Limit != 0 {
		return fmt.Sprintf("%d", offset.Limit), []any{}, nil
	}

	return "", []any{}, nil
}

func (t MySQLTranscriber) processUnions(unions []Union) (string, []any, error) {
	sqls := make([]string, 0)
	args := make([]any, 0)

	for _, u := range unions {
		ts, ta, te := t.Transcribe(u.Query)
		if te != nil {
			return "", nil, te
		}

		sqls = append(sqls, string(u.UnionType)+clauseSeparator+ts)
		args = append(args, ta...)
	}

	return strings.Join(sqls, clauseSeparator), args, nil
}

func (t MySQLTranscriber) processSet(values map[string]any) (string, []any, error) {
	sqls := make([]string, 0)
	args := make([]any, 0)

	keys := make([]string, 0)
	for k, _ := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := values[k]
		vs, va, ve := t.processValue(RValue(v))
		if ve != nil {
			return "", nil, ve
		}

		sqls = append(sqls, k+" = "+vs)
		args = append(args, va...)
	}

	return strings.Join(sqls, ", "), args, nil
}

func normalizeSql(sql string) string {
	re := regexp.MustCompile("^\\s+")
	sql = re.ReplaceAllString(sql, "")
	re = regexp.MustCompile("\\s+$")
	sql = re.ReplaceAllString(sql, "")
	re = regexp.MustCompile("\\s+")
	return re.ReplaceAllString(sql, " ")
}
