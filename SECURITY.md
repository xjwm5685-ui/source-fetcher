# 🔒 Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.1   | :white_check_mark: |
| 1.0.0   | :x: (Upgrade to 1.0.1) |
| < 1.0   | :x:                |

## Security Enhancements in v1.0.1

Version 1.0.1 includes critical security fixes. **All users should upgrade immediately.**

### Fixed Vulnerabilities

1. **CORS Vulnerability (High)**
   - **Issue**: Web GUI accepted requests from any origin
   - **Impact**: Remote attackers could control local instance via malicious websites
   - **Fix**: CORS now restricted to localhost origins only
   - **Affected**: v1.0.0 and earlier

2. **Command Injection Risk (High)**
   - **Issue**: Package names not validated before passing to shell commands
   - **Impact**: Malicious package names could inject arbitrary commands
   - **Fix**: Added comprehensive input validation
   - **Affected**: v1.0.0 and earlier

3. **SSRF Vulnerability (High)**
   - **Issue**: URL dependencies could access private networks
   - **Impact**: Internal network resources could be accessed via SSRF
   - **Fix**: Private IP addresses now blocked by default
   - **Affected**: v1.0.0 and earlier

## Security Features

### CORS Protection
Web GUI only accepts requests from:
- `http://localhost:8765`
- `http://127.0.0.1:8765`

Can be configured in `.source-fetcher.yaml`:
```yaml
gui:
  allowed_origins:
    - http://localhost:8765
    - http://127.0.0.1:8765
```

**Warning**: Do not add external origins unless absolutely necessary.

### Input Validation
All package names and identifiers are validated to ensure they:
- Contain only alphanumeric characters, dashes, underscores, dots, slashes, and @
- Do not contain command injection sequences
- Meet format requirements for the specific package manager

### SSRF Protection
URL dependencies are validated to prevent access to:
- Localhost (127.x.x.x, ::1)
- Private networks (10.x.x.x, 172.16-31.x.x, 192.168.x.x)
- Link-local addresses (169.254.x.x)

Can be configured in `.source-fetcher.yaml`:
```yaml
security:
  block_private_ips: true   # Recommended: true
  https_only: false         # Set to true for maximum security
```

### File Permissions
Sensitive files use restricted permissions:
- Manifest files: `0600` (owner read/write only)
- Report files: `0640` (owner read/write, group read)

### Safe YAML Parsing
All YAML configuration files use explicit safe decoders to prevent:
- Code execution via YAML tags
- Arbitrary object instantiation
- Resource exhaustion attacks

## Security Best Practices

### For End Users

1. **Keep Software Updated**
   - Always use the latest version
   - Subscribe to GitHub releases for security notifications

2. **Web GUI Usage**
   - Only run Web GUI on trusted networks
   - Do not expose port 8765 to the internet
   - Close Web GUI when not in use

3. **Configuration Files**
   - Store `.source-fetcher.yaml` with restricted permissions
   - Never commit files containing tokens to version control
   - Use environment variables for sensitive credentials

4. **Private Registries**
   - Store tokens in environment variables, not config files
   - Use read-only tokens when possible
   - Rotate tokens regularly

5. **URL Dependencies**
   - Only download from trusted sources
   - Verify checksums when available
   - Keep `security.block_private_ips: true`

### For Developers

1. **Input Validation**
   - All user input must be validated before use
   - Use `isValidPackageName()` for package identifiers
   - Use `parseURL()` and `isPrivateOrLocalAddress()` for URLs

2. **Context Propagation**
   - Pass `context.Context` to all long-running operations
   - Respect context cancellation
   - Use appropriate timeouts

3. **Error Handling**
   - Never expose sensitive information in error messages
   - Log errors appropriately
   - Handle edge cases explicitly

4. **Dependencies**
   - Keep Go dependencies up to date
   - Review dependency changes in PRs
   - Use `go mod verify` to check integrity

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

### Preferred Reporting Method

1. Email: [Your security contact email]
2. Subject: `[SECURITY] Source Fetcher - [Brief Description]`
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Release**: Depends on severity
  - Critical: 7 days
  - High: 14 days
  - Medium: 30 days
  - Low: Next release

### Disclosure Policy

- We will coordinate disclosure with the reporter
- Public disclosure after fix is released
- Credit will be given to the reporter (unless anonymity requested)

## Security Update Process

When a security issue is fixed:

1. **Security Patch Release**
   - Version bump (e.g., 1.0.0 → 1.0.1)
   - CHANGELOG entry with severity rating
   - GitHub Security Advisory created

2. **User Notification**
   - GitHub Release with security notice
   - Badges updated in README
   - Announcement in discussions (if severe)

3. **CVE Assignment** (for severe issues)
   - Request CVE from GitHub
   - Update advisory with CVE ID

## Security Checklist for Releases

Before each release, verify:

- [ ] All known security issues are fixed
- [ ] Dependencies are up to date
- [ ] `go vet ./...` passes
- [ ] Security-focused tests pass
- [ ] No credentials in code or configs
- [ ] CHANGELOG includes security fixes
- [ ] Security documentation is current

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE - Common Weakness Enumeration](https://cwe.mitre.org/)
- [Go Security Best Practices](https://go.dev/doc/security/best-practices)

---

**Last Updated**: 2026-06-06 (v1.0.1)
