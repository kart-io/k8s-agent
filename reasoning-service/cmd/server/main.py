"""
Reasoning Service Main Entry Point
"""

import sys
import os
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent.parent.parent
sys.path.insert(0, str(project_root))

import argparse
import signal
from loguru import logger
import uvicorn

from internal.api.server import create_app


def setup_logging(log_level: str = "INFO"):
    """Configure logging"""
    # Remove default logger
    logger.remove()

    # Add console logger with format
    logger.add(
        sys.stderr,
        format="<green>{time:YYYY-MM-DD HH:mm:ss}</green> | <level>{level: <8}</level> | <cyan>{name}</cyan>:<cyan>{function}</cyan>:<cyan>{line}</cyan> - <level>{message}</level>",
        level=log_level,
        colorize=True
    )

    # Add file logger
    logger.add(
        "logs/reasoning-service.log",
        rotation="500 MB",
        retention="10 days",
        level=log_level,
        format="{time:YYYY-MM-DD HH:mm:ss} | {level: <8} | {name}:{function}:{line} - {message}"
    )

    logger.info("Logging configured")


def load_config(config_path: str) -> dict:
    """
    Load configuration from file or environment

    Args:
        config_path: Path to config file

    Returns:
        Configuration dictionary
    """
    config = {
        "server_host": os.getenv("SERVER_HOST", "0.0.0.0"),
        "server_port": int(os.getenv("SERVER_PORT", "8082")),
        "log_level": os.getenv("LOG_LEVEL", "INFO"),
        "neo4j_uri": os.getenv("NEO4J_URI", "bolt://localhost:7687"),
        "neo4j_user": os.getenv("NEO4J_USER", "neo4j"),
        "neo4j_password": os.getenv("NEO4J_PASSWORD", "password"),
        "model_path": os.getenv("MODEL_PATH", "./models"),
        "knowledge_base_path": os.getenv("KNOWLEDGE_BASE_PATH", "./data/knowledge"),
        "enable_gpu": os.getenv("ENABLE_GPU", "false").lower() == "true",
        "max_workers": int(os.getenv("MAX_WORKERS", "4"))
    }

    # Load from YAML file if exists
    if config_path and os.path.exists(config_path):
        try:
            import yaml
            with open(config_path, 'r') as f:
                file_config = yaml.safe_load(f)
                config.update(file_config)
            logger.info(f"Configuration loaded from {config_path}")
        except Exception as e:
            logger.warning(f"Failed to load config file: {e}, using defaults")

    return config


def setup_signal_handlers():
    """Setup graceful shutdown handlers"""
    def signal_handler(sig, frame):
        logger.info(f"Received signal {sig}, shutting down gracefully...")
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)


def main():
    """Main entry point"""
    # Parse arguments
    parser = argparse.ArgumentParser(description="Aetherius Reasoning Service")
    parser.add_argument(
        "--config",
        type=str,
        default="configs/config.yaml",
        help="Path to configuration file"
    )
    parser.add_argument(
        "--host",
        type=str,
        default=None,
        help="Server host (overrides config)"
    )
    parser.add_argument(
        "--port",
        type=int,
        default=None,
        help="Server port (overrides config)"
    )
    parser.add_argument(
        "--log-level",
        type=str,
        default="INFO",
        choices=["DEBUG", "INFO", "WARNING", "ERROR"],
        help="Log level"
    )

    args = parser.parse_args()

    # Setup logging
    setup_logging(args.log_level)

    # Load configuration
    config = load_config(args.config)

    # Override with command-line arguments
    if args.host:
        config["server_host"] = args.host
    if args.port:
        config["server_port"] = args.port
    if args.log_level:
        config["log_level"] = args.log_level

    # Setup signal handlers
    setup_signal_handlers()

    # Create logs directory
    os.makedirs("logs", exist_ok=True)

    # Print startup banner
    logger.info("=" * 60)
    logger.info("Aetherius Reasoning Service")
    logger.info("AI-Powered Root Cause Analysis & Failure Prediction")
    logger.info("=" * 60)
    logger.info(f"Server: {config['server_host']}:{config['server_port']}")
    logger.info(f"Log Level: {config['log_level']}")
    logger.info(f"Neo4j: {config['neo4j_uri']}")
    logger.info(f"GPU Enabled: {config['enable_gpu']}")
    logger.info("=" * 60)

    # Create FastAPI app
    app = create_app(config)

    # Start server
    try:
        uvicorn.run(
            app,
            host=config["server_host"],
            port=config["server_port"],
            log_level=config["log_level"].lower(),
            access_log=True,
            log_config=None  # Use loguru instead
        )
    except Exception as e:
        logger.error(f"Failed to start server: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()