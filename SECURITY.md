# Security Policy

## 🔒 Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## 🐛 Reporting a Vulnerability

We take the security of Source Fetcher seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please Do Not

- **Do not** open a public GitHub issue for security vulnerabilities
- **Do not** disclose the vulnerability publicly until it has been addressed

### Please Do

1. **Email us** at: ckkhua89@gmail.com
2. **Include** the following information:
   - Type of vulnerability
   - Full paths of source file(s) related to the vulnerability
   - Location of the affected source code (tag/branch/commit or direct URL)
   - Step-by-step instructions to reproduce the issue
   - Proof-of-concept or exploit code (if possible)
   - Impact of the vulnerability
   - Suggested fix (if available)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Updates**: We will send you regular updates about our progress
- **Timeline**: We aim to address critical vulnerabilities within 7 days
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

## 🛡️ Security Measures

### Current Security Features

1. **Authentication**
   - Support for Bearer tokens
   - Basic authentication
   - Environment variable-based credential storage
   - No plaintext passwords in config files

2. **Downloads**
   - HTTPS by default
   - Checksum verification (where available)
   - Timeout protection
   - Size limit validation

3. **Script Execution**
   - Lifecycle scripts disabled by default
   - Whitelist-based script execution
   - Configurable script policies (none/root/all)

4. **Input Validation**
   - Sanitized user input
   - Version string validation
   - Path traversal protection
   - Command injection prevention

### Planned Security Enhancements

- [ ] Package signature verification
- [ ] Sandboxed script execution
- [ ] Vulnerability scanning integration
- [ ] Security audit logging
- [ ] Rate limiting for API requests
- [ ] Content Security Policy for downloads

## 🔐 Security Best Practices

### For Users

1. **Keep Updated**
   - Always use the latest version
   - Subscribe to security advisories
   - Review changelogs for security fixes

2. **Credentials**
   - Use environment variables for tokens
   - Never commit credentials to version control
   - Rotate tokens regularly
   - Use least-privilege access

3. **Script Execution**
   - Keep scripts disabled unless necessary
   - Review packages before allowing scripts
   - Use whitelist for trusted packages
   - Monitor script execution logs

4. **Network Security**
   - Use HTTPS mirrors when possible
   - Verify checksums after download
   - Use VPN in untrusted networks
   - Configure firewall rules appropriately

5. **Configuration**
   - Protect config files with appropriate permissions
   - Review auth profiles regularly
   - Use separate profiles for different environments
   - Audit configuration changes

### For Contributors

1. **Code Review**
   - All code changes require review
   - Security-sensitive changes require additional review
   - Use static analysis tools
   - Follow secure coding guidelines

2. **Dependencies**
   - Keep dependencies updated
   - Review dependency changes
   - Use dependency scanning tools
   - Minimize dependency count

3. **Testing**
   - Write security tests
   - Test input validation
   - Test authentication flows
   - Test error handling

## 🚨 Known Security Considerations

### Lifecycle Scripts

**Risk**: npm packages can execute arbitrary code through lifecycle scripts (preinstall, install, postinstall).

**Mitigation**:
- Scripts are disabled by default
- Whitelist mechanism for trusted packages
- Configurable script policies
- Future: Sandboxed execution

### Private Registry Authentication

**Risk**: Credentials could be exposed if not properly secured.

**Mitigation**:
- Environment variable-based storage
- No plaintext passwords in config
- Support for token-based auth
- Future: Encrypted credential storage

### Download Verification

**Risk**: Downloaded packages could be tampered with.

**Mitigation**:
- HTTPS by default
- Checksum verification (where available)
- Future: Signature verification

### Command Injection

**Risk**: User input could be used to execute arbitrary commands.

**Mitigation**:
- Input sanitization
- Parameterized command execution
- Path validation
- Shell escaping

## 📋 Security Checklist for Releases

Before each release, we verify:

- [ ] All dependencies are up to date
- [ ] Security scanning has been performed
- [ ] Known vulnerabilities have been addressed
- [ ] Security tests pass
- [ ] Documentation is updated
- [ ] Security advisories are reviewed

## 🔍 Security Audit History

| Date | Auditor | Scope | Findings | Status |
|------|---------|-------|----------|--------|
| TBD  | TBD     | TBD   | TBD      | TBD    |

## 📚 Security Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Go Security Best Practices](https://golang.org/doc/security/)
- [npm Security Best Practices](https://docs.npmjs.com/security-best-practices)

## 🏆 Security Hall of Fame

We would like to thank the following individuals for responsibly disclosing security vulnerabilities:

<!-- Security researchers will be listed here -->

## 📞 Contact

For security-related questions that are not vulnerabilities, please use:
- GitHub Discussions: [Security Category]
- Email: ckkhua89@gmail.com

---

**Last Updated:** May 31, 2026

Thank you for helping keep Source Fetcher and its users safe!
