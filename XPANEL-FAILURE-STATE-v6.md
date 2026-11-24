# Migration failure state, v6

Normal run checks inbound migration requirements before starting panel/subscription listeners. CLI migrate exits nonzero rather than printing completion after an error.

For SQLite restore requirement failure after database replacement, close the rejected import, move it to a unique x-ui.db.rejected-* path, restore the prior file to x-ui.db, and reopen it before the deferred core restart. The failure response names these outcomes and the rejected-file path. Do not delete a rejected file until it is no longer needed for diagnosis; it contains private database data. Keep the state directory private.

Actual production-router failure injection uses a synthetic SQLite trigger aborting migration orphan deletion. It verifies previous user/database active, injected trigger absent in active DB, old session works, fresh login works, no panel restart, live process and listening core API after restart. This does not prove a full core payload roundtrip during this failure test. Separate run/migrate processes on the rejected DB exit 1 before a panel listener starts.

PostgreSQL is untested live. Its error text makes clear that restored data may already be active and automatic rollback is unavailable. SQLite rollback here covers inbound requirement errors, not every possible restore or filesystem failure. Separate orphan/vision cleanup helpers retain upstream logging semantics. Tag cleanup now reaches databases previously skipped when no external-proxy rows existed, while existing collision avoidance remains tested.

v6 source changes relative v5: main.go, internal/web/service/server.go, this note and generated manifests. Earlier modified upstream inbound_migration_test.go remains listed in the cumulative inventory. All original candidates remain separate and unchanged.
