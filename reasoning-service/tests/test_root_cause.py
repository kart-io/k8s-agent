"""
Unit tests for Root Cause Analyzer
"""
import pytest
from unittest.mock import Mock, patch

from pkg.types import AnalysisContext, AnalysisResult, RootCauseType
from internal.analyzer.root_cause import RootCauseAnalyzer


class TestRootCauseAnalyzer:
    """Test suite for RootCauseAnalyzer"""

    @pytest.fixture
    def analyzer(self):
        """Create analyzer instance"""
        return RootCauseAnalyzer()

    def test_analyze_oom_killer_event(self, analyzer):
        """Test OOMKiller detection from event"""
        context = AnalysisContext(
            event={"reason": "OOMKilled", "message": "Container was killed due to OOM"}
        )

        result = analyzer.analyze(context)

        assert result.root_cause is not None
        assert result.root_cause.type == RootCauseType.OOMKiller
        assert result.confidence >= 0.9
        assert len(result.evidence) > 0

    def test_analyze_oom_killer_logs(self, analyzer):
        """Test OOMKiller detection from logs"""
        context = AnalysisContext(
            logs="fatal error: runtime: out of memory\nOOM killer terminated process"
        )

        result = analyzer.analyze(context)

        assert result.root_cause is not None
        assert result.root_cause.type == RootCauseType.OOMKiller
        assert result.confidence >= 0.8

    def test_analyze_network_error_logs(self, analyzer):
        """Test network error detection"""
        context = AnalysisContext(
            logs="dial tcp 10.0.0.1:3306: connection refused\nno route to host"
        )

        result = analyzer.analyze(context)

        assert result.root_cause is not None
        assert result.root_cause.type == RootCauseType.NetworkError
        assert "connection refused" in str(result.evidence).lower()

    def test_analyze_cpu_throttling_metrics(self, analyzer):
        """Test CPU throttling detection from metrics"""
        context = AnalysisContext(
            metrics={
                "cpu": {
                    "throttling_percent": 85
                }
            }
        )

        result = analyzer.analyze(context)

        assert result.root_cause is not None
        assert result.root_cause.type == RootCauseType.CPUThrottling
        assert result.confidence >= 0.85

    def test_analyze_image_pull_error(self, analyzer):
        """Test image pull error detection"""
        context = AnalysisContext(
            event={"reason": "ImagePullBackOff", "message": "Failed to pull image"}
        )

        result = analyzer.analyze(context)

        assert result.root_cause is not None
        assert result.root_cause.type == RootCauseType.ImagePullError

    def test_analyze_multimodal(self, analyzer):
        """Test multimodal analysis (event + logs + metrics)"""
        context = AnalysisContext(
            event={"reason": "OOMKilled"},
            logs="out of memory",
            metrics={"memory": {"usage_percent": 98}}
        )

        result = analyzer.analyze(context)

        assert result.root_cause is not None
        assert result.root_cause.type == RootCauseType.OOMKiller
        assert result.confidence >= 0.95
        assert len(result.evidence) >= 3  # Evidence from all sources

    def test_analyze_no_root_cause(self, analyzer):
        """Test when no root cause is found"""
        context = AnalysisContext(
            event={"reason": "Unknown", "message": "Something happened"}
        )

        result = analyzer.analyze(context)

        assert result.root_cause is None
        assert result.confidence == 0.0

    def test_analyze_empty_context(self, analyzer):
        """Test with empty context"""
        context = AnalysisContext()

        result = analyzer.analyze(context)

        assert result.root_cause is None

    def test_event_reason_mapping(self, analyzer):
        """Test event reason direct mapping"""
        test_cases = [
            ("OOMKilled", RootCauseType.OOMKiller),
            ("ImagePullBackOff", RootCauseType.ImagePullError),
            ("FailedMount", RootCauseType.VolumeError),
        ]

        for reason, expected_type in test_cases:
            context = AnalysisContext(event={"reason": reason})
            result = analyzer.analyze(context)
            assert result.root_cause.type == expected_type

    def test_log_pattern_matching(self, analyzer):
        """Test log pattern matching"""
        test_cases = [
            ("java.lang.OutOfMemoryError: Java heap space", RootCauseType.OOMKiller),
            ("connection timed out", RootCauseType.NetworkError),
            ("disk quota exceeded", RootCauseType.DiskPressure),
        ]

        for log_message, expected_type in test_cases:
            context = AnalysisContext(logs=log_message)
            result = analyzer.analyze(context)
            assert result.root_cause.type == expected_type

    def test_metric_threshold_detection(self, analyzer):
        """Test metric threshold detection"""
        # High memory usage
        context = AnalysisContext(
            metrics={"memory": {"usage_percent": 96}}
        )
        result = analyzer.analyze(context)
        assert result.root_cause.type == RootCauseType.OOMKiller

        # High disk usage
        context = AnalysisContext(
            metrics={"disk": {"usage_percent": 98}}
        )
        result = analyzer.analyze(context)
        assert result.root_cause.type == RootCauseType.DiskPressure

    def test_confidence_scoring(self, analyzer):
        """Test confidence scoring"""
        # Single evidence source - lower confidence
        context1 = AnalysisContext(event={"reason": "CrashLoopBackOff"})
        result1 = analyzer.analyze(context1)

        # Multiple evidence sources - higher confidence
        context2 = AnalysisContext(
            event={"reason": "OOMKilled"},
            logs="out of memory",
            metrics={"memory": {"usage_percent": 98}}
        )
        result2 = analyzer.analyze(context2)

        if result1.root_cause and result2.root_cause:
            assert result2.confidence > result1.confidence

    def test_crashloop_analysis(self, analyzer):
        """Test CrashLoopBackOff requires further analysis"""
        context = AnalysisContext(
            event={"reason": "CrashLoopBackOff"},
            logs="panic: runtime error"
        )

        result = analyzer.analyze(context)

        # Should analyze logs to determine actual cause
        assert result.root_cause is not None


@pytest.mark.benchmark
class TestRootCauseAnalyzerPerformance:
    """Performance benchmarks for RootCauseAnalyzer"""

    def test_benchmark_simple_analysis(self, benchmark):
        """Benchmark simple event analysis"""
        analyzer = RootCauseAnalyzer()
        context = AnalysisContext(event={"reason": "OOMKilled"})

        result = benchmark(analyzer.analyze, context)

        assert result.root_cause is not None

    def test_benchmark_multimodal_analysis(self, benchmark):
        """Benchmark multimodal analysis"""
        analyzer = RootCauseAnalyzer()
        context = AnalysisContext(
            event={"reason": "OOMKilled"},
            logs="out of memory\n" * 100,  # Large log
            metrics={"memory": {"usage_percent": 98}}
        )

        result = benchmark(analyzer.analyze, context)

        assert result.root_cause is not None
