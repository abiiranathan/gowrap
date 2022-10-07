package gowrap_test

import (
	"os"
	"testing"

	"github.com/abiiranathan/gowrap"
)

func TestDatabaseConnection(t *testing.T) {
	t.Parallel()

	dbname := "testing.db"
	defer os.Remove(dbname)

	db := gowrap.ConnectToSqlite3(dbname, false)

	err := gowrap.Ping(db)
	if err != nil {
		t.Error(err)
	}
}

func TestParseDSN(t *testing.T) {
	t.Parallel()

	dsn := "dbname=dbname user=username password=password host=localhost sslmode=disable TimeZone=Africa/Kampala"
	params := &gowrap.DSNParamas{}
	gowrap.ParseDSN(dsn, params)

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
	gowrap.ParseDSN(dsn, params)
	if params.Port != "80000" {
		t.Errorf("port not parsed properly")
	}

	// Test when params is nil
	dsn = "database=db"
	params = &gowrap.DSNParamas{}
	gowrap.ParseDSN(dsn, nil)

	if params.Database != "" {
		t.Errorf("Expected database to be empty string, got %q", params.Database)
	}
}

func TestConnectToPostgres(t *testing.T) {
	t.Parallel()

	db, err := gowrap.ConnectToPostgres(gowrap.Config{
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
		_, err = gowrap.ConnectToPostgres(gowrap.Config{
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

	levels := []gowrap.SqlLogLevel{
		gowrap.INFO,
		gowrap.SILENT,
		gowrap.ERROR,
		gowrap.WARN,
	}

	for _, level := range levels {
		logger := gowrap.NewLogger(level, os.Stdout)
		if logger == nil {
			t.Error("logger is nil")
		}
	}
}
