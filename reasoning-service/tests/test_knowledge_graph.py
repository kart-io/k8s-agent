"""
Unit tests for Knowledge Graph
"""
import pytest
from datetime import datetime
from unittest.mock import Mock, patch, MagicMock

from pkg.types import (
    CaseStudy, SimilarCase, AnalysisContext, RootCauseType
)
from internal.knowledge.graph import KnowledgeGraph


class TestKnowledgeGraph:
    """Test suite for KnowledgeGraph"""

    @pytest.fixture
    def graph(self):
        """Create knowledge graph instance with in-memory storage"""
        # Force in-memory storage by patching NEO4J_AVAILABLE
        with patch('internal.knowledge.graph.NEO4J_AVAILABLE', False):
            graph = KnowledgeGraph()
        return graph

    @pytest.fixture
    def sample_case(self):
        """Create sample case study"""
        return CaseStudy(
            id="case-001",
            title="OOM Killed Pod",
            description="Application pod was killed due to out of memory",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Increase memory limits",
            outcome="Pod stabilized after memory increase",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["OOMKilled", "high memory usage", "container restart"],
            metadata={"severity": "high"}
        )

    def test_init_memory_storage(self, graph):
        """Test initialization with in-memory storage"""
        assert graph.driver is None
        assert isinstance(graph.cases, dict)
        assert isinstance(graph.nodes, dict)
        assert isinstance(graph.relations, list)

    def test_add_case_study_memory(self, graph, sample_case):
        """Test adding case study to memory storage"""
        result = graph.add_case_study(sample_case)

        assert result is True
        assert sample_case.id in graph.cases
        assert graph.cases[sample_case.id].title == "OOM Killed Pod"

    def test_add_multiple_cases(self, graph):
        """Test adding multiple case studies"""
        cases = [
            CaseStudy(
                id=f"case-{i:03d}",
                title=f"Case {i}",
                description="Test case",
                root_cause=RootCauseType.OOMKiller.value,
                solution="Test solution",
                outcome="Test outcome",
                cluster_id="cluster-001",
                timestamp=datetime.now(),
                symptoms=["test"],
                metadata={}
            )
            for i in range(5)
        ]

        for case in cases:
            result = graph.add_case_study(case)
            assert result is True

        assert len(graph.cases) == 5

    def test_find_similar_cases_empty(self, graph):
        """Test finding similar cases with empty graph"""
        context = AnalysisContext(
            event={"reason": "OOMKilled"}
        )

        similar = graph.find_similar_cases(context)

        assert isinstance(similar, list)
        assert len(similar) == 0

    def test_find_similar_cases_with_root_cause_filter(self, graph):
        """Test finding similar cases with root cause filter"""
        # Add cases with different root causes
        oom_case = CaseStudy(
            id="case-oom",
            title="OOM Case",
            description="OOM issue",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Increase memory",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["OOMKilled"],
            metadata={}
        )

        network_case = CaseStudy(
            id="case-network",
            title="Network Case",
            description="Network issue",
            root_cause=RootCauseType.NetworkError.value,
            solution="Check connectivity",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["Connection refused"],
            metadata={}
        )

        graph.add_case_study(oom_case)
        graph.add_case_study(network_case)

        context = AnalysisContext(
            event={"reason": "OOMKilled"}
        )

        # Find only OOM cases
        similar = graph.find_similar_cases(
            context,
            root_cause_type=RootCauseType.OOMKiller,
            limit=5
        )

        assert len(similar) == 1
        assert similar[0].case_id == "case-oom"
        assert similar[0].root_cause == RootCauseType.OOMKiller.value

    def test_find_similar_cases_all_types(self, graph):
        """Test finding similar cases without filter"""
        # Add cases with different root causes
        oom_case = CaseStudy(
            id="case-oom",
            title="OOM Case",
            description="OOM issue",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Increase memory",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["OOMKilled"],
            metadata={}
        )

        network_case = CaseStudy(
            id="case-network",
            title="Network Case",
            description="Network issue",
            root_cause=RootCauseType.NetworkError.value,
            solution="Check connectivity",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["Connection refused"],
            metadata={}
        )

        graph.add_case_study(oom_case)
        graph.add_case_study(network_case)

        context = AnalysisContext(
            event={"reason": "Error"}
        )

        similar = graph.find_similar_cases(context, limit=10)

        assert len(similar) == 2

    def test_find_similar_cases_limit(self, graph):
        """Test finding similar cases with limit"""
        # Add 10 cases
        for i in range(10):
            case = CaseStudy(
                id=f"case-{i:03d}",
                title=f"Case {i}",
                description="Test case",
                root_cause=RootCauseType.OOMKiller.value,
                solution="Test solution",
                outcome="Test outcome",
                cluster_id="cluster-001",
                timestamp=datetime.now(),
                symptoms=["OOMKilled"],
                metadata={}
            )
            graph.add_case_study(case)

        context = AnalysisContext(
            event={"reason": "OOMKilled"}
        )

        similar = graph.find_similar_cases(context, limit=3)

        assert len(similar) == 3

    def test_similarity_calculation_event_match(self, graph):
        """Test similarity calculation with event matching"""
        case = CaseStudy(
            id="case-001",
            title="OOM Case",
            description="OOM issue",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Increase memory",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["OOMKilled", "container restart"],
            metadata={}
        )

        graph.add_case_study(case)

        context = AnalysisContext(
            event={"reason": "OOMKilled"}
        )

        similar = graph.find_similar_cases(context)

        assert len(similar) == 1
        assert similar[0].similarity_score > 0.5  # Should have good match

    def test_similarity_calculation_log_match(self, graph):
        """Test similarity calculation with log matching"""
        case = CaseStudy(
            id="case-001",
            title="OOM Case",
            description="OOM issue",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Increase memory",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["memory", "OOMKilled"],
            metadata={}
        )

        graph.add_case_study(case)

        context = AnalysisContext(
            logs="Container killed due to memory pressure"
        )

        similar = graph.find_similar_cases(context)

        assert len(similar) == 1
        assert similar[0].similarity_score > 0.3

    def test_similar_cases_sorted_by_similarity(self, graph):
        """Test that similar cases are sorted by similarity score"""
        # Add case with high similarity
        high_sim_case = CaseStudy(
            id="case-high",
            title="High Similarity",
            description="OOM issue",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Solution",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["OOMKilled", "high memory usage"],
            metadata={}
        )

        # Add case with low similarity
        low_sim_case = CaseStudy(
            id="case-low",
            title="Low Similarity",
            description="Different issue",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Solution",
            outcome="Resolved",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["other"],
            metadata={}
        )

        graph.add_case_study(low_sim_case)
        graph.add_case_study(high_sim_case)

        context = AnalysisContext(
            event={"reason": "OOMKilled"},
            logs="high memory usage detected"
        )

        similar = graph.find_similar_cases(context, limit=5)

        # High similarity case should come first
        assert similar[0].case_id == "case-high"
        assert similar[0].similarity_score >= similar[1].similarity_score

    def test_add_feedback_memory(self, graph, sample_case):
        """Test adding feedback in memory storage"""
        graph.add_case_study(sample_case)

        feedback = {
            "feedback_id": "fb-001",
            "rating": 5,
            "was_helpful": True,
            "comments": "Very helpful solution"
        }

        result = graph.add_feedback(sample_case.id, feedback)

        assert result is True
        assert "feedback" in graph.cases[sample_case.id].metadata
        assert len(graph.cases[sample_case.id].metadata["feedback"]) == 1
        assert graph.cases[sample_case.id].metadata["feedback"][0]["rating"] == 5

    def test_add_feedback_nonexistent_case(self, graph):
        """Test adding feedback to non-existent case"""
        feedback = {
            "feedback_id": "fb-001",
            "rating": 5
        }

        result = graph.add_feedback("nonexistent-case", feedback)

        assert result is False

    def test_add_multiple_feedbacks(self, graph, sample_case):
        """Test adding multiple feedbacks to same case"""
        graph.add_case_study(sample_case)

        feedbacks = [
            {"feedback_id": f"fb-{i:03d}", "rating": i, "was_helpful": True}
            for i in range(3)
        ]

        for fb in feedbacks:
            result = graph.add_feedback(sample_case.id, fb)
            assert result is True

        assert len(graph.cases[sample_case.id].metadata["feedback"]) == 3

    def test_get_statistics_empty(self, graph):
        """Test getting statistics from empty graph"""
        stats = graph.get_statistics()

        assert stats["total_cases"] == 0
        assert stats["root_cause_types"] == 0
        assert stats["storage"] == "memory"

    def test_get_statistics_with_cases(self, graph):
        """Test getting statistics with cases"""
        # Add cases with different root causes
        for i, root_cause in enumerate([RootCauseType.OOMKiller, RootCauseType.NetworkError, RootCauseType.OOMKiller]):
            case = CaseStudy(
                id=f"case-{i:03d}",
                title=f"Case {i}",
                description="Test case",
                root_cause=root_cause.value,
                solution="Test solution",
                outcome="Test outcome",
                cluster_id="cluster-001",
                timestamp=datetime.now(),
                symptoms=["test"],
                metadata={}
            )
            graph.add_case_study(case)

        stats = graph.get_statistics()

        assert stats["total_cases"] == 3
        assert stats["root_cause_types"] == 2  # OOMKiller and NetworkError
        assert stats["storage"] == "memory"

    def test_close_without_driver(self, graph):
        """Test closing graph without Neo4j driver"""
        # Should not raise any exceptions
        graph.close()


@pytest.mark.integration
class TestKnowledgeGraphNeo4j:
    """Integration tests for Neo4j backend (requires Neo4j running)"""

    @pytest.fixture
    def neo4j_graph(self):
        """Create knowledge graph with Neo4j connection"""
        try:
            graph = KnowledgeGraph(
                uri="bolt://localhost:7687",
                user="neo4j",
                password="password"
            )
            if graph.driver is None:
                pytest.skip("Neo4j not available")
            yield graph
            graph.close()
        except Exception:
            pytest.skip("Neo4j not available")

    def test_neo4j_add_case(self, neo4j_graph):
        """Test adding case to Neo4j"""
        case = CaseStudy(
            id="case-neo4j-001",
            title="Neo4j Test Case",
            description="Test case for Neo4j",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Test solution",
            outcome="Test outcome",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["test"],
            metadata={}
        )

        result = neo4j_graph.add_case_study(case)

        assert result is True

    def test_neo4j_find_similar(self, neo4j_graph):
        """Test finding similar cases in Neo4j"""
        case = CaseStudy(
            id="case-neo4j-002",
            title="Neo4j Similar Test",
            description="Test case for similarity",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Test solution",
            outcome="Test outcome",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["OOMKilled"],
            metadata={}
        )

        neo4j_graph.add_case_study(case)

        context = AnalysisContext(
            event={"reason": "OOMKilled"}
        )

        similar = neo4j_graph.find_similar_cases(context)

        assert isinstance(similar, list)

    def test_neo4j_add_feedback(self, neo4j_graph):
        """Test adding feedback in Neo4j"""
        case = CaseStudy(
            id="case-neo4j-003",
            title="Neo4j Feedback Test",
            description="Test case for feedback",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Test solution",
            outcome="Test outcome",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["test"],
            metadata={}
        )

        neo4j_graph.add_case_study(case)

        feedback = {
            "feedback_id": "fb-neo4j-001",
            "rating": 5,
            "was_helpful": True
        }

        result = neo4j_graph.add_feedback(case.id, feedback)

        assert result is True

    def test_neo4j_get_statistics(self, neo4j_graph):
        """Test getting statistics from Neo4j"""
        stats = neo4j_graph.get_statistics()

        assert "total_cases" in stats
        assert "root_cause_types" in stats
        assert stats["storage"] == "neo4j"


@pytest.mark.benchmark
class TestKnowledgeGraphPerformance:
    """Performance benchmarks for KnowledgeGraph"""

    def test_benchmark_add_case(self, benchmark):
        """Benchmark adding case study"""
        with patch('internal.knowledge.graph.NEO4J_AVAILABLE', False):
            graph = KnowledgeGraph()

        case = CaseStudy(
            id="case-bench",
            title="Benchmark Case",
            description="Test case",
            root_cause=RootCauseType.OOMKiller.value,
            solution="Test solution",
            outcome="Test outcome",
            cluster_id="cluster-001",
            timestamp=datetime.now(),
            symptoms=["test"],
            metadata={}
        )

        result = benchmark(graph.add_case_study, case)

        assert result is True

    def test_benchmark_find_similar_cases(self, benchmark):
        """Benchmark finding similar cases"""
        with patch('internal.knowledge.graph.NEO4J_AVAILABLE', False):
            graph = KnowledgeGraph()

        # Add 100 cases
        for i in range(100):
            case = CaseStudy(
                id=f"case-{i:04d}",
                title=f"Case {i}",
                description="Test case",
                root_cause=RootCauseType.OOMKiller.value,
                solution="Test solution",
                outcome="Test outcome",
                cluster_id="cluster-001",
                timestamp=datetime.now(),
                symptoms=["OOMKilled"],
                metadata={}
            )
            graph.add_case_study(case)

        context = AnalysisContext(
            event={"reason": "OOMKilled"}
        )

        similar = benchmark(graph.find_similar_cases, context, limit=5)

        assert len(similar) <= 5
