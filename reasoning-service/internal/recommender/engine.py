"""
Recommendation Engine
Provides intelligent remediation recommendations based on root cause
"""

from typing import List, Dict
from loguru import logger

from pkg.types import (
    RootCause, RootCauseType, Recommendation, RiskLevel, AnalysisContext
)


class RecommendationEngine:
    """Generate remediation recommendations"""

    def __init__(self):
        """Initialize recommendation engine"""
        self.recommendation_rules = self._load_recommendation_rules()

    def recommend(self, root_cause: RootCause, context: AnalysisContext) -> List[Recommendation]:
        """
        Generate recommendations based on root cause

        Args:
            root_cause: Identified root cause
            context: Analysis context

        Returns:
            List of recommendations sorted by confidence
        """
        logger.info(f"Generating recommendations for {root_cause.type}")

        recommendations = []

        # Get recommendations for this root cause type
        rules = self.recommendation_rules.get(root_cause.type, [])

        for rule in rules:
            # Check if rule conditions are met
            if self._check_conditions(rule, context):
                recommendation = Recommendation(
                    action=rule["action"],
                    description=rule["description"],
                    confidence=rule["confidence"] * root_cause.confidence,  # Adjust by root cause confidence
                    risk=RiskLevel(rule["risk"]),
                    impact=rule["impact"],
                    steps=rule["steps"],
                    rollback_steps=rule.get("rollback_steps"),
                    estimated_duration=rule.get("estimated_duration"),
                    metadata=rule.get("metadata", {})
                )
                recommendations.append(recommendation)

        # Sort by confidence * (inverse risk weight)
        risk_weights = {
            RiskLevel.LOW: 1.0,
            RiskLevel.MEDIUM: 0.9,
            RiskLevel.HIGH: 0.7,
            RiskLevel.CRITICAL: 0.5
        }

        recommendations.sort(
            key=lambda r: r.confidence * risk_weights[r.risk],
            reverse=True
        )

        logger.info(f"Generated {len(recommendations)} recommendations")
        return recommendations[:5]  # Top 5

    def _check_conditions(self, rule: Dict, context: AnalysisContext) -> bool:
        """Check if rule conditions are met"""
        conditions = rule.get("conditions", {})

        if not conditions:
            return True  # No conditions = always applicable

        # Check event-based conditions
        if "event_reason" in conditions and context.event:
            if context.event.get("reason") != conditions["event_reason"]:
                return False

        # Check metric-based conditions
        if "requires_metrics" in conditions and not context.metrics:
            return False

        return True

    def _load_recommendation_rules(self) -> Dict[RootCauseType, List[Dict]]:
        """Load recommendation rules"""
        return {
            RootCauseType.OOM_KILLER: [
                {
                    "action": "increase_memory_limit",
                    "description": "Increase container memory limits to prevent OOM kills",
                    "confidence": 0.95,
                    "risk": "low",
                    "impact": "Prevents future OOM kills, may increase cluster resource usage",
                    "steps": [
                        "Analyze current memory usage patterns",
                        "Calculate recommended memory limit (current + 50%)",
                        "Update Deployment/StatefulSet memory limits",
                        "kubectl apply -f updated-manifest.yaml",
                        "Monitor for OOM recurrence"
                    ],
                    "rollback_steps": [
                        "Revert to previous memory limits",
                        "kubectl rollout undo deployment/<name>"
                    ],
                    "estimated_duration": "5 minutes",
                    "metadata": {
                        "suggested_increase": "50%",
                        "monitor_period": "24h"
                    }
                },
                {
                    "action": "add_memory_request",
                    "description": "Set appropriate memory requests for better scheduling",
                    "confidence": 0.85,
                    "risk": "low",
                    "impact": "Improves pod scheduling and resource guarantees",
                    "steps": [
                        "Calculate 80th percentile memory usage",
                        "Set memory request to P80 value",
                        "Update pod spec with requests",
                        "Apply changes and monitor"
                    ],
                    "estimated_duration": "5 minutes"
                },
                {
                    "action": "optimize_application",
                    "description": "Optimize application memory usage",
                    "confidence": 0.70,
                    "risk": "medium",
                    "impact": "Reduces memory footprint but requires code changes",
                    "steps": [
                        "Profile application memory usage",
                        "Identify memory leaks or inefficiencies",
                        "Optimize code or dependencies",
                        "Test changes in staging",
                        "Deploy to production"
                    ],
                    "estimated_duration": "Several hours to days"
                }
            ],

            RootCauseType.CPU_THROTTLING: [
                {
                    "action": "increase_cpu_limit",
                    "description": "Increase CPU limits to reduce throttling",
                    "confidence": 0.90,
                    "risk": "low",
                    "impact": "Improves application performance",
                    "steps": [
                        "Analyze CPU usage patterns",
                        "Increase CPU limit by 50-100%",
                        "Update deployment manifest",
                        "Apply changes",
                        "Monitor throttling metrics"
                    ],
                    "estimated_duration": "5 minutes"
                },
                {
                    "action": "optimize_workload",
                    "description": "Optimize CPU-intensive operations",
                    "confidence": 0.75,
                    "risk": "medium",
                    "impact": "Reduces CPU usage through code optimization",
                    "steps": [
                        "Profile CPU usage",
                        "Identify hotspots",
                        "Optimize algorithms or add caching",
                        "Test performance improvements"
                    ],
                    "estimated_duration": "Hours to days"
                }
            ],

            RootCauseType.IMAGE_PULL_ERROR: [
                {
                    "action": "fix_image_reference",
                    "description": "Correct image name or tag",
                    "confidence": 0.95,
                    "risk": "low",
                    "impact": "Resolves image pull failures",
                    "steps": [
                        "Verify image exists in registry",
                        "Check image name and tag spelling",
                        "Update deployment with correct image",
                        "kubectl apply -f deployment.yaml"
                    ],
                    "estimated_duration": "2 minutes"
                },
                {
                    "action": "configure_image_pull_secret",
                    "description": "Add or update image pull secrets for private registry",
                    "confidence": 0.90,
                    "risk": "low",
                    "impact": "Enables pulling from private registries",
                    "steps": [
                        "Create docker-registry secret",
                        "kubectl create secret docker-registry regcred --docker-server=<registry> ...",
                        "Add imagePullSecrets to pod spec",
                        "Apply updated manifest"
                    ],
                    "estimated_duration": "5 minutes"
                }
            ],

            RootCauseType.CONFIG_ERROR: [
                {
                    "action": "fix_configuration",
                    "description": "Correct application configuration or environment variables",
                    "confidence": 0.85,
                    "risk": "medium",
                    "impact": "Resolves configuration-related crashes",
                    "steps": [
                        "Review application logs for config errors",
                        "Identify missing or incorrect configuration",
                        "Update ConfigMap or Secret",
                        "Restart pods to pick up new config"
                    ],
                    "estimated_duration": "10 minutes"
                },
                {
                    "action": "add_missing_env_vars",
                    "description": "Add required environment variables",
                    "confidence": 0.80,
                    "risk": "low",
                    "impact": "Provides required configuration to application",
                    "steps": [
                        "Identify missing environment variables from logs",
                        "Add env vars to deployment spec",
                        "kubectl apply updated deployment",
                        "Verify pods start successfully"
                    ],
                    "estimated_duration": "5 minutes"
                }
            ],

            RootCauseType.NETWORK_ERROR: [
                {
                    "action": "check_service_connectivity",
                    "description": "Verify network connectivity to dependent services",
                    "confidence": 0.85,
                    "risk": "low",
                    "impact": "Identifies network connectivity issues",
                    "steps": [
                        "Check if target service is running",
                        "Verify Service and Endpoints exist",
                        "Test connectivity from pod: kubectl exec <pod> -- curl <service>",
                        "Check NetworkPolicy rules"
                    ],
                    "estimated_duration": "10 minutes"
                },
                {
                    "action": "fix_service_dns",
                    "description": "Resolve DNS resolution issues",
                    "confidence": 0.80,
                    "risk": "low",
                    "impact": "Fixes DNS-related connectivity problems",
                    "steps": [
                        "Check CoreDNS pods are running",
                        "Verify Service name is correct",
                        "Test DNS resolution: kubectl exec <pod> -- nslookup <service>",
                        "Check kube-dns Service"
                    ],
                    "estimated_duration": "15 minutes"
                }
            ],

            RootCauseType.VOLUME_ERROR: [
                {
                    "action": "fix_pvc_binding",
                    "description": "Resolve PersistentVolumeClaim binding issues",
                    "confidence": 0.90,
                    "risk": "medium",
                    "impact": "Enables successful volume mounting",
                    "steps": [
                        "Check PVC status: kubectl get pvc",
                        "Verify StorageClass exists",
                        "Check available PersistentVolumes",
                        "Verify node has volume plugin",
                        "Check PVC access modes match PV"
                    ],
                    "estimated_duration": "15 minutes"
                }
            ],

            RootCauseType.DISK_PRESSURE: [
                {
                    "action": "cleanup_disk_space",
                    "description": "Free up disk space on node",
                    "confidence": 0.85,
                    "risk": "medium",
                    "impact": "Resolves disk pressure condition",
                    "steps": [
                        "Identify large files or logs",
                        "Remove unused container images: docker system prune",
                        "Clean up old logs",
                        "Verify disk usage: df -h"
                    ],
                    "estimated_duration": "10 minutes"
                },
                {
                    "action": "increase_volume_size",
                    "description": "Expand PersistentVolume size",
                    "confidence": 0.80,
                    "risk": "low",
                    "impact": "Provides more storage capacity",
                    "steps": [
                        "Check if StorageClass supports expansion",
                        "Edit PVC to increase size",
                        "Wait for volume expansion",
                        "Verify new size: kubectl get pvc"
                    ],
                    "estimated_duration": "15 minutes"
                }
            ],

            RootCauseType.RESOURCE_LIMIT: [
                {
                    "action": "adjust_resource_quotas",
                    "description": "Increase namespace resource quotas",
                    "confidence": 0.90,
                    "risk": "low",
                    "impact": "Allows pods to be scheduled",
                    "steps": [
                        "Check current quota: kubectl get resourcequota",
                        "Calculate required resources",
                        "Update ResourceQuota",
                        "kubectl apply updated quota"
                    ],
                    "estimated_duration": "5 minutes"
                },
                {
                    "action": "add_node_capacity",
                    "description": "Add more nodes to cluster or increase node size",
                    "confidence": 0.85,
                    "risk": "low",
                    "impact": "Increases cluster capacity",
                    "steps": [
                        "Evaluate current cluster capacity",
                        "Add new nodes or scale node group",
                        "Wait for nodes to become ready",
                        "Verify pod scheduling"
                    ],
                    "estimated_duration": "10-30 minutes"
                }
            ]
        }