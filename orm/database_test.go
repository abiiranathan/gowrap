package orm_test

import (
	"bufio"
	"os"
	"testing"

	"github.com/abiiranathan/gowrap/orm"
)

func TestDatabaseConnection(t *testing.T) {
	t.Parallel()

	dbname := "testing.db"
	defer os.Remove(dbname)

	db := orm.ConnectToSqlite3(dbname, false)

	err := orm.Ping(db)
	if err != nil {
		t.Error(err)
	}
}

func TestParseDSN(t *testing.T) {
	t.Parallel()

	dsn := "dbname=dbname user=username password=password host=localhost sslmode=disable TimeZone=Africa/Kampala"
	params := &orm.DSNParamas{}
	orm.ParseDSN(dsn, params)

	if params.Database != "dbname" {
		t.Errorf("expected database to be dbname, got %q", params.Database)
	}

	if params.Password != "password" {
		t.Errorf("expected password to be password, got %q", params.Password)
	}

	if params.User != "username" {
		t.Errorf("expected User to be username, got %q", params.User)
	}

	if params.Host != "localhost" {
		t.Errorf("expected host to be localhost, got %q", params.Host)
	}

	if params.SSLMode != "disable" {
		t.Errorf("expected sslmode to be disable, got %q", params.SSLMode)
	}

	if params.Timezone != "Africa/Kampala" {
		t.Errorf("expected timezone to be Africa/Kampala, got %q", params.Timezone)
	}

	if params.Port != "5432" {
		t.Errorf("expected timezone to be 5432, got %q", params.Port)
	}

	// Test dsn with port provided
	dsn = "port=80000"
	orm.ParseDSN(dsn, params)
	if params.Port != "80000" {
		t.Errorf("port not parsed properly")
	}

	// Test when params is nil
	dsn = "database=db"
	params = &orm.DSNParamas{}
	orm.ParseDSN(dsn, nil)

	if params.Database != "" {
		t.Errorf("Expected database to be empty string, got %q", params.Database)
	}
}

func TestConnectToPostgres(t *testing.T) {
	t.Parallel()

	db, err := orm.ConnectToPostgres(orm.Config{
		DSN:         "dbname=database host=128.0.01",
		UseConnPool: true,
	})

	if err == nil {
		t.Errorf("expected connection to fail with empty config, %v", err)
	}

	if db != nil {
		t.Errorf("expected db to be nil, got %v", db)
	}

	if os.Getenv("DSN") != "" {
		_, err = orm.ConnectToPostgres(orm.Config{
			DSN:         os.Getenv("DSN"),
			UseConnPool: true,
		})

		if err != nil {
			t.Errorf("connect to real database failed with error: %v", err)
		}

		if db != nil {
			db.NowFunc()
		}
	}

}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	levels := []orm.SqlLogLevel{
		orm.INFO,
		orm.SILENT,
		orm.ERROR,
		orm.WARN,
	}

	for _, level := range levels {
		logger := orm.NewLogger(level, os.Stdout)
		if logger == nil {
			t.Error("logger is nil")
		}
	}
}

type fakeWriter struct {
	written int
}

func (f *fakeWriter) Write(b []byte) (n int, err error) {
	f.written += len(b)
	return f.written, nil
}

func TestMigrationScrips(t *testing.T) {
	if os.Getenv("DSN") == "" {
		return
	}

	db, err := orm.ConnectToPostgres(orm.Config{
		DSN:         os.Getenv("DSN"),
		UseConnPool: true,
	})

	if err != nil {
		t.Fatal(err)
	}

	f := &fakeWriter{}
	w := bufio.NewWriter(f)
	err = orm.WriteDropFunctionsQueries(db, w)
	if err != nil {
		t.Errorf("WriteDropFunctionsQueries failed with err: %v", err)
	}

	err = orm.WriteDropTriggerQueries(db, w)
	if err != nil {
		t.Errorf("WriteDropTriggerQueries failed with err: %v", err)
	}

	err = orm.WriteDropViewQueries(db, w)
	if err != nil {
		t.Errorf("WriteDropViewQueries failed with err: %v", err)
	}
}
