# graph-sql

`graph-sql` is an SQL storage implementation for graph data structures created with
[_graph_](https://github.com/dominikbraun/graph).

## Usage

### 0. Get your database up and running

If you don't have a database running, you may spin up a MariaDB container with a `graph` database
to get started quickly.

```
docker run -e MARIADB_DATABASE=graph -e MARIADB_ROOT_PASSWORD=root -p 3306:3306 mariadb
```

### 1. Connect to your database

The first step is to establish a connection to your database server, more specifically to the actual
database schema.

For instance, a connection to the MariaDB container from the example above can be established as
follows:

```go
package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/graph")
	if err != nil {
		panic(err)
	}
}
```

### 2. Create a new store

Use the retrieved `sql.DB` instance to create a new store.

```go
store := graphsql.New[int, int](db, graphsql.DefaultConfig)
```

The `New` function has two type parameters. These are the types of the vertex hashes and vertex
values, and they have to be the same as for the graph itself. If you're not familiar with graph's
hashing system, check out the [concept of hashes](https://github.com/dominikbraun/graph#hashes).

This example uses a sane default configuration provided by this library, but you can configure it to
your needs (see [Configuration](#configuration)).

### 3. Set up the table schema

Create all required tables by calling `SetupTables`. This should only happen once.

```go
if err := store.SetupTables(); err != nil {
	log.Fatal(err)
}
```

### 4. Create a graph backed by the store

Finally, the store instance can be passed to `graph.NewWithStore`, which will create a graph backed
by this store.

```go 
g := graph.NewWithStore(graph.IntHash, store)
```

### 5. Full example

A complete program that utilizes a MariaDB database to store a graph of integers may look like as
follows:

```go
package main

import (
	"database/sql"
	"log"

	"github.com/dominikbraun/graph"
	graphsql "github.com/dominikbraun/graph-sql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/graph")
	if err != nil {
		panic(err)
	}

	store := graphsql.New[int, int](db, graphsql.DefaultConfig)

	if err = store.SetupTables(); err != nil {
		log.Fatal(err)
	}

	g := graph.NewWithStore(graph.IntHash, store)

	// This will persist two vertices and an edge in the database.
	_ = g.AddVertex(1)
	_ = g.AddVertex(2)
	_ = g.AddEdge(1, 2)
}

```

## Configuration

The table schema created by `graph-sql` can be configured using a `Config` passed to `New`. You can
either use the provided `DefaultConfig` or create your own one.

| Field             | Description                              | Default    |
|-------------------|------------------------------------------|------------|
| `VerticesTable`   | The name of the vertices table.          | `vertices` |
| `EdgesTable`      | The name of the edges table.             | `edges`    |
| `VertexHashType`  | The database type of the vertex hashes.  | `TEXT`     |
| `VertexValueType` | The database type of the vertex values.* | `JSON`     |

*Vertex values are stored as JSON by this library.

## Table schema

### `vertices`

| Column       | Type     | NULL | Key     | Extra            |
|--------------|----------|------|---------|------------------|
| `id`         | `BIGINT` | No   | Primary | `AUTO_INCREMENT` |
| `hash`       | `TEXT`   | Yes  |         |                  |
| `value`      | `JSON`   | Yes  |         |                  |
| `weight`     | `INT`    | Yes  |         |                  |
| `attributes` | `JSON`   | Yes  |         |                  |

### `edges`

| Column        | Type     | NULL | Key     | Extra            |
|---------------|----------|------|---------|------------------|
| `id`          | `BIGINT` | No   | Primary | `AUTO_INCREMENT` |
| `source_hash` | `TEXT`   | Yes  |         |                  |
| `target_hash` | `TEXT`   | Yes  |         |                  |
| `weight`      | `INT`    | Yes  |         |                  |
| `attributes`  | `JSON`   | Yes  |         |                  |
| `data`        | `BLOB`   | Yes  |         |                  |

## Graph operations

Check out the [graph](https://github.com/dominikbraun/graph) repository for an overview of built-in
graph algorithms and operations.
