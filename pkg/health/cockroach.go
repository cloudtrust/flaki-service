package health

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

const (
	createHealthTblStmt = `CREATE TABLE health (
		component_name STRING,
		component_id STRING,
		unit STRING,
		name STRING,
		duration INTERVAL,
		status STRING,
		error STRING,
		last_updated TIMESTAMPTZ,
		PRIMARY KEY (component_name, component_id, unit, name))`
	insertHealthStmt = `INSERT INTO health (
		component_name,
		component_id,
		unit,
		name,
		duration,
		status,
		error,
		last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	updateHealthStmt = `UPDATE health SET (duration, status, error, last_updated) = ($1, $2, $3, $4) 
		WHERE (component_name = $5 AND component_id = $6 AND unit = $7 AND name = $8)`
	selectHealthStmt = `SELECT * FROM health WHERE (component_name = $1 AND component_id = $2 AND unit = $3 AND name = $4)`
)

// CockroachModule is the module that save health checks results in Cockroach DB.
type CockroachModule struct {
	componentName string
	componentID   string
	db            cockroachDB
}

type cockroachDB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// NewCockroachModule returns the cockroach storage module.
func NewCockroachModule(componentName, componentID string, db cockroachDB) *CockroachModule {
	// Init DB: create health table.
	db.Exec(createHealthTblStmt)

	return &CockroachModule{
		componentName: componentName,
		componentID:   componentID,
		db:            db,
	}
}

func (c *CockroachModule) Update(reports []Report) error {

	return nil
}

func (c *CockroachModule) Read(name string) ([]Report, error) {
	var row = c.db.Query(selectHealthStmt, c.componentName, c.componentID, name)
	var (
		cName, cID, hcName, hcDuration, hcStatus, hcError string
		lastUpdated                                       time.Time
	)

	var err = row.Scan(&cName, &cID, &hcName, &hcDuration, &hcStatus, &hcError, &lastUpdated)
	if err != nil {
		return nil, errors.Wrapf(err, "component '%s' with id '%s' could not read health check '%s'", c.componentName, c.componentID, name)
	}

	var d time.Duration
	{
		var err error
		d, err = time.ParseDuration(hcDuration)
		if err != nil {
			return nil, errors.Wrapf(err, "component '%s' with id '%s' could not parse duration '%s'", c.componentName, c.componentID, hcDuration)
		}
	}

	return &table{
		componentName: cName,
		componentID:   cID,
		jobName:       jName,
		jobID:         jID,
		enabled:       enabled,
		status:        status,
		lockTime:      lockTime.UTC(),
	}, nil
}

func (c *CockroachModule) Register(unitName, checkName string) error {
	var _, err = c.db.Exec(insertHealthStmt, c.componentName, c.componentID, unitName, checkName, time.Duration(0).String(), "", "", time.Time{})

	if err != nil {
		return errors.Wrapf(err, "component '%s' with id '%s' could not register health check '%s'", c.componentName, c.componentID, unitName, checkName)
	}

	return nil
}
