# Xpanel managed browser updates and host service, v16

This is an isolated development candidate, not a deployed service or a production security clearance. The v15 package is preserved. No production update endpoint, domain, signing account or private key is configured.

## Normal Linux host installation

Requirements: Linux amd64, compatible glibc (candidate requires symbols through GLIBC_2.34), Python 3, OpenSSL with Ed25519 support, `flock`, and systemd for the host-service route. Review the GPL/source notices and the package digest before installing.

1. Use `python3 xpanel-local.py install --root /opt/xpanel --archive <verified-package.zip> --sha256 <out-of-band-digest>` as a dedicated existing non-root account that will own the installation (for example, an operator-created xpanel account). No account is created automatically. A new owned directory is required; foreign directories are refused.
2. Run `/opt/xpanel/current/start-managed.sh` interactively, complete the local first-admin prompt, and verify login. Stop it before service setup. There are no shipped default credentials. Keep the default loopback listener until an operator deliberately configures access.
3. Prepare a unit without changing systemd: `python3 xpanel-service.py --root /opt/xpanel --unit-output /tmp/xpanel.service`. If the configured listener or base path differs, pass `--health-url` with its literal-loopback HTTP page. Pass `--service-user xpanel` for an explicitly selected existing install owner. Non-root preparation defaults to the current account. Root invocation requires a non-root service user. The generated system unit always contains explicit User/Group, empty capability sets, NoNewPrivileges and RestrictSUIDSGID. Root panel units are refused. A complete current managed release with matching file inventory/hashes and its runtime directory is required; half-installed roots are refused. Non-root --install-system refuses before any preparation/output mutation. Lock files use directory descriptors and no-follow opens; all user-tree operations on the root-install path run after dropping privileges, with no privileged lock creation/chown. The tree must already belong to that account, without group/world write access, unsafe symlinks or inaccessible ancestors. No automatic chown or foreign-directory adoption is done.
4. Inspect the generated unit and run `systemd-analyze verify /tmp/xpanel.service`. On the intended host, explicitly run the same service setup with a fresh output path, explicit `--service-user xpanel`, and `--install-system` as root. Root performs file preparation in a child after dropping to the service account and writes the fixed unit itself, never installing the user-writable preview as privileged configuration. It refuses an existing unit, then daemon-reloads, enables and starts `xpanel.service`. Failures are surfaced, not reported as success. If a later systemctl step fails, the unit or enable symlink may remain; inspect systemctl status before retrying. No hidden overwrite or automatic conversion of another panel is performed.
5. Check `systemctl status xpanel.service`, the local page, and authentication. Use `systemctl stop xpanel.service` before local installer/recovery operations.

The unprivileged unit is intended for ordinary userspace proxy listeners on unprivileged ports and user-space/no-kernel networking. It grants no CAP_NET_ADMIN, CAP_NET_RAW, CAP_NET_BIND_SERVICE or other capabilities. Low ports on hosts that require CAP_NET_BIND_SERVICE, kernel TUN/WireGuard routes/interfaces, transparent proxy routing/iptables/nftables, raw sockets and other privileged operations are not promised by this service profile. A host may allow some of these separately, but this installer does not change sysctls, device permissions, firewall rules or capabilities. A privileged deployment needs its own explicit operator design and review, not a fallback to root here.

Only unit generation/verification and the actual standalone supervisor lifecycle were tested here. Both system and user systemd managers report offline in this environment. Actual systemd enable/start/boot has not been exercised.

## Operator-owned trusted channel

The browser cannot choose a URL, trust key or initial sequence. The install owner creates `CHANNEL-CONFIG.json` in the managed root, mode 0600, with:

- `manifest_url`: final HTTPS manifest URL
- `signature_url`: final HTTPS detached Ed25519 signature URL
- `trusted_key`: absolute path to an independently delivered OpenSSL Ed25519 public key
- `installed_sequence`: independently confirmed installed release sequence, used only before the first channel receipt exists

Use the existing signed manifest schema in CHANNEL-UPDATE.md. Production URLs must use HTTPS. Literal-loopback HTTP is available only through an explicit `test_loopback_http: true` config for isolated fixtures. Never infer the initial sequence or trust key from the feed being checked.

The browser displays the verified release and sends its exact manifest SHA256 on confirmation. Missing, stale or changed hashes are refused. The supervisor checks the hash again after dequeuing, and the apply tool checks it again against the fetched manifest. Channel replay/rollback and signature, architecture, expiry, archive size/hash and inventory checks remain enforced. A current equal sequence can be previewed without offering another update; applying it is still refused.

The independent supervisor keeps the browser process out of its own replacement lifecycle. It stops the managed panel and waits for its runtime lock to be released, applies the existing owned installer with a SQLite snapshot, starts the selected release, and tests a configured loopback HTTP 200 response plus process liveness. This is a bounded startup check, not a proof of all proxy protocols, migrations or lasting health. If the new process fails this check, the prior release, database snapshot and channel receipt are restored. Failed jobs are reported as failed, even after successful recovery. An unavailable recovery receipt is a hard operator-recovery condition, not an excuse to declare success.

Job and result files persist in the owned root/state. A leftover active job is treated as interrupted and recovery is attempted before launching on the next supervisor start. This does not establish power-loss atomicity across filesystem, SQLite, channel receipt and systemd, nor recovery from every kill point. The standalone supervisor is not an OS containment boundary. Systemd stop controls its cgroup; an unmanaged SIGKILL can leave descendants needing operator inspection. Do not run another panel manually against the same state.

The service preparation command copies reviewed management scripts into `host-tools`. Those scripts are deliberately pinned outside the release being replaced. A panel archive does not silently replace the supervisor/service installer. Refreshing host-tools is an explicit stopped-service operator maintenance step: stop the service, back up host-tools/config, inspect verified replacement scripts, rerun the service preparation command as the install owner with a fresh unit-output path, inspect the generated unit, and start and verify the service. Active supervisor/runtime locks cause preparation to refuse. Replacing the existing system unit remains a separate reviewed operator step, not an automatic overwrite.

All files and processes use the same install-owner identity. This is not privilege isolation against an attacker who can already execute code as that identity. Keep the root private, protect key/config custody, and do not expose administration without a separate deployment/security review.

## Evidence

The v16 fixture exercises real production UI confirmation, signed package apply, release switch, restart, readiness/status, current-release preview, changed/missing hash refusals, bad archive refusal with old-panel login, and a signed inventory-valid release whose panel exits, causing database/release/channel rollback and successful old-panel login. Service generation and systemd static verification also pass. See accompanying evidence and artifact metadata for exact candidate binding.
