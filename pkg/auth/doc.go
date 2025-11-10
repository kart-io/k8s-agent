// Package auth provides reusable authentication and authorization utilities
// for the Aetherius (k8s-agent) project.
//
// This package contains business logic specific to authentication and authorization
// that can be shared across multiple services within the project. It includes:
//
// - JWT token generation and validation (pkg/auth/jwt)
// - Password cryptography using bcrypt (pkg/auth/crypto)
// - Input validation functions (pkg/auth/validator)
// - Session types and models (pkg/auth/types)
//
// # Migration from internal/auth
//
// This package was created as part of the AUTH_TO_PKG_MIGRATION (Phase 1) to extract
// reusable components from internal/auth/ into pkg/auth/. This enables other services
// in the project to use these authentication utilities without depending on the auth
// service's internal implementation.
//
// # Package Organization
//
// - pkg/auth/jwt: JWT token operations (generation, validation, token pairs)
// - pkg/auth/crypto: Password hashing and verification using bcrypt
// - pkg/auth/validator: Input validation (username, password, email, phone, UUID)
// - pkg/auth/types: Session-related types (SessionInfo, RevokedSession)
//
// # Design Philosophy
//
// This package follows the project's code organization principles:
//
// - pkg/ contains business logic specific to the Aetherius project
// - common/ contains generic utilities that ANY Go project can use
//
// The auth package is placed in pkg/ rather than common/ because:
// 1. It contains business logic specific to our authentication domain
// 2. It uses project-specific types and configurations
// 3. It's tightly coupled with our JWT and session management requirements
//
// # Backward Compatibility
//
// To maintain backward compatibility during the migration, internal/auth/types
// re-exports SessionInfo and RevokedSession from pkg/auth/types using type aliases.
// This allows existing code to continue using internal/auth/types without changes.
//
// # Usage Example
//
//	import (
//	    "github.com/kart-io/k8s-agent/pkg/auth/jwt"
//	    "github.com/kart-io/k8s-agent/pkg/auth/crypto"
//	    "github.com/kart-io/k8s-agent/pkg/auth/validator"
//	)
//
//	// Generate JWT tokens
//	tokenPair, err := jwt.GenerateTokenPair(userID, username, secret, expiresHours)
//
//	// Validate password
//	if err := validator.ValidatePassword(password); err != nil {
//	    return err
//	}
//
//	// Hash password
//	hashedPassword, err := crypto.HashPassword(password)
//
// # Migration Status
//
// Phase 1 (Completed):
// - JWT operations migrated to pkg/auth/jwt
// - Password crypto migrated to pkg/auth/crypto
// - Validators migrated to pkg/auth/validator
// - Session types migrated to pkg/auth/types
// - All imports in internal/auth updated
// - Backward compatibility maintained
// - All tests passing
//
// Future Phases:
// - Phase 2: Email client, Query filter builder
// - Phase 3: Session repository interface
// - Phase 4: Additional type definitions
package auth
