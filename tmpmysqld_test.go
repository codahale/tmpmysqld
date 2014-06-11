package tmpmysql

import (
	"reflect"
	"testing"
)

func TestMySQLServer(t *testing.T) {
	server, err := NewMySQLServer(10000)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	if err := server.Initialize("test"); err != nil {
		t.Fatal(err)
	}

	if _, err := server.DB.Exec(`
CREATE TABLE things (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL
)
`); err != nil {
		t.Error(err)
	}

	if _, err := server.DB.Exec(`
INSERT INTO things (name) VALUES ("one"), ("two")
`); err != nil {
		t.Error(err)
	}

	rows, err := server.DB.Query("SELECT id, name FROM things")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	actual := make(map[int64]string)

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Error(err)
		}
		actual[id] = name
	}

	expected := map[int64]string{
		1: "one",
		2: "two",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Was %#v, but expected %#v", actual, expected)
	}
}
