package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	// this import  needs to be done otherwise the mysql driver don't work
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	_ "github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logf.Log.WithName("mysql-internal")

// Config is used to connect to a MariaDBCluster
type Config struct {
	User     string
	Password string
	Host     string
	Port     int32
}

// NewConfigFromClusterKey returns a new Config based on a MariaDBCluster key
func NewConfigFromClusterKey(ctx context.Context, c client.Client, clusterKey client.ObjectKey) (*Config, error) {
	cluster := &mariadbv1alpha1.MariaDBCluster{}
	if err := c.Get(ctx, clusterKey, cluster); err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: cluster.Spec.RootPassword.Name, Namespace: cluster.Namespace}

	if err := c.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	if _, ok := secret.Data[cluster.Spec.RootPassword.Key]; !ok {
		return nil, errors.New("missing key in password secret")
	}
	return &Config{
		User:     "root",
		Password: string(secret.Data[cluster.Spec.RootPassword.Key]),
		Host:     cluster.GetPrimaryHeadlessSvc(),
		Port:     3306,
	}, nil
}

// GetMysqlDSN returns a data source name
func (c *Config) GetMysqlDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&multiStatements=true&interpolateParams=true",
		c.User, c.Password, c.Host, c.Port,
	)
}

// Rows interface is a subset of mysql.Rows
type Rows interface {
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

// SQLRunner interface is a subset of mysql.DB
type SQLRunner interface {
	QueryExec(ctx context.Context, query Query) error
	QueryRow(ctx context.Context, query Query, dest ...interface{}) error
	QueryRows(ctx context.Context, query Query) (Rows, error)
}

type sqlRunner struct {
	db *sql.DB
}

// SQLRunnerFactory a function that generates a new SQLRunner
type SQLRunnerFactory func(cfg *Config, errs ...error) (SQLRunner, func(), error)

// NewSQLRunner opens a connections using the given DSN
func NewSQLRunner(cfg *Config, errs ...error) (SQLRunner, func(), error) {
	var db *sql.DB
	var closeFn func()

	// make this factory accept a functions that tries to generate a config
	if len(errs) > 0 && errs[0] != nil {
		return nil, closeFn, errs[0]
	}

	db, err := sql.Open("mysql", cfg.GetMysqlDSN())
	if err != nil {
		return nil, closeFn, err
	}

	// close connection function
	closeFn = func() {
		if cErr := db.Close(); cErr != nil {
			log.Error(cErr, "failed closing the database connection")
		}
	}

	return &sqlRunner{db: db}, closeFn, nil
}

func (sr sqlRunner) QueryExec(ctx context.Context, query Query) error {
	_, err := sr.db.ExecContext(ctx, query.escapedQuery, query.args...)
	return err
}
func (sr sqlRunner) QueryRow(ctx context.Context, query Query, dest ...interface{}) error {
	return sr.db.QueryRowContext(ctx, query.escapedQuery, query.args...).Scan(dest...)
}
func (sr sqlRunner) QueryRows(ctx context.Context, query Query) (Rows, error) {
	rows, err := sr.db.QueryContext(ctx, query.escapedQuery, query.args...)
	if err != nil {
		return nil, err
	}

	return rows, rows.Err()
}
