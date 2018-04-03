package health

import (
	"database/sql"
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	hostPort    = flag.String("hostport", "172.19.0.2:26257", "cockroach host:port")
	user        = flag.String("user", "health", "user name")
	db          = flag.String("db", "health", "database name")
	integration = flag.Bool("integration", false, "run the integration tests")
)

func Test(t *testing.T) {

}

func setupCleanDB(t *testing.T) *sql.DB {
	var db, err = sql.Open("postgres", fmt.Sprintf("postgresql://%s@%s/%s?sslmode=disable", *user, *hostPort, *db))
	assert.Nil(t, err)
	// Clean
	db.Exec("DROP table health")
	return db
}
