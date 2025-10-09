package storage

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/kart-io/k8s-agent/auth-service/pkg/crypto"
)

// AutoMigrate creates all tables with proper indexes
func AutoMigrate(db *sql.DB) error {
	// Create users table
	if err := createUsersTable(db); err != nil {
		return err
	}

	// Create roles table
	if err := createRolesTable(db); err != nil {
		return err
	}

	// Create permissions table
	if err := createPermissionsTable(db); err != nil {
		return err
	}

	// Create user_roles junction table
	if err := createUserRolesTable(db); err != nil {
		return err
	}

	// Create role_permissions junction table
	if err := createRolePermissionsTable(db); err != nil {
		return err
	}

	// Create api_keys table
	if err := createAPIKeysTable(db); err != nil {
		return err
	}

	return nil
}

// Seed inserts default data
func Seed(db *sql.DB) error {
	// Create default admin user
	adminID := uuid.New().String()
	hashedPassword, err := crypto.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, password, email, real_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (username) DO NOTHING
	`, adminID, "admin", hashedPassword, "admin@example.com", "System Administrator", 1)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default roles
	superAdminID := uuid.New().String()
	adminRoleID := uuid.New().String()
	userRoleID := uuid.New().String()

	roles := []struct {
		id, name, code, description string
		sort                        int
	}{
		{superAdminID, "超级管理员", "super_admin", "Full system access", 1},
		{adminRoleID, "管理员", "admin", "Administrative access", 2},
		{userRoleID, "普通用户", "user", "Standard user access", 3},
	}

	for _, role := range roles {
		_, err := db.Exec(`
			INSERT INTO roles (id, name, code, description, status, sort, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (code) DO NOTHING
		`, role.id, role.name, role.code, role.description, 1, role.sort)
		if err != nil {
			return fmt.Errorf("failed to create role %s: %w", role.code, err)
		}
	}

	// Assign super_admin role to admin user
	_, err = db.Exec(`
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE code = $2
		ON CONFLICT DO NOTHING
	`, adminID, "super_admin")
	if err != nil {
		return fmt.Errorf("failed to assign role to admin: %w", err)
	}

	// Create default permissions
	if err := seedPermissions(db, superAdminID); err != nil {
		return err
	}

	return nil
}

func createUsersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(100) UNIQUE,
		real_name VARCHAR(50),
		phone VARCHAR(20),
		avatar VARCHAR(255),
		status INTEGER DEFAULT 1,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	`
	_, err := db.Exec(query)
	return err
}

func createRolesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS roles (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL,
		code VARCHAR(50) UNIQUE NOT NULL,
		description TEXT,
		status INTEGER DEFAULT 1,
		sort INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_roles_sort ON roles(sort);
	`
	_, err := db.Exec(query)
	return err
}

func createPermissionsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS permissions (
		id VARCHAR(36) PRIMARY KEY,
		parent_id VARCHAR(36),
		name VARCHAR(100) NOT NULL,
		code VARCHAR(100) UNIQUE NOT NULL,
		type VARCHAR(20) NOT NULL,
		path VARCHAR(255),
		method VARCHAR(10),
		component VARCHAR(255),
		icon VARCHAR(50),
		sort INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		description TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		FOREIGN KEY (parent_id) REFERENCES permissions(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_permissions_parent_id ON permissions(parent_id);
	CREATE INDEX IF NOT EXISTS idx_permissions_type ON permissions(type);
	CREATE INDEX IF NOT EXISTS idx_permissions_status_sort ON permissions(status, sort);
	`
	_, err := db.Exec(query)
	return err
}

func createUserRolesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id VARCHAR(36) NOT NULL,
		role_id VARCHAR(36) NOT NULL,
		PRIMARY KEY (user_id, role_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
	);
	CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
	`
	_, err := db.Exec(query)
	return err
}

func createRolePermissionsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS role_permissions (
		role_id VARCHAR(36) NOT NULL,
		permission_id VARCHAR(36) NOT NULL,
		PRIMARY KEY (role_id, permission_id),
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
		FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
	`
	_, err := db.Exec(query)
	return err
}

func createAPIKeysTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		key VARCHAR(255) UNIQUE NOT NULL,
		secret VARCHAR(255) NOT NULL,
		user_id VARCHAR(36),
		description TEXT,
		expires_at TIMESTAMP,
		status INTEGER DEFAULT 1,
		last_used_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
	CREATE INDEX IF NOT EXISTS idx_api_keys_expires_status ON api_keys(expires_at, status);
	`
	_, err := db.Exec(query)
	return err
}

func seedPermissions(db *sql.DB, superAdminRoleID string) error {
	// System management menu (root)
	systemID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO permissions (id, parent_id, name, code, type, path, component, icon, sort, status, created_at, updated_at)
		VALUES ($1, NULL, 'System Management', 'system:menu', 'menu', '/system', 'SystemLayout', 'setting', 1, 1, NOW(), NOW())
		ON CONFLICT (code) DO NOTHING
	`, systemID)
	if err != nil {
		return err
	}

	// User management submenu
	userMenuID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO permissions (id, parent_id, name, code, type, path, component, icon, sort, status, created_at, updated_at)
		VALUES ($1, $2, 'User Management', 'user:menu', 'menu', '/system/users', 'UserManagement', 'user', 1, 1, NOW(), NOW())
		ON CONFLICT (code) DO NOTHING
	`, userMenuID, systemID)
	if err != nil {
		return err
	}

	// API permissions for user CRUD
	userPerms := []struct {
		name, code, path, method string
		sort                     int
	}{
		{"View Users", "user:list", "/api/v1/users", "GET", 1},
		{"Create User", "user:create", "/api/v1/users", "POST", 2},
		{"Update User", "user:update", "/api/v1/users/:id", "PUT", 3},
		{"Delete User", "user:delete", "/api/v1/users/:id", "DELETE", 4},
	}

	for _, perm := range userPerms {
		permID := uuid.New().String()
		_, err = db.Exec(`
			INSERT INTO permissions (id, parent_id, name, code, type, path, method, sort, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'api', $5, $6, $7, 1, NOW(), NOW())
			ON CONFLICT (code) DO NOTHING
		`, permID, userMenuID, perm.name, perm.code, perm.path, perm.method, perm.sort)
		if err != nil {
			return err
		}

		// Assign to super_admin role
		_, err = db.Exec(`
			INSERT INTO role_permissions (role_id, permission_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, superAdminRoleID, permID)
		if err != nil {
			return err
		}
	}

	return nil
}
