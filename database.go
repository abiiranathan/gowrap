package gowrap

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SqlLogLevel int

const (
	SILENT = iota
	WARN
	INFO
	ERROR
)

// Creates a new SQL logger for gorm and that will write to w
func NewLogger(logLevel SqlLogLevel, w io.Writer) logger.Interface {
	var level logger.LogLevel

	switch logLevel {
	case INFO:
		level = logger.Info
	case ERROR:
		level = logger.Error
	case WARN:
		level = logger.Warn
	default:
		level = logger.Silent
	}

	return logger.New(
		log.New(w, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		})
}

// Configuration struct for connecting to postgres database
type Config struct {
	// Connection data source name
	DSN string

	// gorm's sql logger.Interface. Create one with helper gowrap.NewLogger
	Logger logger.Interface

	// Use a connection pool.
	// SetMaxIdleConns(20), SetMaxOpenConns(200)
	UseConnPool bool
}

// Connect to the postgres database with the data source name.
func ConnectToPostgres(config Config) (*gorm.DB, error) {
	dsnParams := &DSNParamas{}
	ParseDSN(config.DSN, dsnParams)

	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{
		PrepareStmt:                              true,
		Logger:                                   config.Logger,
		DisableForeignKeyConstraintWhenMigrating: false,
		NowFunc: func() time.Time {
			if dsnParams.Timezone == "" {
				return time.Now()
			}

			utc, err := time.LoadLocation(dsnParams.Timezone)
			if err != nil {
				return time.Now()
			}

			return time.Now().In(utc)
		},
	})

	if err != nil {
		return nil, err
	}

	err = Ping(db)
	if err != nil {
		return nil, err
	}

	if config.UseConnPool {
		setConnPool(db)
	}

	return db, nil
}

const MemorySQLiteDB = "file::memory:"

// Connect to dbname. If dbname is nil, it connect to a memory sqlite database
// ForeignKey pragma is enabled by for all connections
func ConnectToSqlite3(dbname string, walMode bool) *gorm.DB {
	dsn := fmt.Sprintf("%s?cache=shared&_pragma=foreign_keys(1)", dbname)

	if walMode {
		dsn += "&_pragma=journal_mode(WAL)"
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to connect to sqlite database: %v", err)
	}

	return db
}

func Ping(db *gorm.DB) error {
	rawConn, _ := db.DB()
	return rawConn.Ping()
}

func setConnPool(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(200)
}

type DSNParamas struct {
	Database string // dbname
	User     string // user
	Password string // password, default ""
	Host     string // host, default: localhost
	Port     string // postgres port, default 5432
	SSLMode  string // ssl_mode, default=disabled
	Timezone string // Timezone
}

// parse postgres DSN into DSNParams struct
func ParseDSN(dsn string, params *DSNParamas) {
	if params == nil {
		return
	}

	paramsMap := map[string]string{}
	for _, s := range strings.Split(dsn, " ") {
		v := strings.Split(s, "=")

		if len(v) == 2 {
			paramsMap[v[0]] = v[1]
		}
	}

	params.Database = paramsMap["dbname"]
	params.User = paramsMap["user"]
	params.Host = paramsMap["host"]

	if paramsMap["port"] != "" {
		params.Port = paramsMap["port"]
	} else {
		params.Port = "5432"
	}

	params.Password = paramsMap["password"]
	params.SSLMode = paramsMap["sslmode"]
	params.Timezone = paramsMap["TimeZone"]
}

// writes sql statements to drop all postgres functions/procedures to w
func WriteDropFunctionsQueries(db *gorm.DB, w *bufio.Writer) error {
	sql := `SELECT 'DROP FUNCTION IF EXISTS ' || ns.nspname || '.' || proname 
	 || '(' || oidvectortypes(proargtypes) || ');' FROM pg_proc INNER JOIN pg_namespace ns
     ON (pg_proc.pronamespace=ns.oid) WHERE ns.nspname='public' order by proname;`

	return appendToSQL(db, sql, w)

}

// writes sql statements to drop all views to w
// Important for migrations
func WriteDropViewQueries(db *gorm.DB, w *bufio.Writer) error {
	sql := `SELECT 'DROP VIEW IF EXISTS ' || table_name || ' CASCADE;'
	FROM information_schema.views
	WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
	AND table_name !~ '^pg_';`

	return appendToSQL(db, sql, w)
}

// writes sql statements to drop all views to w
// Execute with psql since the postgres driver does not support
// multiple statements.
//
// this is important for migrations
func WriteDropTriggerQueries(db *gorm.DB, w *bufio.Writer) error {
	sql := `SELECT 'DROP TRIGGER IF EXISTS ' || trigger_name || ' ON ' ||
	event_object_table || ' CASCADE;' FROM information_schema.triggers
	WHERE trigger_schema NOT IN ('pg_catalog', 'information_schema')
	AND trigger_name !~ '^pg_';`
	return appendToSQL(db, sql, w)
}

// helper function that writes sql statements from executing sql to w
func appendToSQL(db *gorm.DB, sql string, w *bufio.Writer) error {
	rows, err := db.Raw(sql).Rows() // (*sql.Rows, error)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var statement string
		if err := rows.Scan(&statement); err != nil {
			return err
		}

		w.Write([]byte(statement))
		w.Write([]byte("\n"))

		if err = w.Flush(); err != nil {
			log.Fatalf("could not flush: %v\n", err)
		}

	}

	return nil
}

// Drops all views, functions, triggers
func MigrateViewsFunctionsAndTriggers(db *gorm.DB, database, user string) {
	tempPath := "/tmp/migrations.sql"
	tmpFile, err := os.Create(tempPath)

	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		os.Remove(tempPath)
	}()

	w := bufio.NewWriter(tmpFile)
	w.WriteString("SET client_min_messages TO ERROR;\n")
	if err = WriteDropViewQueries(db, w); err != nil {
		log.Fatalln(err)
	}

	if err = WriteDropTriggerQueries(db, w); err != nil {
		log.Fatalln(err)
	}

	if err = WriteDropFunctionsQueries(db, w); err != nil {
		log.Fatalln(err)
	}

	tmpFile.Close()

	// Execute this file in psql
	cmd := fmt.Sprintf("psql -U %s %s -f %s", user, database, tempPath)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	fmt.Println(string(out))

	if err != nil {
		log.Fatalln(err)
	}
}
