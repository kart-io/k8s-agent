-- 创建所有微服务需要的数据库

-- 认证服务数据库
CREATE DATABASE k8s_agent_auth;

-- 监控服务数据库
CREATE DATABASE monitor_db;

-- 集群服务数据库
CREATE DATABASE cluster_db;

-- Agent 管理服务数据库 (如果还没有)
SELECT 'CREATE DATABASE agent_manager_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'agent_manager_db')\gexec

-- 编排服务数据库 (如果还没有)
SELECT 'CREATE DATABASE orchestrator_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'orchestrator_db')\gexec

-- 授权
GRANT ALL PRIVILEGES ON DATABASE k8s_agent_auth TO postgres;
GRANT ALL PRIVILEGES ON DATABASE monitor_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE cluster_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE agent_manager_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE orchestrator_db TO postgres;
