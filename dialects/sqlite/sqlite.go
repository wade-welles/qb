package qb

import (
	"fmt"
	"strings"

	"github.com/slicebit/qb"
)

// SqliteDialect is a type of dialect that can be used with sqlite driver
type SqliteDialect struct {
	escaping bool
}

// NewSqliteDialect instanciate a SqliteDialect
func NewSqliteDialect() qb.Dialect {
	return &SqliteDialect{false}
}

func init() {
	qb.RegisterDialect("sqlite3", NewSqliteDialect)
	qb.RegisterDialect("sqlite", NewSqliteDialect)
}

// CompileType compiles a type into its DDL
func (d *SqliteDialect) CompileType(t qb.TypeElem) string {
	if t.Name == "UUID" {
		return "VARCHAR(36)"
	}
	return qb.DefaultCompileType(t, d.SupportsUnsigned())
}

// Escape wraps the string with escape characters of the dialect
func (d *SqliteDialect) Escape(str string) string {
	if d.escaping {
		return fmt.Sprintf(`"%s"`, str)
	}
	return str
}

// EscapeAll wraps all elements of string array
func (d *SqliteDialect) EscapeAll(strings []string) []string {
	return qb.EscapeAll(d, strings[0:])
}

// SetEscaping sets the escaping parameter of dialect
func (d *SqliteDialect) SetEscaping(escaping bool) {
	d.escaping = escaping
}

// Escaping gets the escaping parameter of dialect
func (d *SqliteDialect) Escaping() bool {
	return d.escaping
}

// AutoIncrement generates auto increment sql of current dialect
func (d *SqliteDialect) AutoIncrement(column *qb.ColumnElem) string {
	if !column.Options.InlinePrimaryKey {
		panic("Sqlite does not support non-primarykey autoincrement columns")
	}
	return "INTEGER PRIMARY KEY"
}

// SupportsUnsigned returns whether driver supports unsigned type mappings or not
func (d *SqliteDialect) SupportsUnsigned() bool { return false }

// Driver returns the current driver of dialect
func (d *SqliteDialect) Driver() string {
	return "sqlite3"
}

// GetCompiler returns a SqliteCompiler
func (d *SqliteDialect) GetCompiler() qb.Compiler {
	return SqliteCompiler{qb.NewSQLCompiler(d)}
}

// WrapError wraps a native error in a qb Error
func (d *SqliteDialect) WrapError(err error) qb.Error {
	return qb.Error{Orig: err}
}

// SqliteCompiler is a SQLCompiler specialised for Sqlite
type SqliteCompiler struct {
	qb.SQLCompiler
}

// VisitUpsert generates the following sql: REPLACE INTO ... VALUES ...
func (SqliteCompiler) VisitUpsert(context *qb.CompilerContext, upsert qb.UpsertStmt) string {
	var (
		colNames []string
		values   []string
	)
	for k, v := range upsert.ValuesMap {
		colNames = append(colNames, context.Compiler.VisitLabel(context, k))
		context.Binds = append(context.Binds, v)
		values = append(values, "?")
	}

	sql := fmt.Sprintf(
		"REPLACE INTO %s(%s)\nVALUES(%s)",
		context.Compiler.VisitLabel(context, upsert.Table.Name),
		strings.Join(colNames, ", "),
		strings.Join(values, ", "),
	)

	return sql
}
