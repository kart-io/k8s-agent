"""
Learning System for Continuous Improvement
Learns from feedback and outcomes to improve analysis accuracy
"""

from typing import Dict, List, Optional
from datetime import datetime, timedelta
from loguru import logger
import json

from pkg.types import (
    Feedback, FeedbackType, AnalysisResult, RootCauseType
)


class LearningSystem:
    """Continuous learning from user feedback"""

    def __init__(self, knowledge_graph=None):
        """Initialize learning system"""
        self.knowledge_graph = knowledge_graph
        self.feedback_store: Dict[str, List[Feedback]] = {}
        self.accuracy_metrics: Dict[str, Dict] = {}
        self._init_metrics()
        logger.info("Learning system initialized")

    def _init_metrics(self):
        """Initialize accuracy metrics"""
        for root_cause in RootCauseType:
            self.accuracy_metrics[root_cause.value] = {
                "total_diagnoses": 0,
                "correct_diagnoses": 0,
                "accuracy": 0.0,
                "last_updated": datetime.now()
            }

    def process_feedback(self, feedback: Feedback) -> bool:
        """
        Process user feedback and update learning models

        Args:
            feedback: User feedback

        Returns:
            True if processed successfully
        """
        logger.info(f"Processing feedback: {feedback.feedback_id}")

        # Store feedback
        request_id = feedback.request_id
        if request_id not in self.feedback_store:
            self.feedback_store[request_id] = []
        self.feedback_store[request_id].append(feedback)

        # Update metrics based on feedback type
        if feedback.feedback_type == FeedbackType.DIAGNOSIS_ACCURACY:
            self._update_diagnosis_metrics(feedback)
        elif feedback.feedback_type == FeedbackType.RECOMMENDATION_USEFULNESS:
            self._update_recommendation_metrics(feedback)
        elif feedback.feedback_type == FeedbackType.PREDICTION_ACCURACY:
            self._update_prediction_metrics(feedback)

        # Store in knowledge graph if available
        if self.knowledge_graph and feedback.actual_root_cause:
            self._update_knowledge_graph(feedback)

        logger.debug(f"Feedback processed: rating={feedback.rating}, helpful={feedback.was_helpful}")
        return True

    def _update_diagnosis_metrics(self, feedback: Feedback):
        """Update diagnosis accuracy metrics"""
        if not feedback.actual_root_cause:
            return

        actual_cause = feedback.actual_root_cause

        # Initialize if needed
        if actual_cause not in self.accuracy_metrics:
            self.accuracy_metrics[actual_cause] = {
                "total_diagnoses": 0,
                "correct_diagnoses": 0,
                "accuracy": 0.0,
                "last_updated": datetime.now()
            }

        metrics = self.accuracy_metrics[actual_cause]
        metrics["total_diagnoses"] += 1

        # If rating >= 4 and was_helpful, consider it correct
        if feedback.rating >= 4 and feedback.was_helpful:
            metrics["correct_diagnoses"] += 1

        # Update accuracy
        metrics["accuracy"] = metrics["correct_diagnoses"] / metrics["total_diagnoses"]
        metrics["last_updated"] = datetime.now()

        logger.info(f"Updated metrics for {actual_cause}: accuracy={metrics['accuracy']:.2%}")

    def _update_recommendation_metrics(self, feedback: Feedback):
        """Update recommendation usefulness metrics"""
        # Track which recommendations are most useful
        # Could be expanded to adjust recommendation weights
        logger.debug(f"Recommendation feedback: rating={feedback.rating}")

    def _update_prediction_metrics(self, feedback: Feedback):
        """Update prediction accuracy metrics"""
        # Track prediction accuracy over time
        logger.debug(f"Prediction feedback: was_helpful={feedback.was_helpful}")

    def _update_knowledge_graph(self, feedback: Feedback):
        """Update knowledge graph with feedback"""
        try:
            feedback_data = {
                "feedback_id": feedback.feedback_id,
                "rating": feedback.rating,
                "was_helpful": feedback.was_helpful,
                "actual_root_cause": feedback.actual_root_cause,
                "actual_solution": feedback.actual_solution,
                "comments": feedback.comments,
                "timestamp": feedback.timestamp.isoformat()
            }

            # Add feedback to the case study
            self.knowledge_graph.add_feedback(feedback.request_id, feedback_data)
            logger.debug("Feedback added to knowledge graph")
        except Exception as e:
            logger.error(f"Failed to update knowledge graph: {e}")

    def get_accuracy_metrics(self, root_cause_type: Optional[str] = None) -> Dict:
        """
        Get accuracy metrics

        Args:
            root_cause_type: Optional filter by root cause type

        Returns:
            Accuracy metrics
        """
        if root_cause_type:
            return self.accuracy_metrics.get(root_cause_type, {})

        return {
            "overall": self._calculate_overall_accuracy(),
            "by_root_cause": self.accuracy_metrics
        }

    def _calculate_overall_accuracy(self) -> float:
        """Calculate overall diagnosis accuracy"""
        total = sum(m["total_diagnoses"] for m in self.accuracy_metrics.values())
        correct = sum(m["correct_diagnoses"] for m in self.accuracy_metrics.values())

        if total == 0:
            return 0.0

        return correct / total

    def suggest_improvements(self) -> List[Dict]:
        """
        Suggest improvements based on learning

        Returns:
            List of improvement suggestions
        """
        suggestions = []

        # Find root causes with low accuracy
        for root_cause, metrics in self.accuracy_metrics.items():
            if metrics["total_diagnoses"] >= 5 and metrics["accuracy"] < 0.7:
                suggestions.append({
                    "type": "low_accuracy",
                    "root_cause": root_cause,
                    "current_accuracy": metrics["accuracy"],
                    "suggestion": f"Consider improving detection patterns for {root_cause}",
                    "total_cases": metrics["total_diagnoses"]
                })

        # Find frequently misdiagnosed patterns
        misdiagnoses = self._find_common_misdiagnoses()
        for pattern in misdiagnoses:
            suggestions.append({
                "type": "common_misdiagnosis",
                "pattern": pattern,
                "suggestion": "Review and refine detection rules for this pattern"
            })

        logger.info(f"Generated {len(suggestions)} improvement suggestions")
        return suggestions

    def _find_common_misdiagnoses(self) -> List[str]:
        """Find common misdiagnosis patterns"""
        # Analyze feedback to find patterns
        patterns = []

        # Group feedback by actual root cause
        by_actual_cause: Dict[str, int] = {}
        for feedbacks in self.feedback_store.values():
            for feedback in feedbacks:
                if feedback.actual_root_cause and not feedback.was_helpful:
                    cause = feedback.actual_root_cause
                    by_actual_cause[cause] = by_actual_cause.get(cause, 0) + 1

        # Find causes with frequent misdiagnoses
        for cause, count in by_actual_cause.items():
            if count >= 3:
                patterns.append(cause)

        return patterns

    def export_learning_data(self) -> Dict:
        """
        Export learning data for analysis or backup

        Returns:
            Learning data
        """
        return {
            "accuracy_metrics": self.accuracy_metrics,
            "total_feedback": sum(len(f) for f in self.feedback_store.values()),
            "feedback_by_type": self._count_feedback_by_type(),
            "export_time": datetime.now().isoformat()
        }

    def _count_feedback_by_type(self) -> Dict[str, int]:
        """Count feedback by type"""
        counts = {ft.value: 0 for ft in FeedbackType}

        for feedbacks in self.feedback_store.values():
            for feedback in feedbacks:
                counts[feedback.feedback_type.value] += 1

        return counts

    def import_learning_data(self, data: Dict) -> bool:
        """
        Import learning data from backup

        Args:
            data: Learning data to import

        Returns:
            True if successful
        """
        try:
            if "accuracy_metrics" in data:
                self.accuracy_metrics.update(data["accuracy_metrics"])
                logger.info("Learning data imported successfully")
                return True
        except Exception as e:
            logger.error(f"Failed to import learning data: {e}")
            return False

    def analyze_trends(self, time_window: str = "7d") -> Dict:
        """
        Analyze accuracy trends over time

        Args:
            time_window: Time window for analysis (e.g., "7d", "30d")

        Returns:
            Trend analysis
        """
        # Parse time window
        days = int(time_window.rstrip('d'))
        cutoff = datetime.now() - timedelta(days=days)

        # Filter recent feedback
        recent_feedback = []
        for feedbacks in self.feedback_store.values():
            recent_feedback.extend([f for f in feedbacks if f.timestamp >= cutoff])

        # Calculate trend metrics
        if not recent_feedback:
            return {"message": "No recent feedback data"}

        helpful_count = sum(1 for f in recent_feedback if f.was_helpful)
        total_count = len(recent_feedback)

        avg_rating = sum(f.rating for f in recent_feedback) / total_count

        return {
            "time_window": time_window,
            "total_feedback": total_count,
            "helpful_rate": helpful_count / total_count,
            "average_rating": avg_rating,
            "trend": "improving" if avg_rating >= 4 else "needs_attention"
        }

    def get_top_performing_patterns(self, limit: int = 5) -> List[Dict]:
        """
        Get top performing root cause detection patterns

        Args:
            limit: Number of patterns to return

        Returns:
            List of top patterns
        """
        # Sort by accuracy
        sorted_metrics = sorted(
            self.accuracy_metrics.items(),
            key=lambda x: (x[1]["accuracy"], x[1]["total_diagnoses"]),
            reverse=True
        )

        top_patterns = []
        for root_cause, metrics in sorted_metrics[:limit]:
            if metrics["total_diagnoses"] >= 3:  # Minimum sample size
                top_patterns.append({
                    "root_cause": root_cause,
                    "accuracy": metrics["accuracy"],
                    "total_diagnoses": metrics["total_diagnoses"],
                    "correct_diagnoses": metrics["correct_diagnoses"]
                })

        return top_patterns

    def reset_metrics(self, root_cause_type: Optional[str] = None):
        """
        Reset metrics (for testing or major model updates)

        Args:
            root_cause_type: Optional specific root cause to reset
        """
        if root_cause_type:
            if root_cause_type in self.accuracy_metrics:
                self.accuracy_metrics[root_cause_type] = {
                    "total_diagnoses": 0,
                    "correct_diagnoses": 0,
                    "accuracy": 0.0,
                    "last_updated": datetime.now()
                }
                logger.info(f"Metrics reset for {root_cause_type}")
        else:
            self._init_metrics()
            self.feedback_store.clear()
            logger.info("All metrics reset")