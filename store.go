package graphsql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"

	sq "github.com/Masterminds/squirrel"
)

// Store is a graph.Store implementation that uses an SQL database to store and retrieve graphs.
type Store[K comparable, T any] struct {
	db     *sql.DB
	config Config
}

// New creates a new SQL store that can be passed to graph.NewWithStore. It expects a database
// connection directly to the actual database schema in the form of a sql.DB instance.
func New[K comparable, T any](db *sql.DB, config Config) *Store[K, T] {
	return &Store[K, T]{
		db:     db,
		config: config,
	}
}

// SetupTables creates all required tables inside the configured database. The schema is documented
// in this library's README file.
func (s *Store[K, T]) SetupTables() error {
	_, err := s.db.Exec(createVerticesTableSQL(s.config))
	if err != nil {
		return fmt.Errorf("failed to set up %s table: %w", s.config.VerticesTable, err)
	}

	_, err = s.db.Exec(createEdgesTableSQL(s.config))
	if err != nil {
		return fmt.Errorf("failed to set up %s table: %w", s.config.EdgesTable, err)
	}

	return nil
}

// DestroyTables drops all tables and thus removes all data from the database.
func (s *Store[K, T]) DestroyTables() error {
	_, err := s.db.Exec(dropEdgesTableSQL(s.config))
	if err != nil {
		return fmt.Errorf("failed to set up %s table: %w", s.config.EdgesTable, err)
	}

	_, err = s.db.Exec(dropVerticesTableSQL(s.config))
	if err != nil {
		return fmt.Errorf("failed to set up %s table: %w", s.config.VerticesTable, err)
	}

	return nil
}

// AddVertex implements graph.Store.AddVertex.
func (s *Store[K, T]) AddVertex(hash K, value T, properties graph.VertexProperties) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	attributeBytes, err := json.Marshal(properties.Attributes)
	if err != nil {
		return err
	}

	_, err = sq.
		Insert(s.config.VerticesTable).
		Columns("hash", "value", "weight", "attributes").
		Values(hash, valueBytes, properties.Weight, attributeBytes).
		RunWith(s.db).
		Exec()

	return err
}

// RemoveVertex implements graph.Store.RemoveVertex.
func (s *Store[K, T]) RemoveVertex(hash K) error {
	_, err := sq.
		Delete(s.config.VerticesTable).
		Where(sq.Eq{
			"hash": hash,
		}).
		RunWith(s.db).
		Exec()

	return err
}

// Vertex implements graph.Store.Vertex.
func (s *Store[K, T]) Vertex(hash K) (T, graph.VertexProperties, error) {
	var (
		valueBytes      []byte
		attributesBytes []byte
		value           T
		properties      graph.VertexProperties
	)

	err := sq.
		Select("value", "weight", "attributes").
		From(s.config.VerticesTable).
		Where(sq.Eq{"hash": hash}).
		RunWith(s.db).
		QueryRow().
		Scan(&valueBytes, &properties.Weight, &attributesBytes)

	if err != nil {
		return value, properties, fmt.Errorf("failed to query vertex: %w", err)
	}

	if err = json.Unmarshal(valueBytes, &value); err != nil {
		return value, properties, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	if err = json.Unmarshal(attributesBytes, &properties.Attributes); err != nil {
		return value, properties, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	return value, properties, nil
}

// ListVertices implements graph.Store.ListVertices.
func (s *Store[K, T]) ListVertices() ([]K, error) {
	rows, err := sq.
		Select("hash").
		From(s.config.VerticesTable).
		RunWith(s.db).
		Query()

	if err != nil {
		return nil, fmt.Errorf("failed to query vertices: %w", err)
	}

	var hashes []K

	for rows.Next() {
		var hash K
		if err := rows.Scan(&hash); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// VertexCount implements graph.Store.VertexCount.
func (s *Store[K, T]) VertexCount() (int, error) {
	var count int

	err := sq.
		Select("count(hash)").
		From(s.config.VerticesTable).
		RunWith(s.db).
		QueryRow().
		Scan(&count)

	return count, err
}

// AddEdge implements graph.Store.AddEdge.
func (s *Store[K, T]) AddEdge(sourceHash, targetHash K, edge graph.Edge[K]) error {
	attributesBytes, err := json.Marshal(edge.Properties.Attributes)
	if err != nil {
		return err
	}

	_, err = sq.
		Insert(s.config.EdgesTable).
		Columns(
			"source_hash",
			"target_hash",
			"weight",
			"attributes",
			"data",
		).
		Values(
			sourceHash,
			targetHash,
			edge.Properties.Weight,
			attributesBytes,
			edge.Properties.Data,
		).
		RunWith(s.db).
		Exec()

	return err
}

// RemoveEdge implements graph.Store.RemoveEdge.
func (s *Store[K, T]) RemoveEdge(sourceHash, targetHash K) error {
	_, err := sq.
		Delete(s.config.EdgesTable).
		Where(sq.Eq{
			"source_hash": sourceHash,
			"target_hash": targetHash,
		}).
		RunWith(s.db).
		Exec()

	return err
}

// Edge implements graph.Store.Edge.
func (s *Store[K, T]) Edge(sourceHash, targetHash K) (graph.Edge[K], error) {
	edge := graph.Edge[K]{
		Source: sourceHash,
		Target: targetHash,
	}

	var attributesBytes []byte

	err := sq.
		Select("weight", "attributes", "data").
		From(s.config.EdgesTable).
		Where(sq.Eq{
			"source_hash": sourceHash,
			"target_hash": targetHash,
		}).
		RunWith(s.db).
		QueryRow().
		Scan(&edge.Properties.Weight, &attributesBytes, &edge.Properties.Data)

	if errors.Is(err, sql.ErrNoRows) {
		return edge, graph.ErrEdgeNotFound
	}

	if err != nil {
		return edge, fmt.Errorf("failed to scan row: %w", err)
	}

	if err = json.Unmarshal(attributesBytes, &edge.Properties.Attributes); err != nil {
		return edge, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	return edge, err
}

// ListEdges implements graph.Store.ListEdges.
func (s *Store[K, T]) ListEdges() ([]graph.Edge[K], error) {
	rows, err := sq.
		Select(
			"source_hash",
			"target_hash",
			"weight",
			"attributes",
			"data",
		).
		From(s.config.EdgesTable).
		RunWith(s.db).
		Query()

	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}

	var edges []graph.Edge[K]

	for rows.Next() {
		var (
			edge            graph.Edge[K]
			attributesBytes []byte
		)

		if err := rows.Scan(
			&edge.Source,
			&edge.Target,
			&edge.Properties.Weight,
			&attributesBytes,
			&edge.Properties.Data,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal(attributesBytes, &edge.Properties.Attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
		}

		edges = append(edges, edge)
	}

	return edges, nil
}

// EdgeCount implements graph.Store.EdgeCount.
func (s *Store[K, T]) EdgeCount() (int, error) {
	var count int

	// Please note that for some reason count(id) does not return the correct results for sqlite.
	err := sq.
		Select("count(source_hash)").
		From(s.config.EdgesTable).
		RunWith(s.db).
		QueryRow().
		Scan(&count)

	return count, err
}

func (s *Store[K, T]) UpdateEdge(sourceHash, targetHash K, edge graph.Edge[K]) error {

	attributesBytes, err := json.Marshal(edge.Properties.Attributes)
	if err != nil {
		return err
	}

	_, err = sq.Update(s.config.EdgesTable).
		Set("weight", edge.Properties.Weight).
		Set("attributes", attributesBytes).
		Set("data", edge.Properties.Data).
		Where("source_hash = ?", sourceHash).
		Where("target_hash = ?", targetHash).
		RunWith(s.db).
		Exec()

	return err
}
