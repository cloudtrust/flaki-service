package health_test

//go:generate mockgen -destination=./mock/cockroach.go -package=mock -mock_names=Cockroach=Cockroach  github.com/cloudtrust/flaki-service/pkg/health Cockroach

import (
	"encoding/json"
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
		json JSONB,
		last_updated TIMESTAMPTZ,
		valid_until TIMESTAMPTZ,
		PRIMARY KEY (component_name, component_id, unit))`
	upsertHealthStmt = `UPSERT INTO health (
		component_name,
		component_id,
		unit,
		json,
		last_updated,
		valid_until)
		VALUES ($1, $2, $3, $4, $5, $6)`
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
		reports       = json.RawMessage(`{}`)
	)

	mockCockroach.EXPECT().Exec(createHealthTblStmt).Return(nil, nil).Times(1)
	var m = NewCockroachModule(componentName, componentID, mockCockroach)

	mockCockroach.EXPECT().Exec(upsertHealthStmt, componentName, componentID, unit, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	var err = m.Update(unit, 0, reports)
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
		reports       = json.RawMessage(`{}`)
	)

	mockCockroach.EXPECT().Exec(createHealthTblStmt).Return(nil, nil).Times(1)
	var m = NewCockroachModule(componentName, componentID, mockCockroach)

	mockCockroach.EXPECT().Exec(upsertHealthStmt, componentName, componentID, unit, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("fail")).Times(1)
	var err = m.Update(unit, 0, reports)
	assert.NotNil(t, err)
}
