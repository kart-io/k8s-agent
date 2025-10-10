-- Cluster Service Database Initialization Script

-- Create database (run as postgres superuser)
-- CREATE DATABASE cluster_db;
-- CREATE DATABASE cluster_dev;
-- CREATE DATABASE cluster_production;

-- Connect to the database
\c cluster_db;

-- Create clusters table
CREATE TABLE IF NOT EXISTS clusters (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    endpoint VARCHAR(255) NOT NULL,
    version VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    region VARCHAR(100),
    provider VARCHAR(100),
    kubeconfig TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status);
CREATE INDEX IF NOT EXISTS idx_clusters_provider ON clusters(provider);
CREATE INDEX IF NOT EXISTS idx_clusters_region ON clusters(region);
CREATE INDEX IF NOT EXISTS idx_clusters_created_at ON clusters(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for clusters table
DROP TRIGGER IF EXISTS update_clusters_updated_at ON clusters;
CREATE TRIGGER update_clusters_updated_at
    BEFORE UPDATE ON clusters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Sample data (optional, comment out for production)
-- INSERT INTO clusters (id, name, description, endpoint, version, status, region, provider, kubeconfig)
-- VALUES
--     ('cluster-1', 'Development Cluster', 'Local development cluster', 'https://localhost:6443', 'v1.28.0', 'healthy', 'local', 'minikube', 'sample-kubeconfig'),
--     ('cluster-2', 'Staging Cluster', 'Staging environment', 'https://staging.example.com:6443', 'v1.28.0', 'healthy', 'us-west-2', 'aws', 'sample-kubeconfig');

-- Verify tables
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';

-- Show cluster table structure
\d clusters
