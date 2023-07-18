package graphsql

import (
	"database/sql"
	"fmt"

	"testing"
	"github.com/stretchr/testify/assert"

	"github.com/dominikbraun/graph"

	_ "github.com/mattn/go-sqlite3"
)

func createStore[K comparable, T any]() (*Store[K, T], error) {

	db, err := sql.Open("sqlite3", "file::memory:")
	if err != nil {
		panic(err)
	}

	store := New[K, T](db, DefaultConfig)
	if store == nil {
		return nil, fmt.Errorf("failed to create new store")
	}

	if err := store.SetupTables(); err != nil {
		return nil, err
	}

	return store, nil
}

func TestImplementsStoreInterface(t *testing.T) {

	var store = Store[int, int]{}

	// this will throw a compile error if graphsql.Store doesn't implement the graph.Store interface
	var _ graph.Store[int, int] = (*Store[int, int])(&store)
}

func TestEdgeCount(t *testing.T) {

	assert := assert.New(t)

	store, err := createStore[int, int]()
	assert.Nil(err)
	assert.NotNil(store)

	store.AddVertex(1, 1, graph.VertexProperties{})
	store.AddVertex(2, 2, graph.VertexProperties{})

	edgeCount, err := store.EdgeCount()
	assert.Nil(err)
	assert.Equal(0, edgeCount)

	store.AddEdge(1, 2, graph.Edge[int]{ 1, 2, graph.EdgeProperties{} })

	edgeCount, err = store.EdgeCount()
	assert.Nil(err)
	assert.Equal(1, edgeCount)

	store.AddEdge(2, 1, graph.Edge[int]{ 2, 1, graph.EdgeProperties{} })

	edgeCount, err = store.EdgeCount()
	assert.Nil(err)
	assert.Equal(2, edgeCount)

	store.AddEdge(1, 1, graph.Edge[int]{ 1, 1, graph.EdgeProperties{} })

	edgeCount, err = store.EdgeCount()
	assert.Nil(err)
	assert.Equal(3, edgeCount)

	store.AddEdge(2, 2, graph.Edge[int]{ 2, 2, graph.EdgeProperties{} })

	edgeCount, err = store.EdgeCount()
	assert.Nil(err)
	assert.Equal(4, edgeCount)

	store.RemoveEdge(2, 2)

	edgeCount, err = store.EdgeCount()
	assert.Nil(err)
	assert.Equal(3, edgeCount)
}

func TestRemoveVertex(t *testing.T) {

	assert := assert.New(t)

	store, err := createStore[int, int]()
	assert.Nil(err)
	assert.NotNil(store)

	store.AddVertex(1, 1, graph.VertexProperties{})

	vertexCount, err := store.VertexCount()
	assert.Nil(err)
	assert.Equal(1, vertexCount)

	err = store.RemoveVertex(1)
	assert.Nil(err)

	vertexCount, err = store.VertexCount()
	assert.Nil(err)
	assert.Equal(0, vertexCount)

	// larger graph
	store.AddVertex(1, 1, graph.VertexProperties{})
	store.AddVertex(2, 2, graph.VertexProperties{})
	store.AddVertex(3, 3, graph.VertexProperties{})
	store.AddVertex(4, 4, graph.VertexProperties{})

	vertexCount, err = store.VertexCount()
	assert.Nil(err)
	assert.Equal(4, vertexCount)

	err = store.RemoveVertex(3)
	assert.Nil(err)

	vertexCount, err = store.VertexCount()
	assert.Nil(err)
	assert.Equal(3, vertexCount)

	_, _, err = store.Vertex(3)
	assert.NotNil(err)
}

func TestUpdateEdge(t *testing.T) {

	assert := assert.New(t)

	store, err := createStore[int, int]()
	assert.Nil(err)
	assert.NotNil(store)

	store.AddVertex(1, 1, graph.VertexProperties{})
	store.AddVertex(2, 2, graph.VertexProperties{})

	store.AddEdge(1, 2, graph.Edge[int]{ 1, 2, graph.EdgeProperties{} })
	store.AddEdge(2, 1, graph.Edge[int]{ 2, 1, graph.EdgeProperties{} })
	store.AddEdge(1, 1, graph.Edge[int]{ 1, 1, graph.EdgeProperties{} })
	store.AddEdge(2, 2, graph.Edge[int]{ 2, 2, graph.EdgeProperties{} })

	err = store.UpdateEdge(1, 1, graph.Edge[int]{ 1, 1, graph.EdgeProperties{
		Attributes: map[string]string{ "abc": "xyz"},
		Weight: 5,
		Data: "happy",
	}})

	assert.Nil(err)

	edge, err := store.Edge(1, 1)
	assert.Nil(err)
	assert.Equal(5, edge.Properties.Weight)
	assert.Equal("xyz", edge.Properties.Attributes["abc"])
	assert.Equal("happy", edge.Properties.Data)
}
