package dao_utils

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type testDBConfig struct {
	User     string
	Password string
	Name     string
}

func TestOpenPostgresForTesting(t *testing.T) {
	dbConfig := &DataBaseConfig{
		User:     "postgres",
		Password: "admin",
		Name:     "postgres",
	}

	OpenPostgresForTesting(t, dbConfig)

	// Wait a bit for the DB to be ready to accept connections
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name)
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(1 * time.Minute)
	require.NoError(t, err, "could not connect to started postgres container")
	defer db.Close()

	// Simple query to check DB is up
	var one int
	err = db.QueryRow("SELECT 1").Scan(&one)
	require.NoError(t, err, "failed to query postgres container")
	require.Equal(t, 1, one, "unexpected result from postgres container")
}
