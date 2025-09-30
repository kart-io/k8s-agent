"""
Root Cause Analysis Engine
Uses pattern matching, NLP, and ML to determine failure root causes
"""

import re
import json
from typing import Dict, List, Optional, Tuple
from loguru import logger

from pkg.types import (
    AnalysisContext, RootCause, RootCauseType, AnalysisResult
)


class RootCauseAnalyzer:
    """Root cause analyzer using multiple analysis techniques"""

    def __init__(self):
        """Initialize the analyzer"""
        self.patterns = self._load_patterns()
        self.keyword_weights = self._load_keyword_weights()

    def analyze(self, context: AnalysisContext) -> AnalysisResult:
        """
        Analyze context and determine root cause

        Args:
            context: Analysis context with event, logs, metrics, etc.

        Returns:
            Analysis result with root cause and evidence
        """
        logger.info("Starting root cause analysis")

        # Multiple analysis approaches
        analyses = []

        # 1. Event-based analysis
        if context.event:
            event_analysis = self._analyze_event(context.event)
            if event_analysis:
                analyses.append(event_analysis)

        # 2. Log-based analysis
        if context.logs:
            log_analysis = self._analyze_logs(context.logs)
            if log_analysis:
                analyses.append(log_analysis)

        # 3. Metrics-based analysis
        if context.metrics:
            metrics_analysis = self._analyze_metrics(context.metrics)
            if metrics_analysis:
                analyses.append(metrics_analysis)

        # 4. Correlation analysis
        if len(analyses) > 1:
            correlation_analysis = self._correlate_analyses(analyses)
            if correlation_analysis:
                analyses.append(correlation_analysis)

        # Select best analysis
        if not analyses:
            logger.warning("No root cause identified")
            return AnalysisResult(
                confidence=0.0,
                evidence=["Insufficient data for analysis"]
            )

        # Sort by confidence and select best
        best_analysis = max(analyses, key=lambda x: x[1])
        root_cause, confidence, evidence = best_analysis

        logger.info(f"Root cause identified: {root_cause.type} (confidence: {confidence:.2f})")

        return AnalysisResult(
            root_cause=root_cause,
            confidence=confidence,
            evidence=evidence
        )

    def _analyze_event(self, event: Dict) -> Optional[Tuple[RootCause, float, List[str]]]:
        """Analyze Kubernetes event"""
        logger.debug("Analyzing event data")

        reason = event.get("reason", "")
        message = event.get("message", "")

        # Direct mapping from event reason
        reason_map = {
            "OOMKilling": RootCauseType.OOM_KILLER,
            "OOMKilled": RootCauseType.OOM_KILLER,
            "FailedScheduling": RootCauseType.RESOURCE_LIMIT,
            "ImagePullBackOff": RootCauseType.IMAGE_PULL_ERROR,
            "ErrImagePull": RootCauseType.IMAGE_PULL_ERROR,
            "CrashLoopBackOff": RootCauseType.CONFIG_ERROR,  # Default, need more analysis
            "FailedMount": RootCauseType.VOLUME_ERROR,
            "FailedAttachVolume": RootCauseType.VOLUME_ERROR,
        }

        if reason in reason_map:
            root_cause_type = reason_map[reason]
            evidence = [f"Event reason: {reason}"]

            # Add message as evidence
            if message:
                evidence.append(f"Event message: {message}")

            # OOM has high confidence
            confidence = 0.95 if root_cause_type == RootCauseType.OOM_KILLER else 0.85

            root_cause = RootCause(
                type=root_cause_type,
                description=self._get_root_cause_description(root_cause_type, message),
                confidence=confidence,
                evidence=evidence
            )

            return (root_cause, confidence, evidence)

        return None

    def _analyze_logs(self, logs: str) -> Optional[Tuple[RootCause, float, List[str]]]:
        """Analyze application logs"""
        logger.debug(f"Analyzing logs ({len(logs)} chars)")

        evidence = []
        scores = {root_cause_type: 0.0 for root_cause_type in RootCauseType}

        # Pattern matching
        for pattern_info in self.patterns:
            pattern = pattern_info["pattern"]
            root_cause_type = RootCauseType(pattern_info["type"])
            weight = pattern_info.get("weight", 1.0)

            matches = re.findall(pattern, logs, re.IGNORECASE | re.MULTILINE)
            if matches:
                scores[root_cause_type] += len(matches) * weight
                evidence.append(f"Found pattern: {pattern_info['description']} ({len(matches)} occurrences)")

        # Keyword matching
        log_lower = logs.lower()
        for keyword, info in self.keyword_weights.items():
            if keyword in log_lower:
                root_cause_type = RootCauseType(info["type"])
                weight = info["weight"]
                count = log_lower.count(keyword)
                scores[root_cause_type] += count * weight
                evidence.append(f"Found keyword: '{keyword}' ({count} times)")

        # Find highest score
        max_score = max(scores.values())
        if max_score == 0:
            return None

        # Normalize to confidence (0-1)
        confidence = min(max_score / 10.0, 0.95)  # Cap at 0.95

        best_type = max(scores, key=scores.get)

        root_cause = RootCause(
            type=best_type,
            description=self._get_root_cause_description(best_type, logs[:500]),
            confidence=confidence,
            evidence=evidence[:5]  # Top 5 evidence
        )

        return (root_cause, confidence, evidence[:5])

    def _analyze_metrics(self, metrics: Dict) -> Optional[Tuple[RootCause, float, List[str]]]:
        """Analyze resource metrics"""
        logger.debug("Analyzing metrics data")

        evidence = []

        # Check for OOM conditions
        if "memory" in metrics:
            mem = metrics["memory"]
            if isinstance(mem, dict):
                usage = mem.get("usage_percent", 0)
                if usage >= 95:
                    evidence.append(f"Memory usage at {usage}%")
                    root_cause = RootCause(
                        type=RootCauseType.OOM_KILLER,
                        description="Memory usage exceeded limits",
                        confidence=0.9,
                        evidence=evidence
                    )
                    return (root_cause, 0.9, evidence)

        # Check for CPU throttling
        if "cpu" in metrics:
            cpu = metrics["cpu"]
            if isinstance(cpu, dict):
                throttling = cpu.get("throttling_percent", 0)
                if throttling >= 50:
                    evidence.append(f"CPU throttling at {throttling}%")
                    root_cause = RootCause(
                        type=RootCauseType.CPU_THROTTLING,
                        description="CPU throttling detected",
                        confidence=0.85,
                        evidence=evidence
                    )
                    return (root_cause, 0.85, evidence)

        # Check for disk pressure
        if "disk" in metrics:
            disk = metrics["disk"]
            if isinstance(disk, dict):
                usage = disk.get("usage_percent", 0)
                if usage >= 90:
                    evidence.append(f"Disk usage at {usage}%")
                    root_cause = RootCause(
                        type=RootCauseType.DISK_PRESSURE,
                        description="Disk space exhausted",
                        confidence=0.85,
                        evidence=evidence
                    )
                    return (root_cause, 0.85, evidence)

        return None

    def _correlate_analyses(self, analyses: List[Tuple]) -> Optional[Tuple[RootCause, float, List[str]]]:
        """Correlate multiple analyses to improve confidence"""
        logger.debug(f"Correlating {len(analyses)} analyses")

        # Count occurrences of each root cause type
        type_counts = {}
        all_evidence = []

        for root_cause, confidence, evidence in analyses:
            rc_type = root_cause.type
            if rc_type not in type_counts:
                type_counts[rc_type] = {"count": 0, "total_confidence": 0.0, "evidence": []}
            type_counts[rc_type]["count"] += 1
            type_counts[rc_type]["total_confidence"] += confidence
            type_counts[rc_type]["evidence"].extend(evidence)

        # Find most common type
        best_type = max(type_counts, key=lambda x: type_counts[x]["count"])
        info = type_counts[best_type]

        # If multiple analyses agree, boost confidence
        if info["count"] >= 2:
            avg_confidence = info["total_confidence"] / info["count"]
            boosted_confidence = min(avg_confidence * 1.1, 0.98)

            root_cause = RootCause(
                type=best_type,
                description=f"{best_type.value} (confirmed by {info['count']} analyses)",
                confidence=boosted_confidence,
                evidence=list(set(info["evidence"][:5]))
            )

            return (root_cause, boosted_confidence, root_cause.evidence)

        return None

    def _get_root_cause_description(self, root_cause_type: RootCauseType, context: str) -> str:
        """Get human-readable description"""
        descriptions = {
            RootCauseType.OOM_KILLER: "Container was killed due to out of memory (OOM)",
            RootCauseType.CPU_THROTTLING: "CPU throttling due to resource limits",
            RootCauseType.DISK_PRESSURE: "Disk space exhausted or I/O bottleneck",
            RootCauseType.NETWORK_ERROR: "Network connectivity or DNS resolution error",
            RootCauseType.CONFIG_ERROR: "Configuration error or missing environment variable",
            RootCauseType.IMAGE_PULL_ERROR: "Failed to pull container image",
            RootCauseType.VOLUME_ERROR: "Volume mount or attachment failure",
            RootCauseType.DEPENDENCY_ERROR: "External service dependency failure",
            RootCauseType.RESOURCE_LIMIT: "Resource quota exceeded or scheduling constraint",
            RootCauseType.UNKNOWN: "Unable to determine specific root cause",
        }
        return descriptions.get(root_cause_type, "Unknown root cause")

    def _load_patterns(self) -> List[Dict]:
        """Load regex patterns for log analysis"""
        return [
            {
                "pattern": r"out of memory|oom|memory.*exhausted",
                "type": "OOMKiller",
                "description": "OOM indicator",
                "weight": 2.0
            },
            {
                "pattern": r"killed.*signal 9|sigkill",
                "type": "OOMKiller",
                "description": "SIGKILL",
                "weight": 1.5
            },
            {
                "pattern": r"exit code 137",
                "type": "OOMKiller",
                "description": "Exit code 137 (OOMKilled)",
                "weight": 2.0
            },
            {
                "pattern": r"connection refused|connection timeout",
                "type": "NetworkError",
                "description": "Connection error",
                "weight": 1.5
            },
            {
                "pattern": r"cannot pull image|pull.*failed|image pull back",
                "type": "ImagePullError",
                "description": "Image pull failure",
                "weight": 2.0
            },
            {
                "pattern": r"config.*not found|missing.*environment|env.*required",
                "type": "ConfigError",
                "description": "Configuration issue",
                "weight": 1.5
            },
            {
                "pattern": r"permission denied|forbidden|unauthorized",
                "type": "ConfigError",
                "description": "Permission issue",
                "weight": 1.5
            },
            {
                "pattern": r"no space left|disk.*full",
                "type": "DiskPressure",
                "description": "Disk space issue",
                "weight": 2.0
            },
            {
                "pattern": r"panic|fatal error|segmentation fault",
                "type": "ConfigError",
                "description": "Application crash",
                "weight": 1.5
            },
        ]

    def _load_keyword_weights(self) -> Dict:
        """Load keyword weights for analysis"""
        return {
            "oom": {"type": "OOMKiller", "weight": 2.0},
            "killed": {"type": "OOMKiller", "weight": 1.0},
            "timeout": {"type": "NetworkError", "weight": 1.5},
            "refused": {"type": "NetworkError", "weight": 1.5},
            "image": {"type": "ImagePullError", "weight": 1.0},
            "volume": {"type": "VolumeError", "weight": 1.5},
            "disk": {"type": "DiskPressure", "weight": 1.0},
            "cpu": {"type": "CPUThrottling", "weight": 1.0},
            "config": {"type": "ConfigError", "weight": 1.0},
            "panic": {"type": "ConfigError", "weight": 1.5},
        }