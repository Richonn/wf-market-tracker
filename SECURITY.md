# Security Policy

## Supported versions

Only the latest commit on `main` is supported.

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/Richonn/wf-market-tracker/security/advisories/new).

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix if you have one

You will receive a response within 72 hours.

## Scope

This project is a personal tool. The public-facing surface is limited to:
- `GET /health`
- `GET /items`
- `GET /items/{slug}/candles`
- `GET /movers`
- Static files served from `/web`

The database, Grafana, Uptime Kuma, and Dozzle are not publicly exposed.
