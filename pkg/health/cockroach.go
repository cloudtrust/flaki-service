package health


import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
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

// CockroachModule is the module that save health checks results in Cockroach DB.
type CockroachModule struct {
	componentName string
	componentID   string
	db            Cockroach
}

type Cockroach interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// StoredReport is the health report that is stored in DB.
type StoredReport struct {
	Name          string
	Duration      time.Duration
	Status        Status
	Error         string
	LastExecution time.Time
	ValidUntil    time.Time
}

// NewCockroachModule returns the cockroach storage module.
func NewCockroachModule(componentName, componentID string, db Cockroach) *CockroachModule {
	// Init DB: create health table.
	db.Exec(createHealthTblStmt)

	return &CockroachModule{
		componentName: componentName,
		componentID:   componentID,
		db:            db,
	}
}

// Update updates the health checks reports stored in DB with the values 'reports'.
func (c *CockroachModule) Update(unit string, reports []StoredReport) error {
	for _, r := range reports {
		var _, err = c.db.Exec(upsertHealthStmt, c.componentName, c.componentID, unit, r.Name, r.Duration.String(), r.Status.String(), r.Error, r.LastExecution.UTC(), r.ValidUntil.UTC())

		if err != nil {
			return errors.Wrapf(err, "component '%s' with id '%s' could not update health check '%s' for unit '%s'", c.componentName, c.componentID, r.Name, unit)
		}
	}
	return nil
}

// Read reads the reports in DB.
func (c *CockroachModule) Read(unit string) ([]StoredReport, error) {
	var rows, err = c.db.Query(selectHealthStmt, c.componentName, c.componentID, unit)
	if err != nil {
		return nil, errors.Wrapf(err, "component '%s' with id '%s' could not read health check '%s'", c.componentName, c.componentID, unit)
	}
	defer rows.Close()

	var reports = []StoredReport{}
	for rows.Next() {
		var (
			cName, cID, hcUnit, hcName, hcDuration, hcStatus, hcError string
			lastUpdated, validUntil                                   time.Time
		)

		var err = rows.Scan(&cName, &cID, &hcUnit, &hcName, &hcDuration, &hcStatus, &hcError, &lastUpdated, &validUntil)
		if err != nil {
			return nil, errors.Wrapf(err, "component '%s' with id '%s' could not read health check '%s'", c.componentName, c.componentID, unit)
		}

		var d time.Duration
		{
			var err error
			d, err = time.ParseDuration(hcDuration)
			if err != nil {
				return nil, errors.Wrapf(err, "component '%s' with id '%s' could not parse duration '%s'", c.componentName, c.componentID, hcDuration)
			}
		}

		reports = append(reports, StoredReport{
			Name:          hcName,
			Duration:      d,
			Status:        status(hcStatus),
			Error:         hcError,
			LastExecution: lastUpdated.UTC(),
			ValidUntil:    validUntil.UTC(),
		})
	}

	return reports, nil
}

// Clean deletes the old test reports that are no longer valid from the health DB table.
func (c *CockroachModule) Clean() error {
	var _, err = c.db.Exec(cleanHealthStmt, c.componentName, time.Now().UTC())

	if err != nil {
		return errors.Wrapf(err, "component '%s' with id '%s' could not clean health checks", c.componentName, c.componentID)
	}

	return nil
}
