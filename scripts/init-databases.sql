-- 创建所有微服务需要的数据库 (MySQL 版本)

-- 认证服务数据库
CREATE DATABASE IF NOT EXISTS k8s_agent_auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 监控服务数据库
CREATE DATABASE IF NOT EXISTS monitor_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 集群服务数据库
CREATE DATABASE IF NOT EXISTS cluster_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Agent 管理服务数据库
CREATE DATABASE IF NOT EXISTS agent_manager_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 编排服务数据库
CREATE DATABASE IF NOT EXISTS orchestrator_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 授权 (MySQL 使用不同的授权语法)
GRANT ALL PRIVILEGES ON k8s_agent_auth.* TO 'mysql'@'%';
GRANT ALL PRIVILEGES ON monitor_db.* TO 'mysql'@'%';
GRANT ALL PRIVILEGES ON cluster_db.* TO 'mysql'@'%';
GRANT ALL PRIVILEGES ON agent_manager_db.* TO 'mysql'@'%';
GRANT ALL PRIVILEGES ON orchestrator_db.* TO 'mysql'@'%';
FLUSH PRIVILEGES;
