package clickhouse

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kobsio/klogs/pkg/log"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// Row is the structure of a single row in ClickHouse.
type Row struct {
	Timestamp    time.Time
	Cluster      string
	Namespace    string
	App          string
	Pod          string
	Container    string
	Host         string
	FieldsString map[string]string
	FieldsNumber map[string]float64
	Log          string
}

// Client can be used to write data to a ClickHouse instance. The client can be created via the NewClient function.
type Client struct {
	client             *sql.DB
	database           string
	asyncInsert        bool
	waitForAsyncInsert bool
}

// Write writes a list of rows to the configured ClickHouse instance.
func (c *Client) Write(buffer []Row) error {
	var settings string

	if c.asyncInsert {
		if c.waitForAsyncInsert {
			settings = "SETTINGS async_insert = 1, wait_for_async_insert = 1"
		} else {
			settings = "SETTINGS async_insert = 1, wait_for_async_insert = 0"
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s.logs (timestamp, cluster, namespace, app, pod_name, container_name, host, fields_string, fields_number, log) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) %s", c.database, settings)

	tx, err := c.client.Begin()
	if err != nil {
		log.Error(nil, "Begin transaction failure", zap.Error(err))
		return err
	}

	smt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(nil, "Prepare statement failure", zap.Error(err))
		return err
	}

	for _, l := range buffer {
		_, err = smt.Exec(l.Timestamp, l.Cluster, l.Namespace, l.App, l.Pod, l.Container, l.Host, l.FieldsString, l.FieldsNumber, l.Log)

		if err != nil {
			log.Error(nil, "Statement exec failure", zap.Error(err))
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error(nil, "Commit failed failure", zap.Error(err))
		return err
	}

	return nil
}

// Close can be used to close the underlying sql client for ClickHouse.
func (c *Client) Close() error {
	return c.client.Close()
}

// NewClient returns a new client for ClickHouse. The client can then be used to write data to ClickHouse via the
// "Write" method.
func NewClient(address, username, password, database, dialTimeout, connMaxLifetime string, maxIdleConns, maxOpenConns int, asyncInsert, waitForAsyncInsert bool) (*Client, error) {
	parsedDialTimeout, err := time.ParseDuration(dialTimeout)
	if err != nil {
		return nil, err
	}

	parsedConnMaxLifetime, err := time.ParseDuration(connMaxLifetime)
	if err != nil {
		return nil, err
	}

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: strings.Split(address, ","),
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		DialTimeout: parsedDialTimeout,
	})
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(parsedConnMaxLifetime)

	if err := conn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Error(nil, fmt.Sprintf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace))
		} else {
			log.Error(nil, "could not ping database", zap.Error(err))
		}

		return nil, err
	}

	return &Client{
		client:             conn,
		database:           database,
		asyncInsert:        asyncInsert,
		waitForAsyncInsert: waitForAsyncInsert,
	}, nil
}
