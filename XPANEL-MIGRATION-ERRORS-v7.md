# Migration error reporting and private rejected data, v7

This supersedes the v5/v6 notes where they say orphan/vision cleanup only logs errors. MigrateDB now propagates orphan SQL failure, vision load/parse/write/sync failure, and effective-flow lookup failure. The non-migration helper keeps its original best-effort API for existing callers; migration uses an error-returning checked helper. SQLite restore sends these errors into the same previous-file restoration path tested in v6. Startup and migrate CLI fail nonzero.

Actual orphan cleanup SQL failure injection proves API failure, old DB/user restored, old session and fresh login valid, no scheduled panel restart, process alive and core API listening. Requirement failure was separately injected in v6. Vision load, effective-flow query and write failures are regression-tested, not separately injected through a production HTTP upload. Full service package passes.

Uploaded database/dump writes use mode 0600 and explicit chmod, rebuilt SQLite files are chmod 0600, and rejected imports are secured before rename. Direct binary under umask000 was tested: rejected file mode is 0600. This protects file contents, not directory metadata or pre-existing administrator choices. Keep the state directory private.

Rejected databases contain credentials and other private configuration. They are deliberately retained with the exact unique path in the error so diagnosis does not destroy evidence. There is no background deletion or automatic expiry. After confirming the previous database is operating and the rejected import is no longer needed, the administrator should remove the named rejected file locally. Never send it in chat or include it in a support archive. Runtime distributions contain no such files.

Remaining limits: PostgreSQL is not live-tested and has no automatic rollback after imported data becomes active; filesystem failure branches are code-reviewed, not exhaustively injected; full multi-operation atomicity, concurrent request isolation and general production readiness are not established. A listening core API is not a payload proof for this failure path.
