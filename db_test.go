package db

import (
	"database/sql"
	"fmt"
	"testing"
)

func getParentRows() (*sql.Rows, error) {
	db := DB()

	rows, err := db.Query("SELECT * FROM parents")

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func getChildrenRows() (*sql.Rows, error) {
	db := DB()

	rows, err := db.Query(
		"SELECT * " +
			"FROM children " +
			"INNER JOIN parents ON children.parent_id = parents.parent_id")

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func TestResult(t *testing.T) {
	rows, err := getChildrenRows()

	if err != nil {
		t.Fatal(err)
	}

	result := As[Child](rows)

	for result.Next() {
		s := result.Current()
		fmt.Println(s)
	}
}

func TestResultRow(t *testing.T) {
	rows, err := getChildrenRows()

	if err != nil {
		t.Fatal(err)
	}

	result := As[Child](rows)
	row, has := result.Row()
	if !has {
		t.Fatal("row not found")
	}

	if row.ID == 0 {
		t.Errorf("Invalid row returned from query: %#v", row)
	}
}

type aggregate struct {
	*Entity
	Count int64 `field:"count"`
}

func TestAggregate(t *testing.T) {
	db := DB()
	a, _ := Query[aggregate](db, Raw("SELECT COUNT(*) AS count FROM parents")).Row()

	if a.Count == 0 {
		t.Error("Did not populate aggregate successfully")
	}
}

func TestAutoClose(t *testing.T) {
	db := DB()

	for i := 0; i < 1000; i++ {
		r := Query[aggregate](db, Raw("SELECT COUNT(*) AS count FROM parents"))
		for r.Next() {
			// do nothing
		}
	}
}

func TestColumn(t *testing.T) {
	rows, err := getChildrenRows()
	if err != nil {
		t.Fatal(err)
	}

	result := As[Child](rows)

	t.Log(Column[string](result, "child_name"))
}
