package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	pgerrors "github.com/run-ai/runai/runai-common-packages/runaidb/internal/postgres/errors"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type RWSessionCreator struct {
	db orm.DB
}

func NewRWSessionCreator(db orm.DB) *RWSessionCreator {
	return &RWSessionCreator{
		db: db,
	}
}

func (s *RWSessionCreator) Create(user, password, schema string) error {
	sqlDB, err := s.db.Actual().DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reserve a session-creation connection: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.GetLogger().Errorf("failed to return session-creation connection to the pool: %v", err)
		}
	}()

	// CREATE USER, ALTER USER and the GRANTs all rewrite shared catalog rows
	// (pg_authid, pg_class), which Postgres does not serialize for us - running
	// them from several replicas at once fails with "tuple concurrently updated".
	lock := NewLocker(conn)
	acquired, err := lock.Acquire(schema)
	if err != nil {
		return fmt.Errorf("failed to acquire service user lock for schema %q: %w", schema, err)
	}
	if !acquired {
		return fmt.Errorf("service user lock for schema %q is already held", schema)
	}
	defer func() {
		released, err := lock.Release(schema)
		if err != nil {
			logger.GetLogger().Errorf("failed to release service user lock for schema %q: %v", schema, err)
			return
		}
		if !released {
			logger.GetLogger().Errorf("service user lock for schema %q was not held by the releasing session", schema)
		}
	}()

	return s.create(conn, user, password, schema)
}

func (s *RWSessionCreator) create(conn *sql.Conn, user, password, schema string) error {
	sanitizedUser := pgx.Identifier{user}.Sanitize()
	sanitizedSchema := pgx.Identifier{schema}.Sanitize()
	sanitizedPassword := escapeLiteral(password)

	err := s.exec(conn, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", sanitizedUser, sanitizedPassword))
	if err != nil {
		if !pgerrors.HasCode(err, pgerrors.ErrCodeDuplicateObject) {
			return err
		}
		// User already exists, so we update the password in case it changed since the user was first created
		logger.GetLogger().Infof("service user %q already exists, updating password", user)
		err = s.exec(conn, fmt.Sprintf("ALTER USER %s WITH LOGIN PASSWORD '%s';", sanitizedUser, sanitizedPassword))
		if err != nil {
			return err
		}
	} else {
		logger.GetLogger().Infof("created service user %q", user)
	}

	grants := []string{
		fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s;", sanitizedSchema, sanitizedUser),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO %s;", sanitizedSchema, sanitizedUser),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %s TO %s;", sanitizedSchema, sanitizedUser),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA %s TO %s;", sanitizedSchema, sanitizedUser),
	}
	for _, grant := range grants {
		if err := s.exec(conn, grant); err != nil {
			return err
		}
	}

	return nil
}

func (s *RWSessionCreator) exec(conn *sql.Conn, query string) error {
	_, err := conn.ExecContext(context.Background(), query)
	return err
}

func escapeLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
