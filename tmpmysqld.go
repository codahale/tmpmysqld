// Package tmpmysql provides the ability to spin up temporary mysqld instances
// for testing purposes.
package tmpmysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"io/ioutil"
	"os"
	"syscall"
	_ "github.com/go-sql-driver/mysql" // load MySQL driver
)

// A MySQLServer is a temporary instance of mysqld.
type MySQLServer struct {
	database string
	port     int
	mysqld   *exec.Cmd
	dataDir  string

	DB *sql.DB
}

// NewMySQLServer returns a new mysqld instance running on the given port.
func NewMySQLServer(port int) (*MySQLServer, error) {
	baseDir, err := getBaseDir()
	if err != nil {
		return nil, err
	}

	dataDir, err := ioutil.TempDir(os.TempDir(), "tmpmysql")
	if err != nil {
		return nil, err
	}

	if err := installDB(baseDir, dataDir); err != nil {
		return nil, err
	}

	mysqld := exec.Command(
		"mysqld",
		"--no-defaults",
		"--datadir="+dataDir,
		"--bind-address=127.0.0.1",
		"--port="+strconv.Itoa(port),
	)

	if err := mysqld.Start(); err != nil {
		return nil, err
	}

	return &MySQLServer{
		mysqld:  mysqld,
		dataDir: dataDir,
		port:    port,
	}, nil
}

// Initialize waits for the mysqld instance to become available, then creates
// the given database and sets it as the current database.
func (s *MySQLServer) Initialize(database string) error {
	if s.DB != nil {
		panic("already initialized")
	}

	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/", s.port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	s.DB = db

	// wait until the DB is available
	for db.Ping() != nil {
		time.Sleep(50 * time.Millisecond)
	}

	if _, err := db.Exec("CREATE DATABASE " + database); err != nil {
		return err
	}

	if _, err := db.Exec("USE " + database); err != nil {
		return err
	}

	return nil
}

// Stop terminates the mysqld instance and deletes the temporary directory which
// contains the database files.
func (s *MySQLServer) Stop() error {
	if s.mysqld == nil {
		panic("already stopped")
	}

	if err := s.DB.Close(); err != nil {
		return err
	}

	if err := s.mysqld.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	s.mysqld = nil

	if err := s.mysqld.Wait(); err != nil {
		return err
	}

	return os.RemoveAll(s.dataDir)
}

func installDB(baseDir, dataDir string) error {
	cmd := exec.Command(
		"mysql_install_db",
		"--no-defaults",
		"--basedir="+baseDir,
		"--datadir="+dataDir,
	)
	return cmd.Run()
}

func getBaseDir() (string, error) {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("mysql_config", "--variable=pkglibdir")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return filepath.Abs(buf.String() + "/../")
}
