"""
Knowledge Graph for Failure Patterns
Stores and retrieves historical cases, patterns, and relationships
"""

from typing import List, Dict, Optional
from datetime import datetime
from loguru import logger

try:
    from neo4j import GraphDatabase
    NEO4J_AVAILABLE = True
except ImportError:
    NEO4J_AVAILABLE = False
    logger.warning("Neo4j driver not available, using in-memory fallback")

from pkg.types import (
    CaseStudy, KnowledgeNode, KnowledgeRelation, SimilarCase,
    RootCauseType, AnalysisContext
)


class KnowledgeGraph:
    """Knowledge graph for storing and querying failure patterns"""

    def __init__(self, uri: str = "bolt://localhost:7687", user: str = "neo4j", password: str = "password"):
        """Initialize knowledge graph"""
        self.uri = uri
        self.user = user
        self.password = password

        if NEO4J_AVAILABLE:
            try:
                self.driver = GraphDatabase.driver(uri, auth=(user, password))
                logger.info("Connected to Neo4j knowledge graph")
                self._create_indexes()
            except Exception as e:
                logger.warning(f"Failed to connect to Neo4j: {e}, using in-memory fallback")
                self.driver = None
                self._init_memory_storage()
        else:
            self.driver = None
            self._init_memory_storage()

    def _init_memory_storage(self):
        """Initialize in-memory storage as fallback"""
        self.cases: Dict[str, CaseStudy] = {}
        self.nodes: Dict[str, KnowledgeNode] = {}
        self.relations: List[KnowledgeRelation] = []
        logger.info("Using in-memory knowledge storage")

    def _create_indexes(self):
        """Create indexes for faster queries"""
        if not self.driver:
            return

        with self.driver.session() as session:
            try:
                # Create indexes on frequently queried fields
                session.run("CREATE INDEX IF NOT EXISTS FOR (c:CaseStudy) ON (c.id)")
                session.run("CREATE INDEX IF NOT EXISTS FOR (c:CaseStudy) ON (c.root_cause)")
                session.run("CREATE INDEX IF NOT EXISTS FOR (c:CaseStudy) ON (c.cluster_id)")
                logger.debug("Knowledge graph indexes created")
            except Exception as e:
                logger.error(f"Failed to create indexes: {e}")

    def add_case_study(self, case: CaseStudy) -> bool:
        """
        Add a case study to knowledge graph

        Args:
            case: Case study to add

        Returns:
            True if successful
        """
        logger.info(f"Adding case study: {case.id}")

        if self.driver:
            return self._add_case_neo4j(case)
        else:
            return self._add_case_memory(case)

    def _add_case_neo4j(self, case: CaseStudy) -> bool:
        """Add case to Neo4j"""
        query = """
        MERGE (c:CaseStudy {id: $id})
        SET c.title = $title,
            c.description = $description,
            c.root_cause = $root_cause,
            c.solution = $solution,
            c.outcome = $outcome,
            c.cluster_id = $cluster_id,
            c.timestamp = datetime($timestamp),
            c.symptoms = $symptoms,
            c.metadata = $metadata
        RETURN c
        """

        try:
            with self.driver.session() as session:
                session.run(query, {
                    "id": case.id,
                    "title": case.title,
                    "description": case.description,
                    "root_cause": case.root_cause,
                    "solution": case.solution,
                    "outcome": case.outcome,
                    "cluster_id": case.cluster_id,
                    "timestamp": case.timestamp.isoformat(),
                    "symptoms": case.symptoms,
                    "metadata": case.metadata
                })
                logger.debug(f"Case {case.id} added to Neo4j")
                return True
        except Exception as e:
            logger.error(f"Failed to add case to Neo4j: {e}")
            return False

    def _add_case_memory(self, case: CaseStudy) -> bool:
        """Add case to in-memory storage"""
        self.cases[case.id] = case
        logger.debug(f"Case {case.id} added to memory storage")
        return True

    def find_similar_cases(
        self,
        context: AnalysisContext,
        root_cause_type: Optional[RootCauseType] = None,
        limit: int = 5
    ) -> List[SimilarCase]:
        """
        Find similar historical cases

        Args:
            context: Analysis context
            root_cause_type: Optional root cause filter
            limit: Maximum number of cases

        Returns:
            List of similar cases sorted by similarity
        """
        logger.info("Finding similar cases")

        if self.driver:
            cases = self._find_similar_neo4j(context, root_cause_type, limit)
        else:
            cases = self._find_similar_memory(context, root_cause_type, limit)

        logger.info(f"Found {len(cases)} similar cases")
        return cases

    def _find_similar_neo4j(
        self,
        context: AnalysisContext,
        root_cause_type: Optional[RootCauseType],
        limit: int
    ) -> List[SimilarCase]:
        """Find similar cases in Neo4j"""
        # Build query based on available context
        conditions = []
        params = {"limit": limit}

        if root_cause_type:
            conditions.append("c.root_cause = $root_cause")
            params["root_cause"] = root_cause_type.value

        where_clause = "WHERE " + " AND ".join(conditions) if conditions else ""

        query = f"""
        MATCH (c:CaseStudy)
        {where_clause}
        RETURN c.id as case_id,
               c.description as description,
               c.root_cause as root_cause,
               c.solution as solution,
               c.outcome as outcome,
               c.timestamp as timestamp
        ORDER BY c.timestamp DESC
        LIMIT $limit
        """

        try:
            with self.driver.session() as session:
                result = session.run(query, params)
                cases = []
                for record in result:
                    # Calculate similarity score based on context
                    similarity = self._calculate_similarity(context, record)

                    cases.append(SimilarCase(
                        case_id=record["case_id"],
                        description=record["description"],
                        similarity_score=similarity,
                        root_cause=record["root_cause"],
                        solution=record["solution"],
                        outcome=record["outcome"],
                        timestamp=datetime.fromisoformat(str(record["timestamp"]))
                    ))

                # Sort by similarity
                cases.sort(key=lambda x: x.similarity_score, reverse=True)
                return cases[:limit]
        except Exception as e:
            logger.error(f"Failed to query similar cases: {e}")
            return []

    def _find_similar_memory(
        self,
        context: AnalysisContext,
        root_cause_type: Optional[RootCauseType],
        limit: int
    ) -> List[SimilarCase]:
        """Find similar cases in memory"""
        candidates = []

        for case in self.cases.values():
            # Filter by root cause if specified
            if root_cause_type and case.root_cause != root_cause_type.value:
                continue

            # Calculate similarity
            similarity = self._calculate_similarity_from_case(context, case)

            candidates.append(SimilarCase(
                case_id=case.id,
                description=case.description,
                similarity_score=similarity,
                root_cause=case.root_cause,
                solution=case.solution,
                outcome=case.outcome,
                timestamp=case.timestamp
            ))

        # Sort by similarity and return top N
        candidates.sort(key=lambda x: x.similarity_score, reverse=True)
        return candidates[:limit]

    def _calculate_similarity(self, context: AnalysisContext, record: Dict) -> float:
        """Calculate similarity score between context and case"""
        score = 0.5  # Base score

        # Compare event reasons
        if context.event and "reason" in context.event:
            # Would need to fetch symptoms from record for detailed comparison
            score += 0.2

        # Compare logs (keyword overlap)
        if context.logs:
            # Simple heuristic - in production would use text similarity
            score += 0.1

        # Time decay - more recent cases are more relevant
        # Would need to parse timestamp and apply decay

        return min(score, 1.0)

    def _calculate_similarity_from_case(self, context: AnalysisContext, case: CaseStudy) -> float:
        """Calculate similarity from full case object"""
        score = 0.3  # Base score

        # Event reason matching
        if context.event and "reason" in context.event:
            event_reason = context.event["reason"].lower()
            # Check symptoms
            for symptom in case.symptoms:
                if symptom.lower() in event_reason or event_reason in symptom.lower():
                    score += 0.3
                    break

        # Log keyword matching
        if context.logs:
            log_lower = context.logs.lower()
            matched = sum(1 for symptom in case.symptoms if symptom.lower() in log_lower)
            score += min(matched * 0.1, 0.4)

        return min(score, 1.0)

    def add_feedback(self, case_id: str, feedback: Dict) -> bool:
        """
        Add user feedback to improve recommendations

        Args:
            case_id: Case study ID
            feedback: Feedback data

        Returns:
            True if successful
        """
        logger.info(f"Adding feedback for case {case_id}")

        if self.driver:
            return self._add_feedback_neo4j(case_id, feedback)
        else:
            return self._add_feedback_memory(case_id, feedback)

    def _add_feedback_neo4j(self, case_id: str, feedback: Dict) -> bool:
        """Add feedback to Neo4j"""
        query = """
        MATCH (c:CaseStudy {id: $case_id})
        CREATE (f:Feedback {
            id: $feedback_id,
            rating: $rating,
            was_helpful: $was_helpful,
            comments: $comments,
            timestamp: datetime($timestamp)
        })
        CREATE (c)-[:HAS_FEEDBACK]->(f)
        RETURN f
        """

        try:
            with self.driver.session() as session:
                session.run(query, {
                    "case_id": case_id,
                    "feedback_id": feedback.get("feedback_id"),
                    "rating": feedback.get("rating"),
                    "was_helpful": feedback.get("was_helpful"),
                    "comments": feedback.get("comments", ""),
                    "timestamp": datetime.now().isoformat()
                })
                logger.debug(f"Feedback added for case {case_id}")
                return True
        except Exception as e:
            logger.error(f"Failed to add feedback: {e}")
            return False

    def _add_feedback_memory(self, case_id: str, feedback: Dict) -> bool:
        """Add feedback to memory storage"""
        if case_id in self.cases:
            case = self.cases[case_id]
            if "feedback" not in case.metadata:
                case.metadata["feedback"] = []
            case.metadata["feedback"].append(feedback)
            logger.debug(f"Feedback added to case {case_id}")
            return True
        return False

    def get_statistics(self) -> Dict:
        """Get knowledge graph statistics"""
        if self.driver:
            return self._get_stats_neo4j()
        else:
            return self._get_stats_memory()

    def _get_stats_neo4j(self) -> Dict:
        """Get statistics from Neo4j"""
        query = """
        MATCH (c:CaseStudy)
        RETURN count(c) as total_cases,
               collect(DISTINCT c.root_cause) as root_causes
        """

        try:
            with self.driver.session() as session:
                result = session.run(query)
                record = result.single()
                return {
                    "total_cases": record["total_cases"],
                    "root_cause_types": len(record["root_causes"]),
                    "storage": "neo4j"
                }
        except Exception as e:
            logger.error(f"Failed to get stats: {e}")
            return {"error": str(e)}

    def _get_stats_memory(self) -> Dict:
        """Get statistics from memory"""
        root_causes = set(case.root_cause for case in self.cases.values())
        return {
            "total_cases": len(self.cases),
            "root_cause_types": len(root_causes),
            "storage": "memory"
        }

    def close(self):
        """Close database connection"""
        if self.driver:
            self.driver.close()
            logger.info("Neo4j connection closed")