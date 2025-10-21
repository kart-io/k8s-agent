-- Auth Service MySQL 初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS user_auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE user_auth;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    real_name VARCHAR(50),
    phone VARCHAR(20),
    avatar VARCHAR(255),
    status INT DEFAULT 1 COMMENT '1: active, 0: disabled',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    status INT DEFAULT 1 COMMENT '1: active, 0: disabled',
    sort INT DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_code (code),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36) DEFAULT '',
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL COMMENT 'menu, button, api',
    path VARCHAR(255),
    method VARCHAR(10),
    component VARCHAR(255),
    icon VARCHAR(50),
    sort INT DEFAULT 0,
    status INT DEFAULT 1 COMMENT '1: active, 0: disabled',
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_parent_id (parent_id),
    INDEX idx_code (code),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    user_id VARCHAR(36),
    role_id VARCHAR(36),
    PRIMARY KEY (user_id, role_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(36),
    permission_id VARCHAR(36),
    PRIMARY KEY (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- API密钥表
CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    `key` VARCHAR(255) UNIQUE NOT NULL,
    secret VARCHAR(255) NOT NULL,
    user_id VARCHAR(36),
    description TEXT,
    expires_at DATETIME COMMENT '过期时间',
    status INT DEFAULT 1 COMMENT '1: active, 0: disabled',
    last_used_at DATETIME COMMENT '最后使用时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_key (`key`),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API密钥表';

-- 插入默认超级管理员角色
INSERT INTO roles (id, name, code, description, status, sort, created_at, updated_at) VALUES
('role-super-admin', '超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', 1, 1, NOW(), NOW()),
('role-admin', '管理员', 'admin', '系统管理员', 1, 2, NOW(), NOW()),
('role-user', '普通用户', 'user', '普通用户', 1, 3, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 插入默认超级管理员用户（密码: admin123，使用 bcrypt 加密）
-- 注意：这是 "admin123" 的 bcrypt hash (使用 go run scripts/hash_password.go admin123 生成)
INSERT INTO users (id, username, password, email, real_name, status, created_at, updated_at) VALUES
('user-admin', 'admin', '$2a$10$fSA0jhcFVnG..gkMmi5Ypug2YFwjVHCd7rveiwp6XUJGpYlHQ5ZSK', 'admin@example.com', '超级管理员', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE password=VALUES(password);

-- 关联超级管理员用户和角色
INSERT INTO user_roles (user_id, role_id) VALUES
('user-admin', 'role-super-admin')
ON DUPLICATE KEY UPDATE user_id=VALUES(user_id);

-- 插入默认权限 - 系统管理
INSERT INTO permissions (id, parent_id, name, code, type, path, component, icon, sort, status, created_at, updated_at) VALUES
('perm-system', '', '系统管理', 'system', 'menu', '/system', 'Layout', 'setting', 1, 1, NOW(), NOW()),
('perm-user', 'perm-system', '用户管理', 'user', 'menu', '/system/user', 'system/user/index', 'user', 1, 1, NOW(), NOW()),
('perm-user-list', 'perm-user', '用户列表', 'user:list', 'api', '/api/v1/users', 'GET', '', 1, 1, NOW(), NOW()),
('perm-user-create', 'perm-user', '创建用户', 'user:create', 'api', '/api/v1/users', 'POST', '', 2, 1, NOW(), NOW()),
('perm-user-update', 'perm-user', '更新用户', 'user:update', 'api', '/api/v1/users/*', 'PUT', '', 3, 1, NOW(), NOW()),
('perm-user-delete', 'perm-user', '删除用户', 'user:delete', 'api', '/api/v1/users/*', 'DELETE', '', 4, 1, NOW(), NOW()),
('perm-role', 'perm-system', '角色管理', 'role', 'menu', '/system/role', 'system/role/index', 'team', 2, 1, NOW(), NOW()),
('perm-role-list', 'perm-role', '角色列表', 'role:list', 'api', '/api/v1/roles', 'GET', '', 1, 1, NOW(), NOW()),
('perm-role-create', 'perm-role', '创建角色', 'role:create', 'api', '/api/v1/roles', 'POST', '', 2, 1, NOW(), NOW()),
('perm-role-update', 'perm-role', '更新角色', 'role:update', 'api', '/api/v1/roles/*', 'PUT', '', 3, 1, NOW(), NOW()),
('perm-role-delete', 'perm-role', '删除角色', 'role:delete', 'api', '/api/v1/roles/*', 'DELETE', '', 4, 1, NOW(), NOW()),
('perm-role-assign', 'perm-role', '分配权限', 'role:assign', 'api', '/api/v1/roles/*/permissions', 'POST', '', 5, 1, NOW(), NOW()),
('perm-permission', 'perm-system', '权限管理', 'permission', 'menu', '/system/permission', 'system/permission/index', 'lock', 3, 1, NOW(), NOW()),
('perm-permission-list', 'perm-permission', '权限列表', 'permission:list', 'api', '/api/v1/permissions', 'GET', '', 1, 1, NOW(), NOW()),
('perm-permission-create', 'perm-permission', '创建权限', 'permission:create', 'api', '/api/v1/permissions', 'POST', '', 2, 1, NOW(), NOW()),
('perm-permission-update', 'perm-permission', '更新权限', 'permission:update', 'api', '/api/v1/permissions/*', 'PUT', '', 3, 1, NOW(), NOW()),
('perm-permission-delete', 'perm-permission', '删除权限', 'permission:delete', 'api', '/api/v1/permissions/*', 'DELETE', '', 4, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 给超级管理员角色分配所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-super-admin', id FROM permissions
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 显示初始化完成信息
SELECT '数据库初始化完成！' AS message;
SELECT '默认管理员账号: admin' AS info;
SELECT '默认管理员密码: admin123' AS info;
