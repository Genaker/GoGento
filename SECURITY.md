# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

The GoGento team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisories** (Recommended)
   - Go to the [Security tab](https://github.com/Genaker/GoGento/security/advisories) of this repository
   - Click "Report a vulnerability"
   - Fill out the form with details

2. **Email**
   - Send an email to the repository maintainers
   - Include as much information as possible (see below)

### What to Include in Your Report

Please include the following information to help us better understand the nature and scope of the issue:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- **Acknowledgment**: We'll acknowledge your report within 48 hours
- **Updates**: We'll keep you informed about our progress
- **Timeline**: We aim to resolve critical security issues within 7-14 days
- **Credit**: We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

When deploying GoGento in production:

### Environment Variables
- Never commit `.env` files to version control
- Use strong, unique passwords for `API_USER` and `API_PASS`
- Rotate credentials regularly
- Use environment-specific configurations

### Database Security
- Use dedicated database users with minimal required privileges
- Enable SSL/TLS for database connections in production
- Keep MySQL updated to the latest stable version
- Regularly backup your database

### Redis Security
- Set a strong Redis password (`REDIS_PASS`)
- Bind Redis to localhost or use firewall rules
- Enable Redis AUTH
- Consider using Redis ACLs for fine-grained access control

### Authentication
- Use `AUTH_TYPE=key` with a strong API key for production
- Implement rate limiting to prevent brute force attacks
- Consider implementing OAuth2 or JWT for more sophisticated authentication
- Use HTTPS/TLS in production

### API Security
- Always run behind a reverse proxy (nginx, Caddy, etc.) in production
- Enable HTTPS/TLS
- Implement request rate limiting
- Validate and sanitize all user input
- Use prepared statements (GORM does this by default)

### Dependency Management
- Regularly update dependencies: `go get -u ./...`
- Monitor for security advisories: `go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...`
- Review dependency changes before updating

### Logging and Monitoring
- Monitor application logs for suspicious activity
- Set up alerts for repeated authentication failures
- Log security-relevant events
- Use `GORM_LOG=off` in production to avoid logging sensitive data

### Docker Security
- Use specific version tags, not `latest`
- Run containers as non-root user
- Keep base images updated
- Scan images for vulnerabilities
- Use secrets management for sensitive data

### Network Security
- Use firewall rules to restrict access
- Implement network segmentation
- Use VPN for administrative access
- Restrict database and Redis access to application servers only

## Known Security Considerations

### CORS
The application currently uses Echo's default CORS middleware. In production:
- Configure specific allowed origins
- Avoid using wildcard (`*`) in production
- Set appropriate `Access-Control-Allow-Credentials`

### SQL Injection
GORM uses prepared statements by default, providing protection against SQL injection. However:
- Never use raw SQL with user input without parameterization
- Validate and sanitize all user input
- Use GORM's safe query methods

### Authentication
The current basic authentication is suitable for internal APIs but consider:
- Implementing OAuth2 for public APIs
- Using API keys with proper rotation policies
- Implementing multi-factor authentication for sensitive operations

## Disclosure Policy

When we receive a security bug report, we will:

1. Confirm the problem and determine affected versions
2. Audit code to find any similar problems
3. Prepare fixes for all supported versions
4. Release new security patch versions
5. Publish a security advisory

## Comments on This Policy

If you have suggestions on how this process could be improved, please submit a pull request.
