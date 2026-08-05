# Security Policy

## Supported Versions

Security fixes are released against the latest published version. Please upgrade
to the newest release before reporting.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| older   | :x:                |

## Security Posture

PrivUtil is a local-first tool and hardens accordingly:

- Binds to loopback (`127.0.0.1`) by default; exposing it requires an explicit `-host 0.0.0.0`.
- Validates the `Host` header against an allowlist (DNS-rebinding defense) and scopes CORS to allowlisted origins (no wildcard).
- Serves a Content-Security-Policy plus `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, and `Referrer-Policy`.
- Caps request bodies (32 MiB) and bounds compute-heavy handlers (e.g. bcrypt cost ≤ 15, MathEval/BaseConvert/TextSimilarity input limits); recovers from handler panics.
- Release container images are multi-arch, cosign keyless-signed, and ship an SBOM and SLSA provenance.

## Reporting a Vulnerability

We take the security of PrivUtil seriously. If you have discovered a security vulnerability in this project, please report it privately. **Do not disclose it as a public issue.**

### How to Report

Preferred: open a private [GitHub security advisory](https://github.com/odinnordico/PrivUtil/security/advisories/new) for this repository. Alternatively, email **odin.nordico90@gmail.com**.

You should receive a response within 48 hours. If for some reason you do not, please follow up to ensure we received your original message.

### Process

1.  **Triage**: We will review the report and determine if it is a valid vulnerability.
2.  **Fix**: We will work on a fix in a private branch.
3.  **Release**: We will release a patch version containing the fix.
4.  **Disclosure**: Once the fix is released, we will publicly disclose the vulnerability.

Thank you for helping keep PrivUtil safe for everyone!
