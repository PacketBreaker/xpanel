# Xpanel

A web panel for Xray and related proxy services, adapted from **3x-ui v3.11.0**, pinned to upstream commit `c621ba9014a6aa21fd7c9ecf45655e44e24ebf13`.

**Eng.Abdelrahman Rabie 2026**

> Development candidate v20. Local panel, proxy-traffic, update and recovery tests have been run. This is not a production-security certification or a claim that every deployment mode works on every host. No server is installed or changed by opening this repository.

## What Xpanel includes

- Administration for the upstream panel's clients, inbounds, quotas, expiry, subscriptions, QR codes, statistics, routing and nodes. The interface ships in English (LTR) and Arabic (RTL).
- Local installation inside a directory owned by the service account, with interactive creation of the first administrator and no default credentials.
- Browser-based updates through an operator-configured signed Xpanel channel: the panel is stopped, updated and restarted, with a SQLite snapshot for rollback if the new version fails to start.
- systemd service setup with a non-root account, without granting special network capabilities automatically.

A feature being present in the source does not mean it has been tested in every deployment environment. The SSH support requested earlier is not part of this Xpanel scope.

## Install

The prebuilt package targets **Linux amd64** and depends on glibc symbols up to `GLIBC_2.34`. You need Python 3, `flock` and OpenSSL with Ed25519 support, plus systemd only if you want to install a system service.

1. Download the runtime package `xpanel-private-linux-amd64-development-2026-11-24-v20.zip` from Releases. The word "private" in the archive name describes the original local build's naming, not repository access.
2. Verify the SHA256 shown below, then extract the package into your own staging directory. Do not run the legacy 3x-ui install scripts present in the source as if they were the Xpanel installer.
3. Use an existing non-root service account. Create the parent directory with suitable permissions, but leave the new install directory itself to the installer. Run the install commands below as the service account.
4. Run `current/start-managed.sh` from an interactive terminal to create the first administrator. The panel starts on `127.0.0.1:2053` and subscriptions stay disabled until you configure them. For access from another machine, use a secure tunnel or a reverse proxy setup you have reviewed yourself. Do not expose the admin panel to the internet without protection.
5. Stop the interactive run before service setup. Generate the unit file, review it, then explicitly run the system install step as root with the same service account name.

```sh
# From the extracted runtime directory, as the non-root install owner:
python3 xpanel-local.py install \
  --root /opt/xpanel \
  --archive /absolute/path/xpanel-private-linux-amd64-development-2026-11-24-v20.zip \
  --sha256 7ceaa33e3093c25cdedc28dd5b7cc3e8fe5f245351361d5fba9dcec7a70ba0f3

/opt/xpanel/current/start-managed.sh

# Stop the interactive panel before preparing the service.
python3 xpanel-service.py --root /opt/xpanel \
  --service-user xpanel --unit-output /tmp/xpanel-preview.service
systemd-analyze verify /tmp/xpanel-preview.service

# Explicit system installation, run as root only after reviewing the unit:
python3 xpanel-service.py --root /opt/xpanel \
  --service-user xpanel --unit-output /root/xpanel-install.service --install-system
systemctl status xpanel.service
```

`xpanel` is an example name for an existing account, not an account the program creates. The account must own the install directory and its files and be able to traverse the parent directories. The installer refuses foreign directories or incomplete copies and does not import an old database automatically. If the port or panel path changes, set `--health-url` to a suitable literal local HTTP address, such as `http://127.0.0.1:2053/`.

## Usage

After logging in, configure the inbound, clients, quota and expiry date, then verify connectivity with a real client before relying on the service. Enable subscriptions and their addresses only after reviewing the access settings. Keep backups private, because they contain sensitive data.

## Updates and recovery

The browser update does not download stock 3x-ui in place of Xpanel. The operator must first configure a `CHANNEL-CONFIG.json`, an independently delivered trusted public key, HTTPS manifest/signature endpoints, and an independently grounded installed sequence. There is no production channel or embedded private keys. Browser confirmation binds the exact signed manifest hash. The supervisor stops the panel, applies the verified owned package, restarts it and checks panel HTTP 200 and process liveness. A failed new-panel startup restores the previous release, the SQLite snapshot and the channel receipt. This does not establish all-protocol readiness or power-loss atomicity.

In the complete source ZIP, see `XPANEL-MANAGED-UPDATES-v16.md`, `XPANEL-LOCAL-INSTALL.md`, `XPANEL-CHANNEL-v15.md`, and `XPANEL-OPERATOR-RECOVERY.md`. The older local-installer document describes the standalone stopped-state tool; the newer managed-updates document governs the supervised browser/service workflow. `MANAGED-UPDATES.md` inside the runtime package covers the commands, trust and rollback conditions, and host-tools maintenance.

## Verification and limits

Local tests exercised actual panel login, proxy payloads for multiple protocol paths, signed browser updates, archive/hash/replay refusals, a failing new release with rollback and relogin, and non-root service-layout checks. Independent review reproduced selected core and lifecycle cases. Test coverage is not a promise of full upstream parity in every environment.

- Actual systemd boot/start and the root privilege-demotion path remain untested in the available review environment. Unit generation, static verification and local non-root runs passed.
- The service does not grant CAP_NET_ADMIN, CAP_NET_RAW or CAP_NET_BIND_SERVICE. Kernel TUN/WireGuard modes, route or firewall changes and privileged low ports are not guaranteed under this setup.
- The startup check is a panel HTTP 200 response plus process liveness, not proof that every proxy protocol is ready.
- Quota enforcement is poll-based and may allow overshoot before enforcement; there is no guaranteed time or size bound.
- Live PostgreSQL, external provider accounts and every power-loss/kill interruption point are not fully qualified. The host-tools service manager installed by the package updates manually after the service is stopped and does not replace itself automatically.

GitHub Actions is disabled on this repository so preserved upstream workflows do not publish or deploy automatically.

## Source access

The complete source is browsable in this repository. The exact frozen source review ZIP and full Linux amd64 runtime are attached to [the v20 development release](https://github.com/PacketBreaker/xpanel/releases/tag/v20-candidate). The source ZIP retains all 2,458 reviewed files and notices. The repository README is an additional installation/operation guide; the frozen source archive is unchanged. The release tag identifies the frozen source import commit, while main may contain newer documentation.

## Frozen artifact checksums

| Artifact | SHA256 |
| --- | --- |
| Linux amd64 runtime v20 | `7ceaa33e3093c25cdedc28dd5b7cc3e8fe5f245351361d5fba9dcec7a70ba0f3` |
| Complete source review ZIP v20 | `181e86d8257af9556a52e90a866001efceecc347c57d6b2c4722ea30de131df8` |

## License and provenance

GPL-3.0. See `LICENSE` and `XPANEL-NOTICE.md` in the complete source ZIP, `XPANEL-CHANGES.json` and `XPANEL-SHA256SUMS`. Upstream: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Xpanel adaptation credit: **Eng.Abdelrahman Rabie 2026**. Upstream authorship and third-party licenses remain intact.
