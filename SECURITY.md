# Security Policy

## 🔒 Reporting Security Vulnerabilities

If you discover a security vulnerability in the ILEX Backend, please report it responsibly:

1. **DO NOT** create a public issue
2. Send an email to: [security@ilex.com] (replace with your actual email)
3. Include a detailed description of the vulnerability
4. Provide steps to reproduce if possible

We will acknowledge receipt within 48 hours and provide regular updates.

## 🛡️ Security Best Practices

### Environment Configuration

- ✅ **Never commit .env files** - They are automatically ignored by git
- ✅ **Use strong JWT secrets** - Minimum 32 characters, random strings
- ✅ **Rotate secrets regularly** - Especially in production
- ✅ **Use HTTPS in production** - Never send credentials over HTTP

### Authentication & Authorization

- ✅ **OTP expiration** - Default 5 minutes, configurable
- ✅ **JWT token expiration** - Default 24 hours, with refresh tokens
- ✅ **Role-based access control** - Strict role validation on endpoints
- ✅ **Phone number validation** - Server-side validation required

### Database Security

- ✅ **SurrealDB authentication** - Always use credentials
- ✅ **Namespace isolation** - Use dedicated namespace per environment
- ✅ **Connection encryption** - Use WSS in production
- ✅ **Query parameterization** - Always use parameters, never string concatenation

### API Security

- ✅ **Input validation** - All endpoints validate input
- ✅ **CORS configuration** - Restrict origins in production
- ✅ **Rate limiting** - Implement rate limiting (TODO: currently placeholder)
- ✅ **Request size limits** - Prevent large payload attacks

## 🚨 Security Checklist for Production

### Before Deployment

- [ ] Change all default passwords
- [ ] Generate new JWT secrets (32+ characters)
- [ ] Configure SMS/Email providers with production keys
- [ ] Set `ENVIRONMENT=production`
- [ ] Enable HTTPS/WSS for all connections
- [ ] Configure proper CORS origins
- [ ] Set up monitoring and logging
- [ ] Review all environment variables

### Infrastructure

- [ ] Use secure hosting (cloud providers with security compliance)
- [ ] Configure firewalls (restrict SurrealDB access)
- [ ] Set up SSL certificates
- [ ] Enable security headers
- [ ] Configure backup strategies
- [ ] Set up monitoring and alerting

## 🔐 Sensitive Data Handling

### What NOT to commit:

- ❌ `.env` files with real credentials
- ❌ API keys or secrets
- ❌ Database passwords
- ❌ JWT secrets
- ❌ SMS/Email provider credentials
- ❌ Production configuration

### What IS safe to commit:

- ✅ `.env.example` with placeholder values
- ✅ Configuration structure without secrets
- ✅ Default development settings
- ✅ Documentation and code

## 🛠️ Security Tools & Practices

### Code Security

```bash
# Run security audit
go list -json -m all | nancy sleuth

# Check for vulnerabilities
govulncheck ./...

# Static analysis
staticcheck ./...
```

### Dependencies

- Keep Go version updated
- Regularly update dependencies: `go get -u ./...`
- Audit dependencies for known vulnerabilities
- Use minimal dependencies

## 📋 Security Features Implemented

- **JWT Authentication** with HS256 signing
- **Refresh Token** rotation
- **OTP Validation** with expiration
- **Role-based Access Control** (CLIENT/LIVREUR/ADMIN)
- **Input Validation** with struct tags
- **Password-free Authentication** (phone + OTP only)
- **CORS Protection** middleware
- **Request ID Tracking** for audit logs

## 🔍 Known Security Considerations

### Current Limitations

1. **Rate Limiting**: Placeholder implementation - needs Redis/memory store
2. **Session Management**: Relies on JWT stateless tokens
3. **Audit Logging**: Basic logging - consider structured logging
4. **API Versioning**: Single version - plan for backward compatibility

### Future Enhancements

- [ ] Implement proper rate limiting with Redis
- [ ] Add API versioning strategy  
- [ ] Enhanced audit logging with structured data
- [ ] Add request/response sanitization
- [ ] Implement API key authentication for service-to-service
- [ ] Add webhook signature validation

## 📞 Contact

For security-related questions or concerns:
- Security Email: [security@ilex.com]
- General Contact: [contact@ilex.com]

---

**Remember**: Security is everyone's responsibility. When in doubt, ask!