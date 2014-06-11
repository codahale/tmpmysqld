tmpmysqld
=========

`tmpmysqld` allows you to spin up temporary instances of `mysqld` for testing
purposes:

```go
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

    // use temporary mysqld instance
    ...
}

```

For documentation, check [godoc](http://godoc.org/github.com/codahale/tmpmysqld).
