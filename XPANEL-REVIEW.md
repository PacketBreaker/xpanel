# Xpanel upstream-derived development checkpoint
24 November 2026

## Identity and scope
This candidate starts from the complete 3X-UI v3.11.0 source at commit c621ba9014a6aa21fd7c9ecf45655e44e24ebf13. It does not merge or replace the separate earlier Xpanel implementation. The full upstream backend, frontend routes, protocol implementations, database schema, integrations, tests and documents remain in the source. This is feature-source retention, not proof that every feature works in this environment.

Adaptation credit: Eng.Abdelrahman Rabie 2026. Upstream GPLv3 license and notices remain unchanged. Visible UI attribution identifies 3X-UI and GPLv3. See XPANEL-NOTICE.md.

## Changes
Custom Xpanel identity and credit components, login and desktop/mobile sidebar branding, page titles, Arabic/English identity strings, actual RTL direction for Arabic/Persian and Ant Design, bidi-isolated attribution, and focused tests. Backend production source and go.mod/go.sum are unchanged. One compatible transitive lock update changes follow-redirects1.16.1 to1.16.3, removing its reported high advisory. See exact XPANEL-CHANGES.json.

## Verified
- Official Go1.27.3 and Node26.13.1/npm11.20.0 obtained with vendor SHA256 verification. GCC11.4 CGO available.
- Pristine upstream frontend and Go builds pass. Derived frontend build, TypeScript typecheck and final Go build pass.
-8focused brand/language/Ant Design tests pass, including Arabic locale and RTL document direction.
- Real authenticated production HTTP router, SQLite, session/CSRF, websocket and actual Xray26.10.21 process work in a bounded owned-loopback harness. Both EN desktop and AR mobile dashboard pixels inspected; no missing-config toast, browser page errors, external browser requests or horizontal overflow.
- Upstream release-pinned Xray asset SHA256:0eec3ed4163c273d8c8bf821f7d219e257921cf584aa3fb970715e458e631978. Source:https://github.com/XTLS/Xray-core/releases/tag/v26.10.21
- Test harness has no production scheduler/subscription/bot startup, uses temporary database, reserved loopback API/metrics, no proxy inbounds, blackhole-only user egress and clean core Stop. It does not prove production startup or payload forwarding.

## Not yet verified or packaged
A release installer, service lifecycle, full clean-machine package and upgrade path are not done. The original self-update still targets upstream, not a branded update channel. All protocol payloads, quotas, backup/restore, remote nodes, PostgreSQL, MTProto sidecar, bots, provider integration, privileged tunnel/firewall and multiarchitecture deployments remain functionally unqualified for this derivative. They remain in source, not removed or represented as passed. The upstream README warns against production use. No production safety claim follows from this checkpoint.

The upstream release workflow uses mutable latest MTProto/geodata downloads. A reproducible final package must pin them and preserve their licenses/checksums, or clearly report unavailable components. No release should silently omit them.7moderate npm advisories remain,4in runtime dependency chains, pending separate review; no broad security clearance. Full backend/frontend regression suites and independent review remain outstanding.

## Rebuild
Use Go1.27.3 with CGO and Node26/npm11. In frontend run npm ci --ignore-scripts, npm run build and npm run typecheck. At source root run go build. Do not lower toolchain directives. The generated internal/web/dist is included for inspection but should be rebuilt from source for a release. No database, user credentials or runtime configuration is included in this source checkpoint.

## Sources
https://github.com/MHSanaei/3x-ui/tree/c621ba9014a6aa21fd7c9ecf45655e44e24ebf13
https://raw.githubusercontent.com/MHSanaei/3x-ui/v3.11.0/LICENSE
https://raw.githubusercontent.com/MHSanaei/3x-ui/v3.11.0/go.mod
https://raw.githubusercontent.com/MHSanaei/3x-ui/v3.11.0/frontend/package.json
https://raw.githubusercontent.com/MHSanaei/3x-ui/v3.11.0/.github/workflows/release.yml
https://github.com/advisories/GHSA-mrwp-f3q8-3g9p


## V2 packaging review update
V2 adds portable bootstrap/launcher, fail-closed upstream-overwrite guard, PWA name/custom icons and explicit bidi metric/provenance fixes. Bootstrap creates only a new private database, marks ownership, leaves owned repeats byte-identical and refuses an unmarked existing database. Ownership marker is a local operational guard, not authentication against a malicious local writer. No real database imported. Production launcher actually ran with HTTP200, three process-owned loopback-only listeners and clean exit0. This proof predates final ownership-marker addition; final bootstrap new/repeat/foreign-refusal proof passes separately. Final typecheck/eightlanguage tests/updaterguard/build and actual core+authenticated SOCKS payload+EN/AR backend pixels pass. Final pixels inspected numeric/unit/provenance correction and icon.20route navigation checks passed earlier inV2, before final bidi-only edits.

Automatic update now explicitly refuses before network/process effects because no branded release channel exists. This is a documented missing capability, not updater parity. Original updater code remains behind the guard. Portable staging includes pinned Xray26.10.21, MTG1.16.1 with licenses and Xray-archive geodata; additional regional geodata remain unpinned/unbundled. It is not a finished whole-feature package. Full test suites still timed out, no release/security acceptance. Credential-bearing private runtime logs/state are excluded from all deliverables.


## V3 lifecycle and locale correction
All14locale keysets now remain equal; new identity strings in12non-AR/EN locales explicitly use English fallback. Full frontend unit project passes87files1422tests across4shards. Component project1/16shard passes7files51tests; others remain incomplete/time-bounded. Service selection Client/Inbound/Traffic/Quota/Expiry/Backup passes883test/subtest events with17skips. TUIC/MTProto/AmneziaWG package tests pass. None establishes every external protocol/application.

First-admin startup no longer generates/logs a password. On an empty database it requires a controlling local terminal, reads a username and hidden12..72byte password plus confirmation, bcrypt hashes it and creates only if still empty inside a transaction. No terminal fails closed before listeners/admin creation; existing admin starts without prompt. Actual PTY and whole portable-launcher PTY proofs pass HTTP200 and cleanexit0, with password absent from capturedterminal. Initial test exposed prompt-before-echo-disable race, corrected by hiding input before prompt; final proof repeated after final transaction change. Startup now imports existing upstream indirect x/term0.47.0 directly; versions preserved. Top-level install/update/x-ui scripts now refuse wrongproduct operations. Exact originalscriptbytes preserved in upstream-scripts, clearly not Xpanel installers. Genuine automatic installer/updater still missing.

All4regionalgeodata assets acquired separately with exactreleasepins/vendorchecksums/license files for private staging. No runtime package is included in this source archive. Private staging test state must be removed before any eventual package archive. Final frontendbuild/typecheck/nativebuild and actual backend/core/payload/ENARpixels pass. Source checkpoint is for review, not finishedrelease or permission to publish/deploy. Existing user-owned data untouched.


## V4 recovery and restore qualification
V4 adds localTTY-only recover-admin with hidden confirmed password, single-admin refusal boundary, preserved ID and incremented login_epoch. Dedicated unit test and actualPTY recovery pass with no passwordecho, sameID and epoch+1. FirstsetupPTY/noninteractive refusal retested after codechange. Docs distinguish firstinteractive/laterheadless setup and command-line-secret hazard in original compatibility setting route. x/term directannotation corrected, versionunchanged. ActualHTTPbackup/restore marker roundtrip and restart-request capture pass in testharness; actualprocessrestart still not established by that proof. No realuserstate touched. Previousv3 privatepackage remains a separate immutable review artifact; it lacks newrecoverycommand and must not be relabeledv4.
