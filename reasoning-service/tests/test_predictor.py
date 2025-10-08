"""
Unit tests for Prediction Engine
"""
import pytest
import numpy as np
from unittest.mock import Mock, patch

from pkg.types import PredictionRequest, PredictionResult, MetricSnapshot
from internal.predictor.engine import PredictionEngine


class TestPredictionEngine:
    """Test suite for PredictionEngine"""

    @pytest.fixture
    def engine(self):
        """Create engine instance"""
        return PredictionEngine()

    def test_predict_high_memory_usage(self, engine):
        """Test prediction for high memory usage"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 95},
                "cpu": {"usage_percent": 70}
            },
            time_window="24h"
        )

        result = engine.predict_failure(request)

        assert isinstance(result, PredictionResult)
        assert result.failure_probability >= 0.7  # High memory should predict failure
        assert "OOMKiller" in result.failure_types or "memory" in str(result.contributing_factors).lower()

    def test_predict_cpu_throttling(self, engine):
        """Test prediction for CPU throttling"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "cpu": {
                    "usage_percent": 80,
                    "throttling_percent": 75
                }
            }
        )

        result = engine.predict_failure(request)

        assert result.failure_probability > 0
        assert "CPUThrottling" in result.failure_types or "cpu" in str(result.contributing_factors).lower()

    def test_predict_low_risk(self, engine):
        """Test prediction for healthy metrics"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 40},
                "cpu": {"usage_percent": 30}
            }
        )

        result = engine.predict_failure(request)

        assert result.failure_probability < 0.3  # Low risk

    def test_predict_with_history(self, engine):
        """Test prediction with historical data"""
        # Create upward trend
        history = []
        for i in range(10):
            history.append(MetricSnapshot(
                timestamp=f"2025-10-01T{10+i:02d}:00:00Z",
                memory={"usage_percent": 50 + i * 5},
                cpu={"usage_percent": 40}
            ))

        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 95},
                "history": history
            }
        )

        result = engine.predict_failure(request)

        assert result.failure_probability > 0.5  # Upward trend should increase probability
        assert result.predicted_failure_time is not None

    def test_predict_restart_count(self, engine):
        """Test prediction considering restart count"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "restart_count": 5,
                "memory": {"usage_percent": 80}
            }
        )

        result = engine.predict_failure(request)

        assert result.failure_probability > 0.5
        assert any("restart" in factor.lower() for factor in result.contributing_factors)

    def test_prediction_confidence(self, engine):
        """Test prediction confidence scoring"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 95}
            }
        )

        result = engine.predict_failure(request)

        assert 0 <= result.confidence <= 1.0
        assert result.confidence > 0  # Should have some confidence

    def test_prediction_time_estimation(self, engine):
        """Test prediction time estimation"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 95}
            }
        )

        result = engine.predict_failure(request)

        if result.failure_probability > 0.7:
            assert result.predicted_failure_time is not None

    def test_contributing_factors(self, engine):
        """Test that contributing factors are identified"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 95},
                "cpu": {"throttling_percent": 80},
                "restart_count": 3
            }
        )

        result = engine.predict_failure(request)

        assert len(result.contributing_factors) > 0
        # Should mention high resource usage
        factors_text = " ".join(result.contributing_factors).lower()
        assert "memory" in factors_text or "cpu" in factors_text

    def test_isolation_forest_training(self, engine):
        """Test Isolation Forest training"""
        # Generate training data
        training_data = []
        for i in range(30):
            training_data.append({
                "memory_usage": 50 + np.random.randint(-10, 10),
                "cpu_usage": 40 + np.random.randint(-10, 10),
                "restart_count": 0
            })

        success = engine.train_anomaly_detector(training_data)

        assert success is True
        assert engine.is_trained is True

    def test_isolation_forest_insufficient_data(self, engine):
        """Test Isolation Forest with insufficient training data"""
        # Less than 20 samples
        training_data = [
            {"memory_usage": 50, "cpu_usage": 40, "restart_count": 0}
            for _ in range(10)
        ]

        success = engine.train_anomaly_detector(training_data)

        assert success is False
        assert engine.is_trained is False

    def test_anomaly_detection(self, engine):
        """Test anomaly detection after training"""
        # Train with normal data
        training_data = []
        for i in range(30):
            training_data.append({
                "memory_usage": 50 + np.random.randint(-5, 5),
                "cpu_usage": 40 + np.random.randint(-5, 5),
                "restart_count": 0
            })
        engine.train_anomaly_detector(training_data)

        # Test with anomaly
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 99},  # Anomaly
                "cpu": {"usage_percent": 95}
            }
        )

        result = engine.predict_failure(request)

        # Should detect as high risk
        assert result.failure_probability > 0.5

    def test_multiple_failure_types(self, engine):
        """Test prediction of multiple failure types"""
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 95},
                "cpu": {"throttling_percent": 80}
            }
        )

        result = engine.predict_failure(request)

        # Should predict multiple possible failure types
        assert len(result.failure_types) >= 1

    def test_time_window_handling(self, engine):
        """Test different time windows"""
        request_short = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={"memory": {"usage_percent": 90}},
            time_window="1h"
        )

        request_long = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={"memory": {"usage_percent": 90}},
            time_window="24h"
        )

        result_short = engine.predict_failure(request_short)
        result_long = engine.predict_failure(request_long)

        assert isinstance(result_short, PredictionResult)
        assert isinstance(result_long, PredictionResult)


@pytest.mark.benchmark
class TestPredictionEnginePerformance:
    """Performance benchmarks for PredictionEngine"""

    def test_benchmark_simple_prediction(self, benchmark):
        """Benchmark simple prediction"""
        engine = PredictionEngine()
        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={"memory": {"usage_percent": 80}}
        )

        result = benchmark(engine.predict_failure, request)

        assert isinstance(result, PredictionResult)

    def test_benchmark_prediction_with_history(self, benchmark):
        """Benchmark prediction with historical data"""
        engine = PredictionEngine()

        history = [
            MetricSnapshot(
                timestamp=f"2025-10-01T{i:02d}:00:00Z",
                memory={"usage_percent": 50 + i * 2},
                cpu={"usage_percent": 40}
            )
            for i in range(24)
        ]

        request = PredictionRequest(
            cluster_id="cluster-001",
            resource_type="pod",
            resource_name="my-app",
            metrics={
                "memory": {"usage_percent": 90},
                "history": history
            }
        )

        result = benchmark(engine.predict_failure, request)

        assert isinstance(result, PredictionResult)
