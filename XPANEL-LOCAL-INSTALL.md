# Owned local install and offline update

`tools/xpanel-local.py` is a development-candidate local installer. It does not contact a server, install a service, change a firewall, request root, or start Xpanel. It does not replace a foreign 3x-ui installation. Python 3 and Linux `flock` are required. Use a new private directory whose parent already exists.

Install:

    python3 tools/xpanel-local.py install --root /absolute/private/xpanel --archive /absolute/path/runtime.zip --sha256 VERIFIED_ARCHIVE_SHA256

Use the exact SHA256 received with the intended Xpanel release through a trusted channel. Hashes detect changed bytes, they do not authenticate a malicious release or signer. There is no signed release channel or automatic network update in this candidate. Do not download an upstream 3x-ui binary and call it an Xpanel update.

The installer copies the supplied archive into a private temporary snapshot while hashing it, then extracts only that same snapshot. It verifies the archive hash, rejects unsafe paths/special files/private state, verifies every packaged file and the exact inventory against SHA256SUMS, and requires the Xpanel package's linux-amd64 provenance. Releases are private byte-verified directories under `releases`. The managed starter is a packaged hash-manifest file. Core configuration and geodata writes go to a separate per-release copy under `runtime-bin`, not the release tree. The runtime copy may change after launch, while release bytes remain verifiable. Local filesystem owners can still edit files, so this is not protection against a malicious local owner. It creates shared `state` and `logs` directories and switches the relative `current` link only after staging succeeds. A failed verification leaves the old current release selected. Earlier releases are retained.

Start from a local terminal:

    /absolute/private/xpanel/current/start-managed.sh

This uses one shared lock for the lifetime of the process, including its first-admin prompt. Always use this managed entry point. Directly running the binary or `start-local.sh` bypasses the installer lock. The tool cannot prove that an unmanaged process is stopped; you must stop every process using this state before updating or rolling back. Never mix managed and unmanaged starts or edit shared state while running.

Update while stopped:

    python3 tools/xpanel-local.py update --root /absolute/private/xpanel --archive /absolute/path/new-runtime.zip --sha256 VERIFIED_NEW_ARCHIVE_SHA256

The tool refuses while a managed process holds the lock. It creates a consistent SQLite backup and a private receipt under `history`, retains the old release, then selects the staged new release. It does not launch or health-check the new process automatically. Start it yourself and verify local login, settings and intended traffic before reconnecting clients. This initial version supports the owned SQLite portable workflow only, not PostgreSQL or system service installations.

If the new release does not work, stop it and use the exact receipt path printed by that update:

    python3 tools/xpanel-local.py rollback --root /absolute/private/xpanel --receipt /absolute/private/xpanel/history/UPDATE_RECEIPT.json

Rollback is explicit, not automatic. It requires the receipt to describe the currently selected release, backs up current SQLite state, restores the pre-update snapshot and selects the prior release. State written after the update will no longer be active, but a pre-rollback snapshot is retained. When there was no pre-update DB but a new one exists, rollback refuses rather than pairing old code with unknown new state. No releases or backup history are deleted automatically. Keep those files private because they contain credentials and configuration; remove old material only after deciding it is no longer needed.

Tested: full real archive install/update/SQLite snapshot rollback; foreign directory, hash mismatch, path traversal and held-lock refusal; private backup/receipt permissions; actual managed PTY first-admin setup and HTTP200, no password echo, actual running update refusal, clean stop. Abrupt power-loss recovery across multiple filesystem operations, malicious local writers, disk-full/rename fault matrix, signature verification, automatic health-check rollback and broader deployment platforms are not qualified. This is useful local installation and stopped-state recovery, not a production installer or a published update service.

Only v9-or-newer runtime archives with a packaged start-managed.sh are accepted by this version of the installer. Older generated starters are not reused. This is a new owned-directory format, not an in-place migration of older managed layouts.

The v10 installer records ownership kind xpanel-owned-local-install-v2. It refuses v1 roots from v8/v9 experiments before staging or database backup. Install into a new directory. Moving existing state requires a separately planned, stopped-state supported migration; this tool does not do it. Receipts and rollback apply only within the same v2 root and therefore cannot select a v1 generated-starter release.
