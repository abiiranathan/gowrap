package gowrap

import (
	"fmt"
	"io"
	"log"
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
