# Research & Technology Decisions: Auth Service

**Feature**: Authentication and Authorization Service
**Date**: 2025-10-09
**Status**: Based on existing implementation

## Executive Summary

This document captures the technology decisions and design rationale for the auth-service. The implementation follows industry-standard patterns for authentication and authorization in microservices architecture.

## Technology Stack Decisions

### 1. Web Framework: Gin

**Decision**: Use Gin (v1.9.1) as the HTTP web framework

**Rationale**:
- High performance (40x faster than Martini according to benchmarks)
- Mature ecosystem with extensive middleware support
- Native JSON validation and binding
- Excellent documentation and community support
- Low memory footprint suitable for microservices

**Alternatives Considered**:
- **Echo**: Similar performance, but Gin has larger community
- **Fiber**: Fastest, but uses fasthttp which has compatibility issues
- **Chi**: More minimalist, but requires more custom middleware

**References**:
- Gin GitHub: https://github.com/gin-gonic/gin
- Performance benchmarks: https://github.com/gin-gonic/gin#benchmarks

### 2. Authentication: JWT (JSON Web Tokens)

**Decision**: Use golang-jwt/jwt v5 for token-based authentication

**Rationale**:
- Stateless authentication suitable for distributed systems
- Industry standard (RFC 7519)
- No server-side session storage required
- Supports token expiration and refresh mechanisms
- Easy integration with frontend frameworks

**Alternatives Considered**:
- **Session-based auth**: Requires sticky sessions in distributed env
- **OAuth 2.0**: Overkill for internal services, planned for future
- **Paseto**: More secure but less widely adopted

**Implementation Details**:
- HS256 algorithm for token signing
- 24-hour token expiration (configurable)
- Redis blacklist for logout/revocation
- Refresh token mechanism (future enhancement)

**References**:
- JWT RFC: https://tools.ietf.org/html/rfc7519
- golang-jwt: https://github.com/golang-jwt/jwt

### 3. Password Hashing: bcrypt

**Decision**: Use bcrypt from golang.org/x/crypto for password hashing

**Rationale**:
- Designed specifically for password hashing
- Adaptive cost factor (future-proof against hardware improvements)
- Built-in salt generation
- OWASP recommended
- Resistant to rainbow table and brute force attacks

**Alternatives Considered**:
- **scrypt**: More memory-hard, but bcrypt sufficient for current needs
- **argon2**: Winner of PHC, but bcrypt more widely tested
- **PBKDF2**: Older standard, not as resistant to GPU attacks

**Implementation**:
- Default cost factor: 10 (balance between security and performance)
- Automatic salt generation per password
- Password validation in constant time to prevent timing attacks

**References**:
- OWASP Password Storage: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- bcrypt specification: https://en.wikipedia.org/wiki/Bcrypt

### 4. Database: PostgreSQL

**Decision**: Use PostgreSQL as primary data store

**Rationale**:
- ACID compliance for critical auth data
- Excellent support for complex queries (permission inheritance)
- JSON/JSONB support for flexible permission data
- Mature replication and backup solutions
- Strong consistency guarantees

**Alternatives Considered**:
- **MySQL**: Less sophisticated query optimizer
- **MongoDB**: NoSQL not suitable for relational RBAC model
- **SQLite**: Not suitable for concurrent access in production

**Schema Design**:
- Normalized RBAC model (users, roles, permissions)
- Junction tables for many-to-many relationships
- UUID primary keys for distributed systems
- Indexed foreign keys for performance
- Timestamps for audit trail

**References**:
- PostgreSQL docs: https://www.postgresql.org/docs/
- RBAC patterns: https://en.wikipedia.org/wiki/Role-based_access_control

### 5. Cache Layer: Redis

**Decision**: Use Redis for token caching and blacklist

**Rationale**:
- In-memory performance for frequent auth checks
- TTL support for automatic token cleanup
- Atomic operations for blacklist management
- Pub/Sub for distributed cache invalidation (future)
- Session management capabilities

**Alternatives Considered**:
- **Memcached**: No TTL per-key, less feature-rich
- **In-memory cache**: Not suitable for multi-instance deployment
- **Hazelcast**: Overkill for current requirements

**Use Cases**:
- JWT token blacklist (logout/revocation)
- Rate limiting counters
- Session data caching
- Permission cache (future optimization)

**References**:
- Redis documentation: https://redis.io/documentation
- Redis best practices: https://redis.io/topics/optimization

### 6. Logging: Logrus

**Decision**: Use Logrus for structured logging

**Rationale**:
- Structured logging with field support
- Multiple output formats (JSON for production)
- Hook system for log aggregation
- Compatible with standard library logger
- Mature and stable

**Alternatives Considered**:
- **Zap**: Faster but more complex API
- **Zerolog**: Fastest but newer/less proven
- **Standard log**: Lacks structured logging

**Configuration**:
- JSON format in production for log aggregation
- Text format in development for readability
- Log levels: Debug, Info, Warn, Error, Fatal
- Correlation IDs for request tracing

**Future Enhancement**:
- Integration with kart-io/logger package for unified logging
- OpenTelemetry integration for distributed tracing

**References**:
- Logrus: https://github.com/sirupsen/logrus
- Structured logging best practices

## Architecture Decisions

### 7. RBAC Model Design

**Decision**: Hierarchical RBAC with three permission types

**Rationale**:
- **Menu permissions**: Frontend navigation control
- **Button permissions**: Fine-grained UI element control
- **API permissions**: Backend endpoint protection
- Flexible enough for multi-tenant scenarios
- Standard enterprise pattern

**Model Structure**:
```
User → UserRole → Role → RolePermission → Permission
```

**Permission Inheritance**:
- Permissions can have parent-child relationships
- Tree structure for hierarchical menus
- Allows for permission grouping and management

### 8. API Key Authentication

**Decision**: Separate API Key authentication for service-to-service calls

**Rationale**:
- JWT not suitable for long-lived service credentials
- API Keys more appropriate for automation
- Separate from user authentication flow
- Industry standard (similar to AWS, GitHub)

**Implementation**:
- Key + Secret pair generation
- Configurable expiration
- Usage tracking (last_used_at)
- Rotation support

### 9. Security Best Practices

**Decision**: Multiple security layers implemented

**Implementation**:
- HTTPS enforcement in production (configuration)
- CORS middleware for cross-origin protection
- SQL injection prevention (parameterized queries)
- Password complexity requirements
- Rate limiting (planned)
- Audit logging (planned)

## Integration Test Strategy

**NEEDS CLARIFICATION** → Resolved:

**Decision**: Three-tier testing approach

1. **Unit Tests**:
   - Business logic in service layer
   - Password hashing/validation
   - Token generation/validation
   - Mock dependencies

2. **Integration Tests**:
   - Database operations with test database
   - Redis operations with test instance
   - Full authentication flow
   - Permission check scenarios

3. **Contract Tests**:
   - API endpoint contracts
   - Request/response validation
   - Error handling scenarios

**Tools**:
- Go testing framework (`go test`)
- Testify for assertions and mocks
- Test containers for database integration tests
- HTTP test server for handler testing

## API Documentation Strategy

**NEEDS CLARIFICATION** → Resolved:

**Decision**: OpenAPI 3.0 specification

**Rationale**:
- Industry standard for REST APIs
- Swagger UI for interactive documentation
- Client SDK generation capability
- Contract-first development support

**Implementation**:
- Separate YAML files per resource (auth, users, roles, permissions)
- Swagger UI hosted at `/docs` endpoint
- Automated validation against implementation
- Version tracking

## Observability Enhancement

**NEEDS CLARIFICATION** → Resolved:

**Decision**: Three-pillar observability

1. **Logging**: Structured logs with Logrus (already implemented)
   - Request/response logging
   - Security event logging
   - Error logging with stack traces

2. **Metrics** (planned):
   - Prometheus metrics export
   - Authentication success/failure rates
   - API latency histograms
   - Active user counts

3. **Tracing** (planned):
   - OpenTelemetry integration
   - Distributed request tracing
   - Service dependency mapping
   - Performance bottleneck identification

**Integration**:
- Consider migrating to kart-io/logger for unified approach
- OTLP export for centralized observability platform

## Performance Considerations

### Caching Strategy

**Decision**: Multi-level caching

1. **Application Level**:
   - Permission lookup caching
   - Role resolution caching
   - Short-lived in-memory cache

2. **Redis Level**:
   - JWT validation results
   - User session data
   - Permission trees

### Database Optimization

**Decisions**:
- Indexed columns: username, email, role codes, permission codes
- Connection pooling (configured in PostgreSQL driver)
- Prepared statement caching
- Query optimization for permission trees

### Horizontal Scaling

**Considerations**:
- Stateless design allows multiple instances
- Redis shared state for token blacklist
- Database connection pooling per instance
- Load balancer compatibility (no session affinity needed)

## Future Enhancements

### Planned Features

1. **OAuth 2.0 Support**:
   - Authorization code flow
   - Client credentials flow
   - Integration with external identity providers

2. **Single Sign-On (SSO)**:
   - SAML 2.0 support
   - LDAP/Active Directory integration
   - Social login providers

3. **Multi-Factor Authentication (MFA)**:
   - TOTP (Time-based OTP)
   - SMS verification
   - Email verification
   - Backup codes

4. **Audit Logging**:
   - Comprehensive activity logs
   - Security event tracking
   - Compliance reporting
   - Log retention policies

### Technical Debt

1. Integration test suite completion
2. OpenAPI specification generation
3. Metrics and tracing implementation
4. Migration to kart-io/logger
5. Rate limiting implementation
6. Token refresh mechanism

## References

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [golang-jwt](https://github.com/golang-jwt/jwt)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [RBAC Best Practices](https://en.wikipedia.org/wiki/Role-based_access_control)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [OpenAPI Specification](https://swagger.io/specification/)
- [OpenTelemetry](https://opentelemetry.io/)

## Approval Status

- [x] Technology stack finalized
- [x] Architecture patterns defined
- [ ] Integration test strategy approved
- [ ] OpenAPI specification to be generated
- [ ] Observability plan to be implemented

**Next Steps**: Proceed to Phase 1 - Data Model and API Contracts
