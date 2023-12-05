package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
)

var (
	ErrEntityNotFound = errors.New("entity not found")
)

func Query[T IEntity](db *sql.DB, query Transcribeable) *Rows[T] {
	q, args, err := query.Transcribe(db)
	if err != nil {
		panic(err)
	}

	rows := queryStd(db, query)
	r := As[T](rows)
	r.Query = q
	r.Args = args

	return r
}

func queryStd(db *sql.DB, query Transcribeable) *sql.Rows {
	q, args, err := query.Transcribe(db)
	if err != nil {
		panic(err)
	}

	writeLog(LogQueries, "QUERY: %s %+v", q, args)
	rows, err := db.Query(q, args...)
	if err != nil {
		m := fmt.Sprintf("FAILURE: %s in %s %+v", err, q, args)
		writeLog(LogFailures, m)
		panic(m)
	}

	return rows
}

func Exec(db *sql.DB, query Transcribeable) *Result {
	q, args, err := query.Transcribe(db)
	if err != nil {
		panic(err)
	}

	writeLog(LogQueries, "EXEC: %s %+v", q, args)
	result, err := db.Exec(q, args...)
	if err != nil {
		m := fmt.Sprintf("FAILURE: %s in %s %+v", err, q, args)
		writeLog(LogFailures, m)
		panic(m)
	}

	return &Result{result, q, args}
}

func As[T IEntity](rows *sql.Rows) *Rows[T] {
	return &Rows[T]{Rows: rows}
}

type Rows[T IEntity] struct {
	*sql.Rows
	Query string
	Args  []any
}

func (r *Rows[T]) Next() bool {
	hasNext := r.Rows.Next()

	if !hasNext {
		_ = r.Rows.Close()
	}

	return hasNext
}

func (r *Rows[T]) Current() T {
	s, err := FromRows[T](r.Rows)

	if err != nil {
		panic(err)
	}

	return s
}

func (r *Rows[T]) Row() (T, error) {
	hasNext := r.Rows.Next()
	defer r.Close()

	if !hasNext {
		var e T
		return e, ErrEntityNotFound
	}

	e := r.Current()
	return e, nil
}

func (r *Rows[T]) Slice() []T {
	defer r.Close()
	s := make([]T, 0)

	for r.Next() {
		s = append(s, r.Current())
	}

	return s
}

func (r *Rows[T]) Close() {
	_ = r.Rows.Close()
}

type Result struct {
	sql.Result
	Query string
	Args  []any
}

func Column[U any, T IEntity](r *Rows[T], name string) []U {
	col := make([]U, 0)
	defer r.Close()

	for r.Next() {
		c := r.Current()
		f := c.entityFields()

		if v, ok := f[name]; ok {
			var u U
			err := convertAssign(&u, *v.Value)
			if err != nil {
				panic("Column " + name + " of type " + reflect.ValueOf(v).Type().String() + " cannot be converted " + typeOf[U]().String())
			}
			col = append(col, u)
		} else {
			panic("Result of type " + typeOf[T]().String() + " does not have column " + name)
		}
	}

	return col
}

type LogLevel uint

const (
	LogFailures LogLevel = 1
	LogQueries  LogLevel = 2
	LogAll      LogLevel = LogFailures | LogQueries
)

var logger *log.Logger
var logLevel LogLevel

func SetLogger(l *log.Logger, level LogLevel) {
	logger = l
	logLevel = level
}

func writeLog(level LogLevel, message string, args ...any) {
	if logLevel&level > 0 {
		logger.Printf(message+"\n", args...)
	}
}
