// Local testing only (macOS and Linux): makes selected hostnames (e.g. AWS) resolve to this machine and forwards them
// to a local emulator. Core is untouched; it still dials the real hostnames over https.
//
//   bun local-dev/forward.ts setup      once (needs sudo): trust cert, alias, /etc/hosts. Rerun after a reboot or a forwards.json change
//   bun local-dev/forward.ts            start the proxy (no sudo)
//   bun local-dev/forward.ts teardown   undo setup (needs sudo)
//
// While set up, the forwarded hostnames point at this machine, so real calls to them fail when the proxy is not running.
import { existsSync } from 'node:fs';
import { join } from 'node:path';

const forwards: Record<string, string> = await Bun.file(join(import.meta.dir, 'forwards.json')).json();
const hosts = Object.keys(forwards);

// ---------- platform ----------
const isMac = process.platform === 'darwin';
if (!isMac && process.platform !== 'linux') throw new Error(`unsupported platform: ${process.platform}`);

// sudo prompts on the terminal itself, and Bun.$ throws on a non-zero exit
const sudo = (...cmd: string[]) => Bun.$`sudo ${cmd}`;

// ---------- certificate ----------
// Core only talks https and verifies the chain, so the proxy needs a cert core trusts (mkcert's local CA).
const certs = {
    dir: join(import.meta.dir, '.certs'),
    cert: join(import.meta.dir, '.certs', 'cert.pem'),
    key: join(import.meta.dir, '.certs', 'key.pem'),
    async create() {
        await Bun.$`mkdir -p ${this.dir}`;
        // adds mkcert's CA to the system trust store; Go loads roots once, so (re)start core afterwards
        await Bun.$`mkcert -install`;
        await Bun.$`mkcert -cert-file ${this.cert} -key-file ${this.key} ${hosts}`.quiet();
    },
};

// ---------- network ----------
// 198.18.0.0/15 is not loopback/private, so core's SSRF guard (isPublicIP) lets it through.
const ALIAS_IP = '198.18.0.1';
const HOSTS_MARKER = '# open-iga-local'; // tags our lines in /etc/hosts so they can be removed
// Linux only: binding 443 without root needs this lowered; kept in a sysctl.d file so it survives reboots
const PORT_SYSCTL = 'net.ipv4.ip_unprivileged_port_start';
const SYSCTL_FILE = '/etc/sysctl.d/99-open-iga-local.conf';

const network = {
    async up() {
        if (isMac) await sudo('ifconfig', 'lo0', 'alias', ALIAS_IP);
        else await sudo('ip', 'addr', 'add', `${ALIAS_IP}/32`, 'dev', 'lo');

        const lines = hosts.map((host) => `${ALIAS_IP} ${host} ${HOSTS_MARKER}`).join('\n');
        await sudo(
            'sh',
            '-c',
            `sed -i${isMac ? " ''" : ''} '/${HOSTS_MARKER}/d' /etc/hosts && printf '%s\\n' '${lines}' >> /etc/hosts`,
        );

        if (!isMac) {
            await sudo('sh', '-c', `echo '${PORT_SYSCTL}=443' > ${SYSCTL_FILE}`);
            await sudo('sysctl', '-w', `${PORT_SYSCTL}=443`);
        }
    },

    async down() {
        await sudo('sed', '-i', ...(isMac ? [''] : []), `/${HOSTS_MARKER}/d`, '/etc/hosts');
        const removeAlias = isMac
            ? sudo('ifconfig', 'lo0', '-alias', ALIAS_IP)
            : sudo('ip', 'addr', 'del', `${ALIAS_IP}/32`, 'dev', 'lo');
        await removeAlias.catch(() => {});
        if (!isMac) {
            await sudo('rm', '-f', SYSCTL_FILE);
            await sudo('sysctl', '-w', `${PORT_SYSCTL}=1024`); // the kernel default
        }
    },
};

// ---------- proxy ----------
const startProxy = () =>
    Bun.serve({
        // macOS lets non-root bind port 443 only on all addresses, not on a specific one (reported as EADDRINUSE)
        hostname: isMac ? '0.0.0.0' : ALIAS_IP,
        port: 443,
        tls: { cert: Bun.file(certs.cert), key: Bun.file(certs.key) },
        async fetch(request) {
            const incoming = new URL(request.url);
            const upstream = forwards[incoming.hostname];
            if (!upstream) {
                console.error(`no forward for ${incoming.hostname}: add it to forwards.json and rerun setup`);
                return new Response(`no forward for ${incoming.hostname}`, { status: 502 });
            }
            console.log(`${request.method} ${incoming.hostname}${incoming.pathname} -> ${upstream}`);

            return fetch(new URL(incoming.pathname + incoming.search, upstream), {
                method: request.method,
                headers: request.headers,
                body: request.body,
                redirect: 'manual',
            });
        },
    });

// ---------- run ----------
const command = process.argv[2] ?? 'start';

if (command === 'setup') {
    await sudo('-v'); // prompt for the password once up front
    await certs.create();
    await network.up();
    console.log('setup done. Restart core, then run: bun local-dev/forward.ts');
} else if (command === 'teardown') {
    await sudo('-v');
    await network.down();
    console.log('teardown done.');
} else if (command === 'start') {
    if (!existsSync(certs.cert)) throw new Error('not set up yet: run `bun local-dev/forward.ts setup` first');
    startProxy();
    console.log(`forwarding ${hosts.join(', ')} via ${ALIAS_IP}:443. Ctrl+C to stop.`);
} else {
    throw new Error(`unknown command "${command}" (setup | start | teardown)`);
}
