# Xpanel channel update preparation, v15

This is local preparation, not a published update service. No production endpoint, account, signing identity or user machine was created or changed.

## Available command

`python3 tools/xpanel-channel.py --manifest-url HTTPS_URL --signature-url HTTPS_URL --trusted-key PUBLIC_KEY_FILE --root OWNED_INSTALL`

First use requires `--initialize-floor INSTALLED_SEQUENCE`. The operator must obtain that sequence from the installed release publisher through an independent trusted channel, not copy it from the incoming feed. Without it first use refuses, including preview. Incoming releases must have a strictly higher sequence. Existing receipts forbid reinitialization. This is a deliberate operator initialization step, not automatic version detection. No production release sequence is assigned by this development package.

The default verifies a signed release description and previews its sequence, bytes and hash. Add `--apply` only for an explicit stopped-state update. It does not start the panel or a service. OpenSSL and Python3 are required. The trusted public key must be obtained independently of the feed. Signature verification uses OpenSSL pkeyutl with raw input; the acceptance fixture uses Ed25519. No signing private key belongs in a shipped package.

Manifest schema1 requires product Xpanel, architecture linux-amd64, positive integer sequence, future expires_unix, final archive_url, byte count and lowercase SHA256. Signature covers exact manifest bytes. Redirects, credential-bearing URLs and oversized files are refused. HTTP is refused except literal loopback with the explicit test-only flag. A signed archive still passes the existing complete file inventory, product and architecture checks before installation.

The existing local installer handles its lifetime lock, SQLite snapshot, release switch and receipt rollback. A channel lock serializes channel operations. The sequence receipt is saved only after install success. Explicit rollback remains an operator action through the local receipt command; replaying an old channel sequence is refused. The sequence receipt is separate from the release switch, not crash atomic: interruption between them can leave an applied release without a channel receipt and require reconciliation. File-system or malicious-local-owner attacks and unbounded remote package availability are not qualified.

## Local acceptance

A local HTTP fixture with explicit test mode and ephemeral Ed25519 signer verified preview, exact signature tamper refusal, wrong product/architecture refusal, expired feed refusal, wrong archive hash refusal, refusal while the managed runtime lock is held, actual update using a complete frozen v13 package, and replay refusal, signed non-object manifest refusal, missing initial-floor refusal and older/equal first-use sequence refusal. Archive hashing streams fixed-size chunks. Refusals leave current release selection unchanged. Updated file inventory hashes pass. This is not an HTTPS-hosted publication test, browser-button integration, unattended restart, rollback crash test or production signing ceremony.

## Decisions still owned by the user

Before publication: release account/domain and final endpoint, signer custody and independently delivered trust key, and approval to publish. The code needs none of those choices for local acceptance. Existing publication and server-action holds remain.

The original browser/network self-update remains guarded against replacing Xpanel with upstream. This CLI adapter does not silently redirect that UI button or claim browser-updater parity.

## External content and migration boundaries

The retained upstream sponsor feature fetches sponsor data through the panel service and may load sponsor images/links hosted outside the panel. Those third-party requests are inherited behavior, not an endorsement, a new permission grant or trusted instructions. UI acceptance tests blocked external requests; they do not prove third-party content availability. No sponsor feature was removed.

Managed installation uses a new owned directory and refuses an existing upstream root. Native panel DB backup/import and synthetic migration/failure recovery have been tested separately. This does not establish arbitrary historic-schema conversion or authorize using a real user database. In-place deployment conversion was not established as a requirement; this limitation must not be invented into one.
