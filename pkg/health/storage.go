package health

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrInvalid  = errors.New("report not valid")
	ErrNotFound = errors.New("health check report not found")
)

type StoredReport struct {
	ComponentName string
	ComponentID   string
	Module        string
	HealthCheck   string
	Report        json.RawMessage
	LastUpdated   time.Time
	ValidUntil    time.Time
}

// StorageModule is the module that save health checks results in Storage DB.
type StorageModule struct {
	componentName string
	componentID   string
	s             Storage
}

type Storage interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

const createHealthTblStmt = `
CREATE TABLE IF NOT EXISTS health (
	component_name STRING,
	component_id STRING,
	module STRING,
	healthcheck STRING,
	json JSONB,
	last_updated TIMESTAMPTZ,
	valid_until TIMESTAMPTZ,
PRIMARY KEY (component_name, component_id, module, healthcheck)
)`

// NewStorageModule returns the storage module.
func NewStorageModule(componentName, componentID string, s Storage) *StorageModule {
	// Init DB: create health table.
	s.Exec(createHealthTblStmt)

	return &StorageModule{
		componentName: componentName,
		componentID:   componentID,
		s:             s,
	}
}

const upsertHealthStmt = `
UPSERT INTO health (
	component_name,
	component_id,
	module,
	healthcheck,
	json,
	last_updated,
	valid_until)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

// Update updates the health checks reports stored in DB with the values 'jsonReport'.
func (sm *StorageModule) Update(module, jsonReport json.RawMessage, validity time.Duration) error {
	var now = time.Now()
	var _, err = sm.s.Exec(upsertHealthStmt, sm.componentName, sm.componentID, module, healthcheck, string(jsonReport), now.UTC(), now.Add(validity).UTC())

	if err != nil {
		return errors.Wrapf(err, "component '%s' with id '%s' could not update health check '%s' for unit '%s'", sm.componentName, sm.componentID, healthcheck, module)
	}

	return nil
}

const selectHealthStmt = `
SELECT * FROM health 
WHERE (component_name = $1 AND component_id = $2 AND module = $3 AND healthcheck = $4)`

// Read reads the reports in DB.
func (sm *StorageModule) Read(module, healthcheck string) (StoredReport, error) {
	var row = sm.s.QueryRow(selectHealthStmt, sm.componentName, sm.componentID, module, healthcheck)
	var (
		cName, cID, m, hc       string
		report                  json.RawMessage
		lastUpdated, validUntil time.Time
	)

	var err = row.Scan(&cName, &cID, &m, &hc, &report, &lastUpdated, &validUntil)
	if err != nil {
		return StoredReport{}, errors.Wrapf(err, "component '%s' with id '%s' could not read health check '%s' for module '%s': %s", sm.componentName, sm.componentID, healthcheck, module, err)
	}

	// If the health check was executed too long ago, the health check report
	// is considered not pertinant and an error is returned.
	if time.Now().After(validUntil) {
		return StoredReport{}, ErrInvalid
	}

	return StoredReport{
		ComponentName: cName,
		ComponentID:   cID,
		Module:        m,
		HealthCheck:   hc,
		Report:        report,
		LastUpdated:   lastUpdated.UTC(),
		ValidUntil:    validUntil.UTC(),
	}, nil
}

const cleanHealthStmt = `
DELETE from health 
WHERE (component_name = $1 AND valid_until < $2)`

// Clean deletes the old test reports that are no longer valid from the health DB table.
func (sm *StorageModule) Clean() error {
	var _, err = sm.s.Exec(cleanHealthStmt, sm.componentName, time.Now().UTC())

	if err != nil {
		return errors.Wrapf(err, "component '%s' with id '%s' could not clean health checks", sm.componentName, sm.componentID)
	}

	return nil
}
