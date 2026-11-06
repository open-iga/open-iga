# local-dev

Local-only helpers for testing connectors against emulators (e.g. ministack for AWS). No core changes:
the script points chosen hostnames at a local address and forwards them to the emulator.

- `bun local-dev/forward.ts setup` runs once and asks for sudo. It trusts the cert, adds the network alias and the `/etc/hosts` entries. Rerun it after a reboot (the alias is not persistent) or when `forwards.json` changes, then restart core.
- `bun local-dev/forward.ts` starts the proxy, no sudo.
- `bun local-dev/forward.ts teardown` undoes setup. While set up, the forwarded hostnames point at this machine, so real calls to them fail unless the proxy is running.
- `forwards.json` maps hostname -> emulator URL. List every host the connector's endpoints resolve to.
- Works on macOS and Linux (Linux needs `ip` and `sysctl`; setup lowers `net.ipv4.ip_unprivileged_port_start` to 443 via a sysctl.d file; teardown removes it).
- Needs `mkcert`. Start core after the script, since Go loads trusted CAs once at startup. Run core on the host (not in Docker).
