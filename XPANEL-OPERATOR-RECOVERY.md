# Local recovery when migration stops startup

This procedure applies to the SQLite development package only. PostgreSQL needs a database administrator and is not covered here. Do not bypass the startup check, remove data, reset the database or paste database contents into chat.

1. Stop every Xpanel process using this state directory. Do not edit a live database. Keep the existing package and its private state directory unchanged until a backup exists.
2. Read the local error. Migration settings errors and vision repair errors identify the inbound number without printing its private JSON. If the error names an unavailable table, disk failure or another SQL error rather than malformed JSON, do not use the JSON repair below. Fix the storage/schema problem or restore a known-good backup instead.
3. Use Python 3 locally to make a consistent private backup before editing. Replace the path with the actual package's state/x-ui.db path:

    umask 077
    python3 - /absolute/path/to/state/x-ui.db <<'PY'
    import sqlite3, sys, pathlib, time, os
    source=pathlib.Path(sys.argv[1]).resolve()
    backup=source.with_name(source.name+'.before-repair-'+str(time.time_ns()))
    source_db=sqlite3.connect(source.as_uri()+'?mode=ro',uri=True)
    dest=sqlite3.connect(backup)
    source_db.backup(dest)
    dest.close(); source_db.close(); os.chmod(backup,0o600)
    print('Private backup:',backup)
    PY

4. Inspect only the named inbound locally using a trusted SQLite editor. The table is inbounds; the key is id; the relevant JSON columns are settings and stream_settings. These can contain credentials. Do not print or share them in a support report. The inbound number is not a client ID.
5. Replace the malformed value only with the exact valid JSON recovered from your own known-good backup or saved configuration for that inbound. Do not replace it with an empty object or delete the inbound to make startup pass, that can erase client configuration or change traffic behavior. If no trustworthy value is available, restore the full known-good backup instead of inventing a repair.
6. Validate the replacement locally with Python's json.loads or the editor's JSON validation, save the one identified column in a transaction, and keep the database file mode0600. This validates JSON syntax, not Xray protocol semantics.
7. Run the package binary's migrate command with XUI_DB_FOLDER set to this package's state directory. It must exit0 and report completion. If it fails, stop and inspect the new error rather than repeating edits. Then start the local package normally and verify login, inbound settings and intended traffic before reconnecting clients. A successful migration is not a full data-plane test.

To restore the full known-good SQLite backup: stop Xpanel, preserve the failed database privately under a separate filename, copy the known-good backup to the exact active x-ui.db path with mode0600, and then run migrate before starting. Never overwrite the only backup. WAL/SHM sidecars can make manual file replacement unsafe if any process still has the database open; keep all processes stopped and use a SQLite-aware local tool when unsure.

Rejected restore uploads use a separate x-ui.db.rejected-* filename named in the failure response. The previous database is restored to x-ui.db on the tested migration-error rollback path. Rejected files contain secrets and are not automatically expired. Once the previous installation is verified and the rejected upload is no longer needed, remove only that exact rejected file locally. Do not include it in support archives.
