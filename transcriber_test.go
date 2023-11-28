package db

import (
	"reflect"
	"testing"
)

func TestTranscribeSelect(t *testing.T) {
	transcriber := MySQLTranscriber{}

	q := NewQuery().
		Select(
			"name",
			NewQuery().
				Select("name").
				From("table2").
				As("alias"),
		).
		From("users").
		LeftJoinEq("roles", "users.role_id", "roles.role_id").
		WhereIn("category", []string{"A", "B", "C"}).
		WhereIn("category1", []string{}).
		Where(
			Or().
				Eq("age", 17).
				NotEq("age", 19),
		).
		GroupBy("field1").GroupBy("field2").
		HavingGt("field3", 1000.0).
		UnionAll(
			NewQuery().
				Select("*").
				From("table2").
				Limit(10, 0),
		)

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `
		SELECT name, (SELECT name FROM table2) AS alias
        FROM users
        LEFT JOIN roles ON users.role_id = roles.role_id
        WHERE category IN('A', 'B', 'C') AND FALSE AND (age = 17 OR age != 19)
        GROUP BY field1, field2
        HAVING field3 > 1000.000000
        UNION ALL
        SELECT *
        FROM table2
        LIMIT 10, 0`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Error("Failed asserting queries are the same")
		t.Log(normalizeSql(sql))
		t.Log(normalizeSql(expectedSql))
	}

	if !reflect.DeepEqual(args, []any{}) {
		t.Error("Failed asserting argument sets are the same")
	}
}

func TestTranscribeSelectWithArgs(t *testing.T) {
	transcriber := MySQLTranscriber{UsePlaceholders: true}

	q := NewQuery().
		Select(
			Raw("MAX(age)"),
			Ident("user.name"),
		).
		From(
			NewQuery().
				Select("name", "age").
				From("users").
				As("derived"),
		).
		RightJoin(
			NewQuery().
				Select("role_id").
				From("roles").
				As("roles"),
			Condition().Lt(3, Ident("users.role_id")),
		).
		InnerJoin("profiles", Condition().NotIn("profile_id", []any{12, 13})).
		WhereIsNotNull("category").
		WhereIsTrue("condition1").
		WhereIsNotFalse("condition2").
		Having(
			Or().
				LtEq("age", 17).
				GtEq("age", 19),
		).
		OrderBy("field1", Asc).OrderBy("field2", Desc)

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `SELECT MAX(age), user.name
        FROM (SELECT name, age FROM users) AS derived
        RIGHT JOIN (SELECT role_id FROM roles) AS roles ON ? < users.role_id
        INNER JOIN profiles ON profile_id NOT IN(?, ?)
        WHERE category IS NOT NULL AND condition1 IS TRUE AND condition2 IS NOT FALSE
        HAVING (age <= ? OR age >= ?)
        ORDER BY field1 ASC, field2 DESC`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Error("Failed asserting queries are the same")
	}

	if !reflect.DeepEqual(args, []any{3, 12, 13, 17, 19}) {
		t.Error("Failed asserting argument sets are the same")
	}
}

func TestTranscribeUpdateQuery(t *testing.T) {
	transcriber := MySQLTranscriber{UsePlaceholders: true}

	q := NewQuery().
		Update("users").
		Set(map[string]any{
			"field1": "value1",
			"field2": 2,
		}).
		InnerJoinEq("profiles", "profiles.profile_id", "users.profile_id").
		WhereIsNotNull("category").
		HavingIsNull("field3").
		Limit(1, 1)

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `UPDATE users
        INNER JOIN profiles ON profiles.profile_id = users.profile_id
        SET field1 = ?, field2 = ?
        WHERE category IS NOT NULL
        HAVING field3 IS NULL
        LIMIT 1, 1`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Error("Failed asserting queries are the same")
	}

	if !reflect.DeepEqual(args, []any{"value1", 2}) {
		t.Error("Failed asserting argument sets are the same")
	}
}

func TestTranscribeInsertQuery(t *testing.T) {
	transcriber := MySQLTranscriber{}

	q := NewQuery().
		InsertInto("users").
		Set(map[string]any{
			"field1": "value1",
			"field2": 2,
		})

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `INSERT INTO users SET field1 = 'value1', field2 = 2`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Errorf("Failed asserting queries are the same \n%s (actual) VS:\n%s", normalizeSql(sql), normalizeSql(expectedSql))
	}

	if !reflect.DeepEqual(args, []any{}) {
		t.Error("Failed asserting argument sets are the same")
	}
}

func TestTranscribeInsertIgnoreQuery(t *testing.T) {
	transcriber := MySQLTranscriber{UsePlaceholders: true}

	q := NewQuery().
		InsertIgnoreInto("users").
		Set(map[string]any{
			"field1": "value1",
			"field2": 2,
		})

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `INSERT IGNORE INTO users SET field1 = ?, field2 = ?`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Errorf("Failed asserting queries are the same \n%s VS:\n%s", normalizeSql(sql), normalizeSql(expectedSql))
	}

	if !reflect.DeepEqual(args, []any{"value1", 2}) {
		t.Error("Failed asserting argument sets are the same")
	}
}

func TestTranscribeInsertUpdateQuery(t *testing.T) {
	transcriber := MySQLTranscriber{UsePlaceholders: true}

	q := NewQuery().
		InsertUpdateInto("users").
		Set(map[string]any{
			"field1": "value1",
			"field2": 2,
		})

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `INSERT INTO users SET field1 = ?, field2 = ? ON DUPLICATE KEY UPDATE field1 = ?, field2 = ?`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Errorf("Failed asserting queries are the same \n%s VS:\n%s", normalizeSql(sql), normalizeSql(expectedSql))
	}

	if !reflect.DeepEqual(args, []any{"value1", 2, "value1", 2}) {
		t.Error("Failed asserting argument sets are the same")
	}
}

func TestTranscribeDeleteQuery(t *testing.T) {
	transcriber := MySQLTranscriber{UsePlaceholders: true}

	q := NewQuery().
		Delete("users.*").
		From("users").
		InnerJoinEq("profiles", "profiles.profile_id", "users.profile_id").
		WhereEq("category", 5).
		HavingIsNull("field3").
		Limit(1, 1)

	sql, args, err := transcriber.Transcribe(q)
	if err != nil {
		t.Error(err)
	}

	expectedSql := `DELETE users.*
		FROM users
        INNER JOIN profiles ON profiles.profile_id = users.profile_id
        WHERE category = ?
        HAVING field3 IS NULL
        LIMIT 1, 1`

	if normalizeSql(sql) != normalizeSql(expectedSql) {
		t.Error("Failed asserting queries are the same")
	}

	if !reflect.DeepEqual(args, []any{5}) {
		t.Error("Failed asserting argument sets are the same")
	}
}
