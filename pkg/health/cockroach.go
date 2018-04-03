package health

const (
	createHealthTblStmt = `CREATE TABLE health (
		component_name STRING,
		component_id STRING,
		name STRING,
		duration INTERVAL,
		status STRING,
		error STRING,
		last_updated TIMESTAMPTZ,
		PRIMARY KEY (component_name, component_id, name))`
	insertHealthStmt = `INSERT INTO health (
		component_name,
		component_id,
		name,
		duration,
		status,
		error,
		last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	updateHealthStmt = `UPDATE health SET (duration, status, error, last_updated) = ($1, $2, $3, $4) 
		WHERE (component_name = $5 AND component_id = $6 AND name = $7)`
	selectHealthStmt = `SELECT * FROM health WHERE (component_name = $1 AND component_id = $2)`
)

// CockroachModule is the module that save health checks results in Cockroach DB.
type CockroachModule struct {
}

// NewCockroachModule returns the jaeger health module.
func NewCockroachModule(conn SystemDConn, httpClient JaegerHTTPClient, collectorHealthCheckURL string, enabled bool) *CockroachModule {
	return &CockroachModule{}
}
