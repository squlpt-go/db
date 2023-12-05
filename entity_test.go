package db

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEntityFromRow(t *testing.T) {
	rows, err := getParentRows()

	if err != nil {
		t.Fatal(err)
	}

	rows.Next()
	e, err := entityFromRows(rows)

	if err != nil {
		t.Fatal(err)
	}

	name, _ := e.fields["parent_name"]

	fmt.Println(string((*name.Value).([]byte)))
}

func TestStructFromRow(t *testing.T) {
	rows, err := getParentRows()

	if err != nil {
		t.Fatal(err)
	}

	for rows.Next() {
		e, err := FromRows[Parent](rows)

		if err != nil {
			t.Fatal(err)
		}

		if e.ID == 0 {
			t.Errorf("Invalid struct from row: %#v", e)
		}

		if reflect.ValueOf(e.Timestamp).IsZero() {
			t.Errorf("Invalid timestamp from row: %#v", e.Timestamp)
		}

		t.Logf("%+v\n", e)
	}
}

func TestStructFromRowWithJoin(t *testing.T) {
	rows, err := getChildrenRows()

	if err != nil {
		t.Fatal(err)
	}

	rows.Next()
	e, err := FromRows[Child](rows)

	if err != nil {
		t.Fatal(err)
	}

	if e.Name == "" {
		t.Fatalf("Name didn't hydrate: %+v", e)
	}

	if e.Parent.Status != "active" {
		t.Fatal("Invalid Parent.Status:", e.Parent.Status)
	}

	t.Logf("%+v\n", e)
}

func TestStructFromMap(t *testing.T) {
	s, err := FromMap[Parent](map[string]any{
		"parent_id":   int64(123),
		"parent_name": nil,
		"parent_data": map[string]any{"key": "value"},
	})

	if err != nil {
		t.Fatal(s)
	}

	if s.ID != int64(123) {
		t.Fatal("Invalid Parent.ID")
	}

	if s.Name.Valid {
		t.Fatal("Parent.Name should be invalid")
	}

	if !reflect.DeepEqual(s.Data.Encoded, map[string]any{"key": "value"}) {
		t.Fatal("Parent.Data don't match")
	}
}

func TestToMap(t *testing.T) {
	i := Parent{
		ID:   int64(123),
		Data: NewJson(map[string]any{"key": "value"}),
	}

	m := ToMap(i)
	e := map[string]any{
		"parent_id":   int64(123),
		"parent_data": map[string]any{"key": "value"},
	}

	if !reflect.DeepEqual(m, e) {
		t.Errorf("not equal %+v, %+v", m, e)
	}
}

func GetField[T any](fields map[string]any, key string) (T, bool) {
	if value, ok := fields[key]; ok {
		if v, ok := value.(T); ok {
			return v, true
		}
	}

	var t T
	return t, false
}

func TestHydrate(t *testing.T) {
	s, err := FromMap[Parent](map[string]any{
		"parent_id":    int64(123),
		"hydrate_name": "hydrated",
		"parent_data":  map[string]any{"key": "value"},
	})

	if err != nil {
		t.Fatal(s, err)
	}

	if s.ID != int64(123) {
		t.Fatal("Invalid Parent.ID")
	}

	if !s.Name.Valid || s.Name.Wrapped != "hydrated" {
		t.Fatal("Parent.Name should be 'hydrated'", s.Name.Wrapped)
	}

	if !reflect.DeepEqual(s.Data.Encoded, map[string]any{"key": "value"}) {
		t.Fatal("Parent.Data don't match")
	}
}

func TestFlatten(t *testing.T) {
	actual := Flatten([]Parent{
		{
			ID:   int64(123),
			Data: NewJson(map[string]any{"key": "value"}),
		},
		{
			ID:   int64(124),
			Data: NewJson(map[string]any{"key1": "value1"}),
		},
	})

	expected := []map[string]any{
		{
			"parent_id":   int64(123),
			"parent_data": map[string]any{"key": "value"},
		},
		{
			"parent_id":   int64(124),
			"parent_data": map[string]any{"key1": "value1"},
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Error("failed asserting slices are equal")
		t.Log(actual)
		t.Log(expected)
	}
}

func TestInflate(t *testing.T) {
	actual, err := Inflate[Parent]([]map[string]any{
		{
			"parent_id":   int64(123),
			"parent_data": map[string]any{"key": "value"},
		},
		{
			"parent_id":   int64(124),
			"parent_data": map[string]any{"key1": "value1"},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	expected := []Parent{
		{
			ID:   int64(123),
			Data: NewJson(map[string]any{"key": "value"}),
		},
		{
			ID:   int64(124),
			Data: NewJson(map[string]any{"key1": "value1"}),
		},
	}

	for i := 0; i < len(expected); i++ {
		a := actual[i]
		e := expected[i]

		if !reflect.DeepEqual(a.ID, e.ID) {
			t.Error("failed asserting ids are equal")
		}

		if !reflect.DeepEqual(a.Data, e.Data) {
			t.Error("failed asserting data is equal")
		}
	}
}
