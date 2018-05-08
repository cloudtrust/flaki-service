// +build integration

package health_test

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var (
	hostPort = flag.String("hostport", "127.0.0.1:26257", "cockroach host:port")
	user     = flag.String("user", "health", "user name")
	db       = flag.String("db", "health", "database name")
)

func TestIntNewCockroachModule(t *testing.T) {
	var db = setupCleanDB(t)
	rand.Seed(time.Now().UnixNano())

	var (
		componentName = "flaki-service"
		componentID   = strconv.FormatUint(rand.Uint64(), 10)
	)

	// The table health does not exists.
	_, err := db.Exec("SELECT * from health")
	assert.NotNil(t, err)

	var _ = NewCockroachModule(componentName, componentID, db)

	// NewCockroachModule create table health.
	_, err = db.Exec("SELECT * from health")
	assert.Nil(t, err)
}

func TestIntRead(t *testing.T) {
	var db = setupCleanDB(t)
	rand.Seed(time.Now().UnixNano())

	var (
		componentName = "flaki-service"
		componentID   = strconv.FormatUint(rand.Uint64(), 10)
		unit          = "influx"
		now           = time.Now().UTC().Round(time.Millisecond)
		reports       = []StoredReport{{Name: "ping", Duration: 1 * time.Second, Status: OK, Error: "", LastExecution: now, ValidUntil: now.Add(1 * time.Hour)}}
	)

	var m = NewCockroachModule(componentName, componentID, db)

	// Read health checks report for 'influx', it should be empty now.
	var r, err = m.Read(unit)
	assert.Nil(t, err)
	assert.Zero(t, len(r))

	// Save a health check report in DB.
	err = m.Update(unit, reports)
	assert.Nil(t, err)

	// Read health checks report for 'influx', now there is one result.
	r, err = m.Read(unit)
	assert.Nil(t, err)
	assert.Equal(t, reports, r)
}

func setupCleanDB(t *testing.T) *sql.DB {
	var db, err = sql.Open("postgres", fmt.Sprintf("postgresql://%s@%s/%s?sslmode=disable", *user, *hostPort, *db))
	assert.Nil(t, err)
	// Clean
	db.Exec("DROP table health")
	return db
}
