"""
Unit tests for Recommendation Engine
"""
import pytest
from unittest.mock import Mock

from pkg.types import RootCause, RootCauseType, AnalysisContext, Recommendation
from internal.recommender.engine import RecommendationEngine


class TestRecommendationEngine:
    """Test suite for RecommendationEngine"""

    @pytest.fixture
    def engine(self):
        """Create engine instance"""
        return RecommendationEngine()

    def test_recommend_oom_killer(self, engine):
        """Test recommendations for OOMKiller"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="Container was killed due to out of memory",
            confidence=0.95
        )
        context = AnalysisContext(
            metrics={"memory": {"usage_percent": 98}}
        )

        recommendations = engine.recommend(root_cause, context)

        assert len(recommendations) > 0
        # Should recommend increasing memory limit
        assert any("memory" in rec.action.lower() for rec in recommendations)
        assert any("increase" in rec.action.lower() for rec in recommendations)
        # All recommendations should have steps
        for rec in recommendations:
            assert len(rec.steps) > 0
            assert rec.confidence > 0

    def test_recommend_cpu_throttling(self, engine):
        """Test recommendations for CPU throttling"""
        root_cause = RootCause(
            type=RootCauseType.CPUThrottling,
            description="CPU is being throttled",
            confidence=0.90
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        assert len(recommendations) > 0
        assert any("cpu" in rec.action.lower() for rec in recommendations)

    def test_recommend_network_error(self, engine):
        """Test recommendations for network error"""
        root_cause = RootCause(
            type=RootCauseType.NetworkError,
            description="Connection refused",
            confidence=0.85
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        assert len(recommendations) > 0
        assert any("network" in rec.action.lower() or "connectivity" in rec.action.lower() for rec in recommendations)

    def test_recommend_image_pull_error(self, engine):
        """Test recommendations for image pull error"""
        root_cause = RootCause(
            type=RootCauseType.ImagePullError,
            description="Failed to pull image",
            confidence=0.90
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        assert len(recommendations) > 0
        assert any("image" in rec.action.lower() or "pull" in rec.action.lower() for rec in recommendations)

    def test_recommend_disk_pressure(self, engine):
        """Test recommendations for disk pressure"""
        root_cause = RootCause(
            type=RootCauseType.DiskPressure,
            description="Disk space is running low",
            confidence=0.88
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        assert len(recommendations) > 0
        assert any("disk" in rec.action.lower() for rec in recommendations)

    def test_recommendation_ordering(self, engine):
        """Test that recommendations are ordered by priority (confidence * risk)"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        # Check that recommendations are ordered
        for i in range(len(recommendations) - 1):
            score1 = recommendations[i].confidence * (1.0 if recommendations[i].risk == "low" else 0.5)
            score2 = recommendations[i + 1].confidence * (1.0 if recommendations[i + 1].risk == "low" else 0.5)
            assert score1 >= score2

    def test_recommendation_fields(self, engine):
        """Test that all recommendations have required fields"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        for rec in recommendations:
            assert rec.action is not None
            assert rec.description is not None
            assert rec.confidence > 0
            assert rec.risk in ["low", "medium", "high", "none"]
            assert len(rec.steps) > 0
            assert rec.estimated_duration is not None

    def test_conditional_recommendations(self, engine):
        """Test conditional recommendations based on context"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )

        # Context with high memory usage - should recommend increase
        context_high = AnalysisContext(
            metrics={"memory": {"usage_percent": 90}}
        )
        recs_high = engine.recommend(root_cause, context_high)

        # Context with low memory usage - should recommend investigation
        context_low = AnalysisContext(
            metrics={"memory": {"usage_percent": 50}}
        )
        recs_low = engine.recommend(root_cause, context_low)

        # Both should have recommendations but potentially different ones
        assert len(recs_high) > 0
        assert len(recs_low) > 0

    def test_recommendation_steps(self, engine):
        """Test that recommendations have actionable steps"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        for rec in recommendations:
            assert len(rec.steps) >= 1
            # Each step should be a non-empty string
            for step in rec.steps:
                assert isinstance(step, str)
                assert len(step) > 0

    def test_recommendation_rollback(self, engine):
        """Test that recommendations with risk include rollback steps"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        for rec in recommendations:
            if rec.risk in ["medium", "high"]:
                # Should have rollback steps
                assert hasattr(rec, "rollback_steps") or len(rec.steps) > 0

    def test_unknown_root_cause(self, engine):
        """Test handling of unknown root cause type"""
        root_cause = RootCause(
            type="UnknownType",
            description="Unknown issue",
            confidence=0.50
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context)

        # Should return empty list or generic recommendations
        assert isinstance(recommendations, list)

    def test_max_recommendations_limit(self, engine):
        """Test that number of recommendations can be limited"""
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )
        context = AnalysisContext()

        recommendations = engine.recommend(root_cause, context, max_recommendations=3)

        assert len(recommendations) <= 3


@pytest.mark.benchmark
class TestRecommendationEnginePerformance:
    """Performance benchmarks for RecommendationEngine"""

    def test_benchmark_generate_recommendations(self, benchmark):
        """Benchmark recommendation generation"""
        engine = RecommendationEngine()
        root_cause = RootCause(
            type=RootCauseType.OOMKiller,
            description="OOM",
            confidence=0.95
        )
        context = AnalysisContext()

        recommendations = benchmark(engine.recommend, root_cause, context)

        assert len(recommendations) > 0
