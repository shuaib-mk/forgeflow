# Security policy

## Reporting

Please report suspected vulnerabilities privately through the repository's security advisory feature. Do not open a public issue for an unpatched vulnerability.

## Supported versions

Security fixes are provided for the latest release on `main`.

## Deployment boundary

ForgeFlow is local-first. Its workflow runner executes configured commands and therefore must run as an unprivileged operating-system account. Do not mount a Docker socket, SSH agent, broad home directory, or production credentials into workers. For untrusted users, deploy an external hardened runner; the built-in worker is not a multi-tenant sandbox.

Passwords use bcrypt. Session tokens are stored only as SHA-256 digests. Logs redact values whose keys indicate tokens, passwords, secrets, or authorization headers. CORS origins are explicit, repository paths are restricted to the configured workspace root, and API authorization is enforced in services.

