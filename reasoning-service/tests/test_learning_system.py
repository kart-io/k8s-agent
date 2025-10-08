"""
Unit tests for Learning System
"""
import pytest
from datetime import datetime, timedelta
from unittest.mock import Mock, MagicMock

from pkg.types import (
    Feedback, FeedbackType, RootCauseType
)
from internal.learning.system import LearningSystem


class TestLearningSystem:
    """Test suite for LearningSystem"""

    @pytest.fixture
    def learning_system(self):
        """Create learning system instance"""
        return LearningSystem()

    @pytest.fixture
    def sample_feedback(self):
        """Create sample feedback"""
        return Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=5,
            was_helpful=True,
            actual_root_cause=RootCauseType.OOMKiller.value,
            actual_solution="Increased memory limits",
            comments="Diagnosis was spot on",
            timestamp=datetime.now()
        )

    def test_init(self, learning_system):
        """Test learning system initialization"""
        assert learning_system is not None
        assert isinstance(learning_system.feedback_store, dict)
        assert isinstance(learning_system.accuracy_metrics, dict)
        # Should have metrics for all root cause types
        assert len(learning_system.accuracy_metrics) > 0

    def test_process_feedback_success(self, learning_system, sample_feedback):
        """Test processing feedback successfully"""
        result = learning_system.process_feedback(sample_feedback)

        assert result is True
        assert sample_feedback.request_id in learning_system.feedback_store
        assert len(learning_system.feedback_store[sample_feedback.request_id]) == 1

    def test_process_multiple_feedbacks(self, learning_system):
        """Test processing multiple feedbacks"""
        feedbacks = [
            Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=4 + (i % 2),
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            for i in range(5)
        ]

        for fb in feedbacks:
            result = learning_system.process_feedback(fb)
            assert result is True

        assert len(learning_system.feedback_store) == 5

    def test_update_diagnosis_metrics(self, learning_system):
        """Test updating diagnosis metrics"""
        feedback = Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=5,
            was_helpful=True,
            actual_root_cause=RootCauseType.OOMKiller.value,
            timestamp=datetime.now()
        )

        learning_system.process_feedback(feedback)

        metrics = learning_system.accuracy_metrics[RootCauseType.OOMKiller.value]
        assert metrics["total_diagnoses"] == 1
        assert metrics["correct_diagnoses"] == 1
        assert metrics["accuracy"] == 1.0

    def test_update_diagnosis_metrics_incorrect(self, learning_system):
        """Test updating metrics for incorrect diagnosis"""
        feedback = Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=2,
            was_helpful=False,
            actual_root_cause=RootCauseType.OOMKiller.value,
            timestamp=datetime.now()
        )

        learning_system.process_feedback(feedback)

        metrics = learning_system.accuracy_metrics[RootCauseType.OOMKiller.value]
        assert metrics["total_diagnoses"] == 1
        assert metrics["correct_diagnoses"] == 0
        assert metrics["accuracy"] == 0.0

    def test_update_diagnosis_metrics_mixed(self, learning_system):
        """Test metrics with mixed feedback"""
        feedbacks = [
            Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=5 if i < 3 else 2,
                was_helpful=i < 3,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            for i in range(5)
        ]

        for fb in feedbacks:
            learning_system.process_feedback(fb)

        metrics = learning_system.accuracy_metrics[RootCauseType.OOMKiller.value]
        assert metrics["total_diagnoses"] == 5
        assert metrics["correct_diagnoses"] == 3
        assert metrics["accuracy"] == 0.6

    def test_update_recommendation_metrics(self, learning_system):
        """Test updating recommendation metrics"""
        feedback = Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.RECOMMENDATION_USEFULNESS,
            rating=4,
            was_helpful=True,
            timestamp=datetime.now()
        )

        result = learning_system.process_feedback(feedback)

        assert result is True
        assert feedback.request_id in learning_system.feedback_store

    def test_update_prediction_metrics(self, learning_system):
        """Test updating prediction metrics"""
        feedback = Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.PREDICTION_ACCURACY,
            rating=5,
            was_helpful=True,
            timestamp=datetime.now()
        )

        result = learning_system.process_feedback(feedback)

        assert result is True

    def test_update_knowledge_graph(self, learning_system):
        """Test updating knowledge graph with feedback"""
        mock_kg = Mock()
        mock_kg.add_feedback = Mock(return_value=True)
        learning_system.knowledge_graph = mock_kg

        feedback = Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=5,
            was_helpful=True,
            actual_root_cause=RootCauseType.OOMKiller.value,
            actual_solution="Increased memory",
            timestamp=datetime.now()
        )

        learning_system.process_feedback(feedback)

        mock_kg.add_feedback.assert_called_once()

    def test_get_accuracy_metrics_specific(self, learning_system):
        """Test getting metrics for specific root cause"""
        feedback = Feedback(
            feedback_id="fb-001",
            request_id="req-001",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=5,
            was_helpful=True,
            actual_root_cause=RootCauseType.OOMKiller.value,
            timestamp=datetime.now()
        )

        learning_system.process_feedback(feedback)

        metrics = learning_system.get_accuracy_metrics(RootCauseType.OOMKiller.value)

        assert metrics["total_diagnoses"] == 1
        assert metrics["accuracy"] == 1.0

    def test_get_accuracy_metrics_overall(self, learning_system):
        """Test getting overall accuracy metrics"""
        feedbacks = [
            Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=5,
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value if i < 3 else RootCauseType.NetworkError.value,
                timestamp=datetime.now()
            )
            for i in range(5)
        ]

        for fb in feedbacks:
            learning_system.process_feedback(fb)

        metrics = learning_system.get_accuracy_metrics()

        assert "overall" in metrics
        assert "by_root_cause" in metrics
        assert metrics["overall"] == 1.0  # All correct

    def test_calculate_overall_accuracy_empty(self, learning_system):
        """Test calculating overall accuracy with no data"""
        learning_system._init_metrics()  # Reset metrics

        accuracy = learning_system._calculate_overall_accuracy()

        assert accuracy == 0.0

    def test_suggest_improvements_low_accuracy(self, learning_system):
        """Test suggesting improvements for low accuracy"""
        # Add many incorrect diagnoses
        for i in range(10):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=2 if i < 8 else 5,
                was_helpful=i >= 8,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        suggestions = learning_system.suggest_improvements()

        assert len(suggestions) > 0
        assert any(s["type"] == "low_accuracy" for s in suggestions)
        low_acc_suggestion = next(s for s in suggestions if s["type"] == "low_accuracy")
        assert low_acc_suggestion["current_accuracy"] < 0.7

    def test_suggest_improvements_high_accuracy(self, learning_system):
        """Test no suggestions for high accuracy"""
        # Add all correct diagnoses
        for i in range(10):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=5,
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        suggestions = learning_system.suggest_improvements()

        # Should have no low accuracy suggestions
        low_acc_suggestions = [s for s in suggestions if s["type"] == "low_accuracy"]
        assert len(low_acc_suggestions) == 0

    def test_find_common_misdiagnoses(self, learning_system):
        """Test finding common misdiagnosis patterns"""
        # Add multiple unhelpful feedbacks for same root cause
        for i in range(5):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=2,
                was_helpful=False,
                actual_root_cause=RootCauseType.NetworkError.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        patterns = learning_system._find_common_misdiagnoses()

        assert RootCauseType.NetworkError.value in patterns

    def test_export_learning_data(self, learning_system, sample_feedback):
        """Test exporting learning data"""
        learning_system.process_feedback(sample_feedback)

        data = learning_system.export_learning_data()

        assert "accuracy_metrics" in data
        assert "total_feedback" in data
        assert "feedback_by_type" in data
        assert "export_time" in data
        assert data["total_feedback"] == 1

    def test_import_learning_data(self, learning_system):
        """Test importing learning data"""
        data = {
            "accuracy_metrics": {
                RootCauseType.OOMKiller.value: {
                    "total_diagnoses": 100,
                    "correct_diagnoses": 90,
                    "accuracy": 0.9,
                    "last_updated": datetime.now()
                }
            }
        }

        result = learning_system.import_learning_data(data)

        assert result is True
        metrics = learning_system.accuracy_metrics[RootCauseType.OOMKiller.value]
        assert metrics["total_diagnoses"] == 100
        assert metrics["accuracy"] == 0.9

    def test_import_invalid_data(self, learning_system):
        """Test importing invalid learning data"""
        data = {"invalid": "data"}

        result = learning_system.import_learning_data(data)

        # Should still succeed but not update anything
        assert result is True

    def test_analyze_trends_no_data(self, learning_system):
        """Test analyzing trends with no data"""
        trends = learning_system.analyze_trends(time_window="7d")

        assert "message" in trends
        assert trends["message"] == "No recent feedback data"

    def test_analyze_trends_with_data(self, learning_system):
        """Test analyzing trends with feedback data"""
        # Add recent feedback
        for i in range(10):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=4 + (i % 2),
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        trends = learning_system.analyze_trends(time_window="7d")

        assert "time_window" in trends
        assert "total_feedback" in trends
        assert "helpful_rate" in trends
        assert "average_rating" in trends
        assert "trend" in trends
        assert trends["total_feedback"] == 10
        assert trends["helpful_rate"] == 1.0

    def test_analyze_trends_old_data(self, learning_system):
        """Test analyzing trends with old data outside window"""
        # Add old feedback
        old_time = datetime.now() - timedelta(days=30)
        feedback = Feedback(
            feedback_id="fb-old",
            request_id="req-old",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=5,
            was_helpful=True,
            actual_root_cause=RootCauseType.OOMKiller.value,
            timestamp=old_time
        )
        learning_system.process_feedback(feedback)

        trends = learning_system.analyze_trends(time_window="7d")

        # Old feedback should not be included
        assert "message" in trends or trends.get("total_feedback", 0) == 0

    def test_get_top_performing_patterns_empty(self, learning_system):
        """Test getting top patterns with no data"""
        learning_system._init_metrics()  # Reset

        patterns = learning_system.get_top_performing_patterns(limit=5)

        assert isinstance(patterns, list)
        assert len(patterns) == 0

    def test_get_top_performing_patterns(self, learning_system):
        """Test getting top performing patterns"""
        # Add feedback for different root causes with different accuracies
        root_causes = [
            (RootCauseType.OOMKiller, 5, 5),  # 100% accuracy
            (RootCauseType.NetworkError, 5, 3),  # 60% accuracy
            (RootCauseType.CPUThrottling, 5, 4),  # 80% accuracy
        ]

        for root_cause, total, correct in root_causes:
            for i in range(total):
                feedback = Feedback(
                    feedback_id=f"fb-{root_cause.value}-{i:03d}",
                    request_id=f"req-{root_cause.value}-{i:03d}",
                    feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                    rating=5 if i < correct else 2,
                    was_helpful=i < correct,
                    actual_root_cause=root_cause.value,
                    timestamp=datetime.now()
                )
                learning_system.process_feedback(feedback)

        patterns = learning_system.get_top_performing_patterns(limit=3)

        assert len(patterns) == 3
        # OOMKiller should be first (100% accuracy)
        assert patterns[0]["root_cause"] == RootCauseType.OOMKiller.value
        assert patterns[0]["accuracy"] == 1.0

    def test_get_top_performing_patterns_min_sample_size(self, learning_system):
        """Test that patterns require minimum sample size"""
        # Add only 2 feedbacks (below minimum of 3)
        for i in range(2):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=5,
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        patterns = learning_system.get_top_performing_patterns()

        # Should not include patterns with < 3 samples
        assert len(patterns) == 0

    def test_reset_metrics_all(self, learning_system, sample_feedback):
        """Test resetting all metrics"""
        learning_system.process_feedback(sample_feedback)

        learning_system.reset_metrics()

        assert len(learning_system.feedback_store) == 0
        metrics = learning_system.accuracy_metrics[RootCauseType.OOMKiller.value]
        assert metrics["total_diagnoses"] == 0
        assert metrics["accuracy"] == 0.0

    def test_reset_metrics_specific(self, learning_system):
        """Test resetting specific root cause metrics"""
        # Add feedback for two different root causes
        for root_cause in [RootCauseType.OOMKiller, RootCauseType.NetworkError]:
            feedback = Feedback(
                feedback_id=f"fb-{root_cause.value}",
                request_id=f"req-{root_cause.value}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=5,
                was_helpful=True,
                actual_root_cause=root_cause.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        # Reset only OOMKiller
        learning_system.reset_metrics(RootCauseType.OOMKiller.value)

        oom_metrics = learning_system.accuracy_metrics[RootCauseType.OOMKiller.value]
        network_metrics = learning_system.accuracy_metrics[RootCauseType.NetworkError.value]

        assert oom_metrics["total_diagnoses"] == 0
        assert network_metrics["total_diagnoses"] == 1  # Not reset

    def test_feedback_storage_multiple_per_request(self, learning_system):
        """Test storing multiple feedbacks for same request"""
        request_id = "req-001"

        feedbacks = [
            Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=request_id,
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=4 + i,
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            for i in range(3)
        ]

        for fb in feedbacks:
            learning_system.process_feedback(fb)

        assert len(learning_system.feedback_store[request_id]) == 3


@pytest.mark.benchmark
class TestLearningSystemPerformance:
    """Performance benchmarks for LearningSystem"""

    def test_benchmark_process_feedback(self, benchmark):
        """Benchmark processing feedback"""
        learning_system = LearningSystem()

        feedback = Feedback(
            feedback_id="fb-bench",
            request_id="req-bench",
            feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
            rating=5,
            was_helpful=True,
            actual_root_cause=RootCauseType.OOMKiller.value,
            timestamp=datetime.now()
        )

        result = benchmark(learning_system.process_feedback, feedback)

        assert result is True

    def test_benchmark_get_accuracy_metrics(self, benchmark):
        """Benchmark getting accuracy metrics"""
        learning_system = LearningSystem()

        # Add some feedback first
        for i in range(50):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=4 + (i % 2),
                was_helpful=True,
                actual_root_cause=RootCauseType.OOMKiller.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        metrics = benchmark(learning_system.get_accuracy_metrics)

        assert "overall" in metrics

    def test_benchmark_suggest_improvements(self, benchmark):
        """Benchmark suggesting improvements"""
        learning_system = LearningSystem()

        # Add feedback with varying accuracy
        for i in range(100):
            feedback = Feedback(
                feedback_id=f"fb-{i:03d}",
                request_id=f"req-{i:03d}",
                feedback_type=FeedbackType.DIAGNOSIS_ACCURACY,
                rating=3 + (i % 3),
                was_helpful=i % 2 == 0,
                actual_root_cause=RootCauseType.OOMKiller.value if i < 50 else RootCauseType.NetworkError.value,
                timestamp=datetime.now()
            )
            learning_system.process_feedback(feedback)

        suggestions = benchmark(learning_system.suggest_improvements)

        assert isinstance(suggestions, list)
