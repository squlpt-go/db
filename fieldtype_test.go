package db

import (
	"database/sql"
	"database/sql/driver"
	"testing"
)

func TestNullable(t *testing.T) {
	var n = NewNullable(0)
	var e = &n
	var ea any = e
	var ep any = *e

	if !n.Valid {
		t.Error("Should be valid")
	}

	if _, ok := ea.(sql.Scanner); !ok {
		t.Error("Does  not implement Scanner interface")
	}

	if _, ok := ep.(driver.Valuer); !ok {
		t.Error("Does  not implement Valuer interface")
	}
}

func TestJsonEncoded(t *testing.T) {
	var j = NewJson(map[string]any{})
	var e = &j
	var ea any = e
	var ep any = *e

	if _, ok := ea.(sql.Scanner); !ok {
		t.Error("Does  not implement Scanner interface")
	}

	if _, ok := ep.(driver.Valuer); !ok {
		t.Error("Does  not implement Valuer interface")
	}
}
