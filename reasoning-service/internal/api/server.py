"""
FastAPI Server for Reasoning Service
Provides REST API for AI analysis, prediction, and recommendations
"""

import time
from typing import Dict, Optional
from fastapi import FastAPI, HTTPException, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from loguru import logger
from contextlib import asynccontextmanager

from pkg.types import (
    AnalysisRequest, AnalysisResponse, AnalysisResult, AnalysisType,
    PredictionRequest, PredictionResult, Feedback, CaseStudy
)
from internal.analyzer.root_cause import RootCauseAnalyzer
from internal.recommender.engine import RecommendationEngine
from internal.knowledge.graph import KnowledgeGraph
from internal.predictor.engine import PredictionEngine
from internal.learning.system import LearningSystem


# Global service components
analyzer: Optional[RootCauseAnalyzer] = None
recommender: Optional[RecommendationEngine] = None
knowledge_graph: Optional[KnowledgeGraph] = None
predictor: Optional[PredictionEngine] = None
learning_system: Optional[LearningSystem] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifecycle manager for startup and shutdown"""
    # Startup
    global analyzer, recommender, knowledge_graph, predictor, learning_system

    logger.info("Initializing Reasoning Service components...")

    try:
        # Initialize components
        analyzer = RootCauseAnalyzer()
        recommender = RecommendationEngine()
        knowledge_graph = KnowledgeGraph(
            uri=app.state.config.get("neo4j_uri", "bolt://localhost:7687"),
            user=app.state.config.get("neo4j_user", "neo4j"),
            password=app.state.config.get("neo4j_password", "password")
        )
        predictor = PredictionEngine()
        learning_system = LearningSystem(knowledge_graph=knowledge_graph)

        logger.info("All components initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize components: {e}")
        raise

    yield

    # Shutdown
    logger.info("Shutting down Reasoning Service...")
    if knowledge_graph:
        knowledge_graph.close()
    logger.info("Shutdown complete")


# Create FastAPI app
app = FastAPI(
    title="Aetherius Reasoning Service",
    description="AI-powered root cause analysis and failure prediction",
    version="1.0.0",
    lifespan=lifespan
)

# Initialize config storage
app.state.config = {}

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Request logging middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    """Log all requests"""
    start_time = time.time()

    # Log request
    logger.info(f"→ {request.method} {request.url.path}")

    # Process request
    response = await call_next(request)

    # Log response
    duration = time.time() - start_time
    logger.info(f"← {request.method} {request.url.path} - {response.status_code} ({duration:.3f}s)")

    return response


# Health check endpoint
@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "reasoning-service",
        "components": {
            "analyzer": analyzer is not None,
            "recommender": recommender is not None,
            "knowledge_graph": knowledge_graph is not None,
            "predictor": predictor is not None,
            "learning_system": learning_system is not None
        }
    }


# Root cause analysis endpoint
@app.post("/api/v1/analyze/root-cause", response_model=AnalysisResponse)
async def analyze_root_cause(request: AnalysisRequest):
    """
    Analyze root cause of failures

    Args:
        request: Analysis request with context

    Returns:
        Analysis result with root cause and recommendations
    """
    logger.info(f"Analyzing root cause for request: {request.request_id}")
    start_time = time.time()

    try:
        # Validate components
        if not analyzer or not recommender:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Analysis components not initialized"
            )

        # Perform root cause analysis
        analysis_result = analyzer.analyze(request.context)

        # Generate recommendations if root cause identified
        if analysis_result.root_cause:
            recommendations = recommender.recommend(
                analysis_result.root_cause,
                request.context
            )
            analysis_result.recommendations = recommendations

            # Find similar cases if requested
            if request.options.include_similar_cases and knowledge_graph:
                similar_cases = knowledge_graph.find_similar_cases(
                    request.context,
                    analysis_result.root_cause.type,
                    limit=request.options.max_recommendations
                )
                analysis_result.similar_cases = similar_cases

        # Calculate processing time
        processing_time = time.time() - start_time

        return AnalysisResponse(
            request_id=request.request_id,
            status="completed",
            result=analysis_result,
            processing_time=processing_time
        )

    except Exception as e:
        logger.error(f"Analysis failed: {e}")
        processing_time = time.time() - start_time

        return AnalysisResponse(
            request_id=request.request_id,
            status="failed",
            error=str(e),
            processing_time=processing_time
        )


# Prediction endpoint
@app.post("/api/v1/analyze/predict", response_model=PredictionResult)
async def predict_failure(request: PredictionRequest):
    """
    Predict potential failures

    Args:
        request: Prediction request with metrics

    Returns:
        Prediction result with probability and estimated time
    """
    logger.info(f"Predicting failures for {request.resource_type}/{request.resource_name}")

    try:
        if not predictor:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Predictor not initialized"
            )

        result = predictor.predict_failure(request)
        return result

    except Exception as e:
        logger.error(f"Prediction failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


# Recommendation endpoint
@app.post("/api/v1/analyze/recommend")
async def get_recommendations(request: AnalysisRequest):
    """
    Get recommendations based on analysis context

    Args:
        request: Analysis request

    Returns:
        Recommendations
    """
    logger.info(f"Getting recommendations for request: {request.request_id}")

    try:
        if not analyzer or not recommender:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Analysis components not initialized"
            )

        # Quick root cause analysis
        analysis_result = analyzer.analyze(request.context)

        if not analysis_result.root_cause:
            return {"recommendations": [], "message": "No root cause identified"}

        # Get recommendations
        recommendations = recommender.recommend(
            analysis_result.root_cause,
            request.context
        )

        return {
            "root_cause": analysis_result.root_cause.dict(),
            "recommendations": [r.dict() for r in recommendations]
        }

    except Exception as e:
        logger.error(f"Recommendation generation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


# Feedback endpoint
@app.post("/api/v1/feedback")
async def submit_feedback(feedback: Feedback):
    """
    Submit feedback for learning

    Args:
        feedback: User feedback

    Returns:
        Confirmation
    """
    logger.info(f"Receiving feedback: {feedback.feedback_id}")

    try:
        if not learning_system:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Learning system not initialized"
            )

        success = learning_system.process_feedback(feedback)

        if success:
            return {"status": "accepted", "message": "Feedback processed successfully"}
        else:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Failed to process feedback"
            )

    except Exception as e:
        logger.error(f"Feedback processing failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


# Case study endpoints
@app.post("/api/v1/cases")
async def add_case_study(case: CaseStudy):
    """
    Add a case study to knowledge graph

    Args:
        case: Case study

    Returns:
        Confirmation
    """
    logger.info(f"Adding case study: {case.id}")

    try:
        if not knowledge_graph:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Knowledge graph not initialized"
            )

        success = knowledge_graph.add_case_study(case)

        if success:
            return {"status": "created", "case_id": case.id}
        else:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Failed to add case study"
            )

    except Exception as e:
        logger.error(f"Failed to add case study: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


@app.get("/api/v1/cases/similar")
async def find_similar_cases(event_reason: Optional[str] = None, limit: int = 5):
    """
    Find similar historical cases

    Args:
        event_reason: Optional event reason filter
        limit: Maximum number of cases

    Returns:
        List of similar cases
    """
    logger.info(f"Finding similar cases: event_reason={event_reason}")

    try:
        if not knowledge_graph:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Knowledge graph not initialized"
            )

        # Create minimal context
        from pkg.types import AnalysisContext
        context = AnalysisContext(event={"reason": event_reason} if event_reason else None)

        cases = knowledge_graph.find_similar_cases(context, limit=limit)

        return {"cases": [c.dict() for c in cases]}

    except Exception as e:
        logger.error(f"Failed to find similar cases: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


# Learning metrics endpoint
@app.get("/api/v1/metrics/accuracy")
async def get_accuracy_metrics(root_cause_type: Optional[str] = None):
    """
    Get accuracy metrics

    Args:
        root_cause_type: Optional filter by root cause type

    Returns:
        Accuracy metrics
    """
    logger.info(f"Getting accuracy metrics: root_cause_type={root_cause_type}")

    try:
        if not learning_system:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Learning system not initialized"
            )

        metrics = learning_system.get_accuracy_metrics(root_cause_type)
        return metrics

    except Exception as e:
        logger.error(f"Failed to get metrics: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


@app.get("/api/v1/metrics/suggestions")
async def get_improvement_suggestions():
    """
    Get improvement suggestions based on learning

    Returns:
        List of suggestions
    """
    logger.info("Getting improvement suggestions")

    try:
        if not learning_system:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Learning system not initialized"
            )

        suggestions = learning_system.suggest_improvements()
        return {"suggestions": suggestions}

    except Exception as e:
        logger.error(f"Failed to get suggestions: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


# Knowledge graph statistics
@app.get("/api/v1/knowledge/stats")
async def get_knowledge_stats():
    """Get knowledge graph statistics"""
    logger.info("Getting knowledge graph statistics")

    try:
        if not knowledge_graph:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="Knowledge graph not initialized"
            )

        stats = knowledge_graph.get_statistics()
        return stats

    except Exception as e:
        logger.error(f"Failed to get stats: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


# Error handlers
@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    """Handle HTTP exceptions"""
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": exc.detail}
    )


@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception):
    """Handle general exceptions"""
    logger.error(f"Unhandled exception: {exc}")
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={"error": "Internal server error"}
    )


def create_app(config: Dict) -> FastAPI:
    """
    Create and configure FastAPI app

    Args:
        config: Service configuration

    Returns:
        Configured FastAPI app
    """
    app.state.config = config
    return app