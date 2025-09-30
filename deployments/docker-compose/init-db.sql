-- Aetherius Database Initialization Script

-- Create databases for each service
CREATE DATABASE aetherius_agent_manager;
CREATE DATABASE aetherius_orchestrator;

-- Connect to agent_manager database
\c aetherius_agent_manager;

-- Agent Manager Tables
CREATE TABLE IF NOT EXISTS agents (
    id VARCHAR(255) PRIMARY KEY,
    cluster_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    version VARCHAR(100),
    last_heartbeat TIMESTAMP,
    connection_info JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS clusters (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    config JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    cluster_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    reason VARCHAR(255),
    message TEXT,
    namespace VARCHAR(255),
    labels JSONB,
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS commands (
    id VARCHAR(255) PRIMARY KEY,
    cluster_id VARCHAR(255) NOT NULL,
    agent_id VARCHAR(255),
    type VARCHAR(100) NOT NULL,
    tool VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    args JSONB,
    namespace VARCHAR(255),
    timeout INTEGER,
    status VARCHAR(50) NOT NULL,
    result JSONB,
    issued_by VARCHAR(255),
    issued_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_agents_cluster_id ON agents(cluster_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_events_cluster_id ON events(cluster_id);
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);
CREATE INDEX idx_events_severity ON events(severity);
CREATE INDEX idx_commands_cluster_id ON commands(cluster_id);
CREATE INDEX idx_commands_status ON commands(status);

-- Connect to orchestrator database
\c aetherius_orchestrator;

-- Orchestrator Tables
CREATE TABLE IF NOT EXISTS workflows (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(100) NOT NULL,
    trigger_config JSONB,
    status VARCHAR(50) NOT NULL,
    steps JSONB NOT NULL,
    priority INTEGER DEFAULT 0,
    timeout INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_executions (
    id VARCHAR(255) PRIMARY KEY,
    workflow_id VARCHAR(255) NOT NULL REFERENCES workflows(id),
    status VARCHAR(50) NOT NULL,
    trigger_event JSONB,
    context JSONB,
    step_executions JSONB,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    duration INTEGER,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS strategies (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    symptoms JSONB NOT NULL,
    workflow_id VARCHAR(255) NOT NULL REFERENCES workflows(id),
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflow_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX idx_workflow_executions_status ON workflow_executions(status);
CREATE INDEX idx_workflow_executions_started_at ON workflow_executions(started_at DESC);
CREATE INDEX idx_strategies_enabled ON strategies(enabled);
CREATE INDEX idx_strategies_priority ON strategies(priority DESC);