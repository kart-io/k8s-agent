"""
Type definitions for Aetherius Reasoning Service
"""

from enum import Enum
from typing import Dict, List, Optional, Any
from datetime import datetime
from pydantic import BaseModel, Field


class AnalysisType(str, Enum):
    """Analysis type enumeration"""
    ROOT_CAUSE = "root_cause"
    PREDICTION = "prediction"
    RECOMMENDATION = "recommendation"


class RiskLevel(str, Enum):
    """Risk level enumeration"""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class RootCauseType(str, Enum):
    """Root cause types"""
    OOM_KILLER = "OOMKiller"
    CPU_THROTTLING = "CPUThrottling"
    DISK_PRESSURE = "DiskPressure"
    NETWORK_ERROR = "NetworkError"
    CONFIG_ERROR = "ConfigError"
    IMAGE_PULL_ERROR = "ImagePullError"
    VOLUME_ERROR = "VolumeError"
    DEPENDENCY_ERROR = "DependencyError"
    RESOURCE_LIMIT = "ResourceLimit"
    UNKNOWN = "Unknown"


# Request Models

class AnalysisContext(BaseModel):
    """Analysis context data"""
    event: Optional[Dict[str, Any]] = None
    logs: Optional[str] = None
    metrics: Optional[Dict[str, Any]] = None
    topology: Optional[Dict[str, Any]] = None
    historical_data: Optional[List[Dict[str, Any]]] = None


class AnalysisOptions(BaseModel):
    """Analysis options"""
    timeout: str = "30s"
    min_confidence: float = 0.7
    include_similar_cases: bool = True
    max_recommendations: int = 5


class AnalysisRequest(BaseModel):
    """AI analysis request"""
    request_id: str
    workflow_id: Optional[str] = None
    analysis_type: AnalysisType
    context: AnalysisContext
    options: AnalysisOptions = Field(default_factory=AnalysisOptions)


# Response Models

class RootCause(BaseModel):
    """Root cause analysis result"""
    type: RootCauseType
    description: str
    confidence: float = Field(ge=0.0, le=1.0)
    evidence: List[str] = Field(default_factory=list)


class Recommendation(BaseModel):
    """Remediation recommendation"""
    action: str
    description: str
    confidence: float = Field(ge=0.0, le=1.0)
    risk: RiskLevel
    impact: str
    steps: List[str] = Field(default_factory=list)
    rollback_steps: Optional[List[str]] = None
    estimated_duration: Optional[str] = None
    metadata: Dict[str, Any] = Field(default_factory=dict)


class SimilarCase(BaseModel):
    """Similar historical case"""
    case_id: str
    description: str
    similarity_score: float
    root_cause: str
    solution: str
    outcome: str
    timestamp: datetime


class AnalysisResult(BaseModel):
    """Complete analysis result"""
    root_cause: Optional[RootCause] = None
    recommendations: List[Recommendation] = Field(default_factory=list)
    confidence: float = Field(ge=0.0, le=1.0)
    evidence: List[str] = Field(default_factory=list)
    similar_cases: List[SimilarCase] = Field(default_factory=list)
    metadata: Dict[str, Any] = Field(default_factory=dict)


class AnalysisResponse(BaseModel):
    """Analysis response"""
    request_id: str
    status: str = "completed"
    result: Optional[AnalysisResult] = None
    error: Optional[str] = None
    processing_time: float = 0.0
    timestamp: datetime = Field(default_factory=datetime.now)


# Prediction Models

class PredictionRequest(BaseModel):
    """Failure prediction request"""
    cluster_id: str
    resource_type: str  # pod, node, service
    resource_name: str
    metrics: Dict[str, Any]
    time_window: str = "24h"


class PredictionResult(BaseModel):
    """Prediction result"""
    failure_probability: float = Field(ge=0.0, le=1.0)
    predicted_failure_time: Optional[datetime] = None
    failure_types: List[str] = Field(default_factory=list)
    confidence: float = Field(ge=0.0, le=1.0)
    contributing_factors: List[str] = Field(default_factory=list)


# Knowledge Graph Models

class KnowledgeNode(BaseModel):
    """Knowledge graph node"""
    id: str
    type: str  # failure, cause, solution, resource
    label: str
    properties: Dict[str, Any] = Field(default_factory=dict)


class KnowledgeRelation(BaseModel):
    """Knowledge graph relationship"""
    from_node: str
    to_node: str
    type: str  # causes, resolves, affects, similar_to
    properties: Dict[str, Any] = Field(default_factory=dict)


class CaseStudy(BaseModel):
    """Historical case study"""
    id: str
    title: str
    description: str
    symptoms: List[str]
    root_cause: str
    solution: str
    outcome: str
    cluster_id: str
    timestamp: datetime
    metadata: Dict[str, Any] = Field(default_factory=dict)


# Feedback Models

class FeedbackType(str, Enum):
    """Feedback type"""
    DIAGNOSIS_ACCURACY = "diagnosis_accuracy"
    RECOMMENDATION_USEFULNESS = "recommendation_usefulness"
    PREDICTION_ACCURACY = "prediction_accuracy"


class Feedback(BaseModel):
    """User feedback for learning"""
    feedback_id: str
    request_id: str
    feedback_type: FeedbackType
    rating: int = Field(ge=1, le=5)
    was_helpful: bool
    actual_root_cause: Optional[str] = None
    actual_solution: Optional[str] = None
    comments: Optional[str] = None
    submitted_by: str
    timestamp: datetime = Field(default_factory=datetime.now)


# Configuration Models

class Config(BaseModel):
    """Service configuration"""
    server_host: str = "0.0.0.0"
    server_port: int = 8082
    log_level: str = "INFO"
    neo4j_uri: str = "bolt://localhost:7687"
    neo4j_user: str = "neo4j"
    neo4j_password: str = "password"
    model_path: str = "./models"
    knowledge_base_path: str = "./data/knowledge"
    enable_gpu: bool = False
    max_workers: int = 4