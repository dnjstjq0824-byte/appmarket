# Database contract

Runtime SQL is hand-written `pgx/v5` inside each module's `infrastructure`
package. There is intentionally no sqlc pipeline: the previous `sqlc.yaml`
generated nothing (`dbgen` never existed, `queries/` was empty) while
repositories shipped real SQL, so the config was dead weight that hid drift.

Drift protection without a live database:

- `migrations/0001_initial_schema.up.sql` is the single source of truth for
  tables and columns.
- `query_conformance_test.go` fails when repository SQL references a table or
  column the schema never created. CI runs it on every push.
- `schema_test.go` guards migration file integrity (tables, down migration,
  ledger-balance trigger, seeds).

When adding a query: write parameterized SQL in the owning module's
infrastructure repository, then extend `query_conformance_test.go` with the
tables and columns the new query depends on.
