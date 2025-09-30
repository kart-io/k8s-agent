"""
Failure Prediction Engine
Uses time series analysis and anomaly detection to predict failures
"""

import numpy as np
from typing import Dict, List, Optional, Tuple
from datetime import datetime, timedelta
from loguru import logger

try:
    from sklearn.ensemble import IsolationForest
    from sklearn.preprocessing import StandardScaler
    SKLEARN_AVAILABLE = True
except ImportError:
    SKLEARN_AVAILABLE = False
    logger.warning("scikit-learn not available, prediction features disabled")

try:
    import statsmodels.api as sm
    STATSMODELS_AVAILABLE = True
except ImportError:
    STATSMODELS_AVAILABLE = False
    logger.warning("statsmodels not available, time series prediction disabled")

from pkg.types import PredictionRequest, PredictionResult, RootCauseType


class PredictionEngine:
    """Predict potential failures based on metrics"""

    def __init__(self):
        """Initialize prediction engine"""
        self.scaler = StandardScaler() if SKLEARN_AVAILABLE else None
        self.anomaly_detector = None
        self.thresholds = self._load_thresholds()
        logger.info("Prediction engine initialized")

    def predict_failure(self, request: PredictionRequest) -> PredictionResult:
        """
        Predict failure probability and time

        Args:
            request: Prediction request with metrics

        Returns:
            Prediction result with probability and estimated time
        """
        logger.info(f"Predicting failures for {request.resource_type}/{request.resource_name}")

        # Multiple prediction methods
        predictions = []

        # 1. Threshold-based prediction
        threshold_pred = self._predict_by_threshold(request.metrics)
        if threshold_pred:
            predictions.append(threshold_pred)

        # 2. Trend-based prediction
        if "history" in request.metrics:
            trend_pred = self._predict_by_trend(request.metrics["history"])
            if trend_pred:
                predictions.append(trend_pred)

        # 3. Anomaly detection
        if SKLEARN_AVAILABLE and "history" in request.metrics:
            anomaly_pred = self._predict_by_anomaly(request.metrics["history"])
            if anomaly_pred:
                predictions.append(anomaly_pred)

        # Combine predictions
        if not predictions:
            logger.info("No failure predicted")
            return PredictionResult(
                failure_probability=0.0,
                confidence=0.5,
                contributing_factors=["Insufficient data for prediction"]
            )

        # Aggregate predictions
        result = self._aggregate_predictions(predictions)
        logger.info(f"Failure probability: {result.failure_probability:.2f}")
        return result

    def _predict_by_threshold(self, metrics: Dict) -> Optional[Tuple[float, List[str], List[str]]]:
        """Predict based on metric thresholds"""
        failure_prob = 0.0
        failure_types = []
        factors = []

        # Check memory usage
        if "memory" in metrics:
            mem = metrics["memory"]
            if isinstance(mem, dict):
                usage = mem.get("usage_percent", 0)
                if usage >= 95:
                    failure_prob = max(failure_prob, 0.9)
                    failure_types.append(RootCauseType.OOM_KILLER.value)
                    factors.append(f"Memory usage at {usage}%")
                elif usage >= 85:
                    failure_prob = max(failure_prob, 0.6)
                    failure_types.append(RootCauseType.OOM_KILLER.value)
                    factors.append(f"Memory usage approaching limit ({usage}%)")

        # Check CPU throttling
        if "cpu" in metrics:
            cpu = metrics["cpu"]
            if isinstance(cpu, dict):
                throttling = cpu.get("throttling_percent", 0)
                usage = cpu.get("usage_percent", 0)
                if throttling >= 70:
                    failure_prob = max(failure_prob, 0.8)
                    failure_types.append(RootCauseType.CPU_THROTTLING.value)
                    factors.append(f"CPU throttling at {throttling}%")
                elif usage >= 90:
                    failure_prob = max(failure_prob, 0.5)
                    failure_types.append(RootCauseType.CPU_THROTTLING.value)
                    factors.append(f"High CPU usage ({usage}%)")

        # Check disk usage
        if "disk" in metrics:
            disk = metrics["disk"]
            if isinstance(disk, dict):
                usage = disk.get("usage_percent", 0)
                if usage >= 95:
                    failure_prob = max(failure_prob, 0.85)
                    failure_types.append(RootCauseType.DISK_PRESSURE.value)
                    factors.append(f"Disk usage at {usage}%")
                elif usage >= 85:
                    failure_prob = max(failure_prob, 0.5)
                    failure_types.append(RootCauseType.DISK_PRESSURE.value)
                    factors.append(f"Disk usage approaching limit ({usage}%)")

        # Check network errors
        if "network" in metrics:
            net = metrics["network"]
            if isinstance(net, dict):
                error_rate = net.get("error_rate", 0)
                if error_rate >= 0.1:  # 10% error rate
                    failure_prob = max(failure_prob, 0.7)
                    failure_types.append(RootCauseType.NETWORK_ERROR.value)
                    factors.append(f"Network error rate at {error_rate*100:.1f}%")

        # Check restart count
        if "restart_count" in metrics:
            restarts = metrics["restart_count"]
            if restarts >= 5:
                failure_prob = max(failure_prob, 0.8)
                failure_types.append(RootCauseType.CONFIG_ERROR.value)
                factors.append(f"Pod restarted {restarts} times")

        if failure_prob > 0:
            return (failure_prob, failure_types, factors)
        return None

    def _predict_by_trend(self, history: List[Dict]) -> Optional[Tuple[float, List[str], List[str]]]:
        """Predict based on metric trends"""
        if not history or len(history) < 3:
            return None

        failure_prob = 0.0
        failure_types = []
        factors = []

        # Analyze memory trend
        mem_values = [h.get("memory", {}).get("usage_percent", 0) for h in history if "memory" in h]
        if len(mem_values) >= 3:
            trend = self._calculate_trend(mem_values)
            if trend > 5:  # Increasing by 5% per interval
                # Predict time to 100%
                current = mem_values[-1]
                time_to_failure = (100 - current) / trend if trend > 0 else None
                if time_to_failure and time_to_failure < 10:  # Less than 10 intervals
                    failure_prob = max(failure_prob, 0.7)
                    failure_types.append(RootCauseType.OOM_KILLER.value)
                    factors.append(f"Memory increasing at {trend:.1f}% per interval, {time_to_failure:.1f} intervals to exhaustion")

        # Analyze disk trend
        disk_values = [h.get("disk", {}).get("usage_percent", 0) for h in history if "disk" in h]
        if len(disk_values) >= 3:
            trend = self._calculate_trend(disk_values)
            if trend > 3:  # Increasing by 3% per interval
                current = disk_values[-1]
                time_to_failure = (100 - current) / trend if trend > 0 else None
                if time_to_failure and time_to_failure < 20:
                    failure_prob = max(failure_prob, 0.6)
                    failure_types.append(RootCauseType.DISK_PRESSURE.value)
                    factors.append(f"Disk usage increasing at {trend:.1f}% per interval")

        # Analyze restart frequency
        restart_values = [h.get("restart_count", 0) for h in history]
        if len(restart_values) >= 3:
            restart_rate = restart_values[-1] - restart_values[0]
            if restart_rate >= 3:  # 3 or more restarts in time window
                failure_prob = max(failure_prob, 0.75)
                failure_types.append(RootCauseType.CONFIG_ERROR.value)
                factors.append(f"Restart frequency increasing ({restart_rate} restarts in window)")

        if failure_prob > 0:
            return (failure_prob, failure_types, factors)
        return None

    def _predict_by_anomaly(self, history: List[Dict]) -> Optional[Tuple[float, List[str], List[str]]]:
        """Predict using anomaly detection"""
        if not SKLEARN_AVAILABLE or len(history) < 5:
            return None

        try:
            # Extract features
            features = []
            for h in history:
                feature_vector = [
                    h.get("memory", {}).get("usage_percent", 0),
                    h.get("cpu", {}).get("usage_percent", 0),
                    h.get("disk", {}).get("usage_percent", 0),
                    h.get("network", {}).get("error_rate", 0) * 100,
                    h.get("restart_count", 0)
                ]
                features.append(feature_vector)

            features_array = np.array(features)

            # Train anomaly detector if not exists
            if self.anomaly_detector is None:
                self.anomaly_detector = IsolationForest(contamination=0.1, random_state=42)
                self.anomaly_detector.fit(features_array)

            # Predict anomalies
            predictions = self.anomaly_detector.predict(features_array)
            scores = self.anomaly_detector.score_samples(features_array)

            # Check if recent data points are anomalies
            recent_anomalies = sum(1 for p in predictions[-3:] if p == -1)
            if recent_anomalies >= 2:
                # Recent data shows anomalies
                anomaly_score = abs(scores[-1])  # More negative = more anomalous
                failure_prob = min(anomaly_score * 0.5, 0.8)

                return (
                    failure_prob,
                    [RootCauseType.UNKNOWN.value],
                    ["Anomalous metrics pattern detected"]
                )

        except Exception as e:
            logger.error(f"Anomaly detection failed: {e}")

        return None

    def _calculate_trend(self, values: List[float]) -> float:
        """Calculate linear trend (slope)"""
        if len(values) < 2:
            return 0.0

        n = len(values)
        x = list(range(n))
        y = values

        # Simple linear regression
        x_mean = sum(x) / n
        y_mean = sum(y) / n

        numerator = sum((x[i] - x_mean) * (y[i] - y_mean) for i in range(n))
        denominator = sum((x[i] - x_mean) ** 2 for i in range(n))

        if denominator == 0:
            return 0.0

        slope = numerator / denominator
        return slope

    def _aggregate_predictions(self, predictions: List[Tuple[float, List[str], List[str]]]) -> PredictionResult:
        """Aggregate multiple predictions"""
        # Use maximum probability
        max_prob = max(p[0] for p in predictions)

        # Collect all failure types
        all_types = []
        for _, types, _ in predictions:
            all_types.extend(types)
        failure_types = list(set(all_types))

        # Collect all factors
        all_factors = []
        for _, _, factors in predictions:
            all_factors.extend(factors)

        # Estimate failure time based on probability
        predicted_time = None
        if max_prob >= 0.8:
            predicted_time = datetime.now() + timedelta(hours=1)
        elif max_prob >= 0.6:
            predicted_time = datetime.now() + timedelta(hours=6)
        elif max_prob >= 0.4:
            predicted_time = datetime.now() + timedelta(hours=24)

        # Confidence based on number of agreeing predictions
        confidence = min(len(predictions) * 0.3, 0.9)

        return PredictionResult(
            failure_probability=max_prob,
            predicted_failure_time=predicted_time,
            failure_types=failure_types,
            confidence=confidence,
            contributing_factors=all_factors
        )

    def _load_thresholds(self) -> Dict:
        """Load threshold configurations"""
        return {
            "memory": {
                "critical": 95,
                "warning": 85
            },
            "cpu": {
                "critical": 90,
                "warning": 75
            },
            "disk": {
                "critical": 95,
                "warning": 85
            },
            "network_error_rate": {
                "critical": 0.1,
                "warning": 0.05
            },
            "restart_count": {
                "critical": 5,
                "warning": 3
            }
        }

    def update_thresholds(self, thresholds: Dict):
        """Update threshold configurations"""
        self.thresholds.update(thresholds)
        logger.info("Thresholds updated")

    def train_anomaly_detector(self, historical_data: List[Dict]):
        """Train anomaly detector with historical data"""
        if not SKLEARN_AVAILABLE:
            logger.warning("scikit-learn not available, cannot train anomaly detector")
            return

        try:
            features = []
            for data in historical_data:
                feature_vector = [
                    data.get("memory", {}).get("usage_percent", 0),
                    data.get("cpu", {}).get("usage_percent", 0),
                    data.get("disk", {}).get("usage_percent", 0),
                    data.get("network", {}).get("error_rate", 0) * 100,
                    data.get("restart_count", 0)
                ]
                features.append(feature_vector)

            features_array = np.array(features)

            # Train new model
            self.anomaly_detector = IsolationForest(contamination=0.1, random_state=42)
            self.anomaly_detector.fit(features_array)

            logger.info(f"Anomaly detector trained with {len(historical_data)} samples")
        except Exception as e:
            logger.error(f"Failed to train anomaly detector: {e}")