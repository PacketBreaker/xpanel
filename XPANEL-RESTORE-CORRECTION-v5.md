# Restore correction, v5

Based on pinned 3x-ui v3.11.0 commit c621ba9014a6aa21fd7c9ecf45655e44e24ebf13.

Empty selected inbound slices caused GORM to reject the bulk save and roll back migration work. Skip only the empty save, not later cleanup. A second early return when no external-proxy rows matched skipped legacy tag cleanup; now only actual query errors return early.

MigrateDB returns requirement errors to import callers rather than logging and reporting success. SQLite keeps its prior-file fallback on this failure. This does not establish atomic restore across all database operations or general propagation from separate orphan/vision repair functions, which retain upstream logging behavior. PostgreSQL restore is source-reviewed only, not exercised against a live PostgreSQL server.

Regression coverage includes empty and tunnel-only databases, later cleanup commits, nonempty VLESS/backfill cases, and injected write failure with rollback through MigrateDB. Actual own-synthetic SQLite production backup/import/restart/fresh login succeeded without migration errors. No user database was used.
