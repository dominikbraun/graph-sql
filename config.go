package graphsql

import "fmt"

const (
	createVerticesTable = `
CREATE TABLE %s (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    hash %s,
    value %s,
	weight INT,
	attributes JSON
);`
	createEdgesTable = `
CREATE TABLE %s (
	id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
	source_hash %s,
	target_hash %s,
	weight INT,
	attributes JSON,
	data BLOB
);`
	dropVerticesTable = `DROP TABLE %s;`
	dropEdgesTable    = `DROP TABLE %s;`
)

// DefaultConfig is a sane default configuration of the table schema. Using DefaultConfig when
// creating a store using New makes sense for most users.
var DefaultConfig = Config{
	VerticesTable:   "vertices",
	EdgesTable:      "edges",
	VertexHashType:  "TEXT",
	VertexValueType: "JSON",
}

// Config configures the table schema, i.e. the table names and some data types of its columns.
type Config struct {
	VerticesTable   string
	EdgesTable      string
	VertexHashType  string
	VertexValueType string
}

func createVerticesTableSQL(c Config) string {
	return fmt.Sprintf(
		createVerticesTable,
		c.VerticesTable,
		c.VertexHashType,
		c.VertexValueType,
	)
}

func createEdgesTableSQL(c Config) string {
	return fmt.Sprintf(
		createEdgesTable,
		c.EdgesTable,
		c.VertexHashType,
		c.VertexHashType,
	)
}

func dropVerticesTableSQL(c Config) string {
	return fmt.Sprintf(
		dropVerticesTable,
		c.VerticesTable,
	)
}

func dropEdgesTableSQL(c Config) string {
	return fmt.Sprintf(
		dropEdgesTable,
		c.EdgesTable,
	)
}
