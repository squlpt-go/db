
# Squlpt Database Library

Squlpt is a lightweight database library created to add functionality to the standard 
library's `database/sql` package.

It can be incrementally or only partially adopted based on your needs, and includes three 
basic components:

* A generic-based struct hydration layer
* A query builder 
* A simple schema-definition layer with utility functions to perform common operations based 
  on relations between tables

## Compatibility

It currently only operated with MySQL, but is built to be compatible with other database 
engines by implementing the `db.Transcriber` interface.

## Installation

```bash
go get github.com/squlpt-go/db
```
