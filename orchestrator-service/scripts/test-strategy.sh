#!/bin/bash

# Script to create a test strategy in the database

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Creating test strategy in PostgreSQL..."

# Port forward if needed
if ! nc -z localhost 5432 2>/dev/null; then
    echo "PostgreSQL not accessible on localhost:5432"
    echo "Please run: make forward-dev-postgres (in deployments/kustomize directory)"
    exit 1
fi

# Create test strategy SQL
cat <<'EOF' | PGPASSWORD=dev-postgres-password psql -h localhost -U postgres -d aetherius_orchestrator

-- Insert a test workflow first
INSERT INTO workflows (id, name, description, trigger_type, trigger_config, steps, status, priority, timeout, metadata, created_at, updated_at)
VALUES (
    'wf-pod-failure-diagnostic',
    'Pod Failure Diagnostic Workflow',
    'Diagnose and remediate pod failures',
    'event',
    '{"event_types": ["pod_failure"]}',
    '[
        {
            "id": "step-1",
            "type": "command",
            "name": "Describe Pod",
            "description": "Get pod details",
            "config": {"command": "kubectl describe pod {{.pod_name}} -n {{.namespace}}"},
            "timeout": "30s",
            "on_success": ["step-2"]
        },
        {
            "id": "step-2",
            "type": "ai_analysis",
            "name": "Analyze Pod Logs",
            "description": "AI analysis of pod logs",
            "config": {"analysis_type": "pod_failure"},
            "timeout": "60s",
            "on_success": ["step-3"]
        },
        {
            "id": "step-3",
            "type": "remediation",
            "name": "Auto Remediation",
            "description": "Apply recommended fixes",
            "config": {"auto_apply": false},
            "timeout": "30s"
        }
    ]',
    'active',
    100,
    300000000000,
    '{}',
    NOW(),
    NOW()
) ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

-- Insert a matching strategy
INSERT INTO strategies (id, name, category, description, symptoms, workflow_id, priority, enabled, metadata, created_at, updated_at)
VALUES (
    'strat-pod-crashloop',
    'Pod CrashLoopBackOff Strategy',
    'pod_failure',
    'Strategy for handling pods in CrashLoopBackOff state',
    '[
        {
            "type": "event",
            "pattern": "pod_failure",
            "conditions": {
                "severity": "critical",
                "reason": "CrashLoopBackOff"
            }
        }
    ]',
    'wf-pod-failure-diagnostic',
    100,
    true,
    '{"auto_remediation": false}',
    NOW(),
    NOW()
) ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

SELECT 'Test strategy created successfully!' as status;

-- Show created records
SELECT id, name, status FROM workflows WHERE id = 'wf-pod-failure-diagnostic';
SELECT id, name, enabled FROM strategies WHERE id = 'strat-pod-crashloop';

EOF

echo ""
echo "Test strategy setup complete!"
echo ""
echo "To test the orchestrator:"
echo "1. Make sure NATS is running: make dev-nats (in deployments/kustomize)"
echo "2. Make sure orchestrator is running: make run"
echo "3. Send test event: ./scripts/test-event.sh"
