package health_test

//go:generate mockgen -destination=./mock/cockroach.go -package=mock -mock_names=Cockroach=Cockroach  github.com/cloudtrust/flaki-service/pkg/health Cockroach


import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/health/mock"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

const (
	createHealthTblStmt = `CREATE TABLE IF NOT EXISTS health (
		component_name STRING,
		component_id STRING,
		unit STRING,
		name STRING,
		duration INTERVAL,
		status STRING,
		error STRING,
		last_updated TIMESTAMPTZ,
		valid_until TIMESTAMPTZ,
		PRIMARY KEY (component_name, component_id, unit, name))`
	upsertHealthStmt = `UPSERT INTO health (
		component_name,
		component_id,
		unit,
		name,
		duration,
		status,
		error,
		last_updated,
		valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	selectHealthStmt = `SELECT * FROM health WHERE (component_name = $1 AND component_id = $2 AND unit = $3)`
	cleanHealthStmt  = `DELETE from health WHERE (component_name = $1 AND valid_until < $2)`
)

func TestNewCockroachModule(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroach(mockCtrl)
	rand.Seed(time.Now().UnixNano())

	var (
		componentName = "flaki-service"
		componentID   = strconv.FormatUint(rand.Uint64(), 10)
	)

	mockCockroach.EXPECT().Exec(createHealthTblStmt).Return(nil, nil).Times(1)
	_ = NewCockroachModule(componentName, componentID, mockCockroach)
}

func TestUpdate(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroach(mockCtrl)
	rand.Seed(time.Now().UnixNano())

	var (
		componentName = "flaki-service"
		componentID   = strconv.FormatUint(rand.Uint64(), 10)
		unit          = "influx"
		reports       = []StoredReport{
			{Name: "ping", Duration: 1 * time.Second, Status: OK, Error: "", LastExecution: time.Now(), ValidUntil: time.Now().Add(1 * time.Hour)},
			{Name: "pong", Duration: 2 * time.Second, Status: KO, Error: "fail", LastExecution: time.Now(), ValidUntil: time.Now().Add(1 * time.Hour)},
		}
	)

	mockCockroach.EXPECT().Exec(createHealthTblStmt).Return(nil, nil).Times(1)
	var m = NewCockroachModule(componentName, componentID, mockCockroach)

	mockCockroach.EXPECT().Exec(upsertHealthStmt, componentName, componentID, unit, "ping", (1*time.Second).String(), OK.String(), "", gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	mockCockroach.EXPECT().Exec(upsertHealthStmt, componentName, componentID, unit, "pong", (2*time.Second).String(), KO.String(), "fail", gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	var err = m.Update(unit, reports)
	assert.Nil(t, err)
}

func TestUpdateFail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroach(mockCtrl)
	rand.Seed(time.Now().UnixNano())

	var (
		componentName = "flaki-service"
		componentID   = strconv.FormatUint(rand.Uint64(), 10)
		unit          = "influx"
		reports       = []StoredReport{
			{Name: "ping", Duration: 1 * time.Second, Status: OK, Error: "", LastExecution: time.Now(), ValidUntil: time.Now().Add(1 * time.Hour)},
		}
	)

	mockCockroach.EXPECT().Exec(createHealthTblStmt).Return(nil, nil).Times(1)
	var m = NewCockroachModule(componentName, componentID, mockCockroach)

	mockCockroach.EXPECT().Exec(upsertHealthStmt, componentName, componentID, unit, "ping", (1*time.Second).String(), OK.String(), "", gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("fail")).Times(1)
	var err = m.Update(unit, reports)
	assert.NotNil(t, err)
}
