"""
Configuration settings for Football Manager Big Data Analytics project.
"""

import os
from pathlib import Path

# Project paths
PROJECT_ROOT = Path(__file__).parent.parent
DATA_DIR = PROJECT_ROOT / "data"
ARCHIVE_DIR = PROJECT_ROOT / "archive"
SRC_DIR = PROJECT_ROOT / "src"
REPORTS_DIR = PROJECT_ROOT / "reports"
NOTEBOOKS_DIR = PROJECT_ROOT / "notebooks"

# Data file paths
FM_DATASET_PATH = ARCHIVE_DIR / "dataset.csv"
CLUB_DATA_PATH = DATA_DIR / "club_data"
MAPPED_DATA_PATH = DATA_DIR / "mapped_data"
RESULTS_PATH = DATA_DIR / "results"

# Create directories if they don't exist
for path in [DATA_DIR, CLUB_DATA_PATH, MAPPED_DATA_PATH, RESULTS_PATH, REPORTS_DIR, NOTEBOOKS_DIR]:
    path.mkdir(parents=True, exist_ok=True)

# Data sources configuration
TRANSFERMARKT_BASE_URL = "https://www.transfermarkt.com"
FBREF_BASE_URL = "https://fbref.com"

# Target clubs for analysis (focus on Premier League clubs)
TARGET_CLUBS = {
    "MANU985": {
        "name": "Manchester United",
        "transfermarkt_id": "985",
        "budget": 100000000,  # £100M
        "league": "Premier League",
        "positions_needed": ["RW", "ST", "CM"],
        "formation": "4-3-3"
    },
    "ARS11": {
        "name": "Arsenal",
        "transfermarkt_id": "11",
        "budget": 80000000,  # £80M
        "league": "Premier League",
        "positions_needed": ["CM", "ST"],
        "formation": "4-3-3"
    },
    "CHELSEA": {
        "name": "Chelsea",
        "transfermarkt_id": "631",
        "budget": 120000000,  # £120M
        "league": "Premier League",
        "positions_needed": ["ST", "CB"],
        "formation": "4-2-3-1"
    },
    "LIV": {
        "name": "Liverpool",
        "transfermarkt_id": "31",
        "budget": 90000000,  # £90M
        "league": "Premier League",
        "positions_needed": ["DM", "CB"],
        "formation": "4-3-3"
    }
}

# Position mappings for FM to standard positions
POSITION_MAPPINGS = {
    "GK": ["Goalkeeper"],
    "CB": ["DefenderCentral"],
    "LB": ["DefenderLeft"],
    "RB": ["DefenderRight"],
    "DM": ["DefensiveMidfielder"],
    "CM": ["MidfielderCentral"],
    "LM": ["MidfielderLeft"],
    "RM": ["MidfielderRight"],
    "AMC": ["AttackingMidCentral"],
    "AML": ["AttackingMidLeft"],
    "AMR": ["AttackingMidRight"],
    "ST": ["Striker"],
    "WBL": ["WingBackLeft"],
    "WBR": ["WingBackRight"]
}

# Elo rating system parameters
ELO_CONFIG = {
    "initial_rating": 1500,
    "k_factor": 32,
    "rating_range": (1000, 2000),
    "position_weights": {
        "GK": 1.0,
        "CB": 1.0,
        "LB": 0.9,
        "RB": 0.9,
        "DM": 1.1,
        "CM": 1.0,
        "LM": 0.95,
        "RM": 0.95,
        "AMC": 1.05,
        "AML": 1.0,
        "AMR": 1.0,
        "ST": 1.1,
        "WBL": 0.9,
        "WBR": 0.9
    }
}

# FM attribute weights for Elo calculation
ATTRIBUTE_WEIGHTS = {
    "technical": {
        "Passing": 0.15,
        "Dribbling": 0.12,
        "Finishing": 0.10,
        "FirstTouch": 0.08,
        "Technique": 0.10,
        "Crossing": 0.08,
        "LongShots": 0.06,
        "Marking": 0.08,
        "Tackling": 0.08,
        "PenaltyTaking": 0.05
    },
    "mental": {
        "Decisions": 0.12,
        "Vision": 0.10,
        "Anticipation": 0.08,
        "Composure": 0.08,
        "Concentration": 0.06,
        "Determination": 0.06,
        "Leadership": 0.04,
        "OffTheBall": 0.08,
        "Positioning": 0.08,
        "Teamwork": 0.06
    },
    "physical": {
        "Acceleration": 0.08,
        "Pace": 0.08,
        "Stamina": 0.06,
        "Strength": 0.06,
        "Agility": 0.04,
        "Balance": 0.04,
        "Jumping": 0.04,
        "NaturalFitness": 0.04
    }
}

# Position-specific attribute importance
POSITION_ATTRIBUTES = {
    "GK": ["Handling", "Reflexes", "OneOnOnes", "CommandOfArea", "Communication"],
    "CB": ["Marking", "Tackling", "Heading", "Strength", "Positioning"],
    "LB": ["Marking", "Tackling", "Crossing", "Pace", "Stamina"],
    "RB": ["Marking", "Tackling", "Crossing", "Pace", "Stamina"],
    "DM": ["Marking", "Tackling", "Passing", "Positioning", "Stamina"],
    "CM": ["Passing", "Vision", "Decisions", "Stamina", "Technique"],
    "LM": ["Dribbling", "Crossing", "Pace", "Stamina", "Technique"],
    "RM": ["Dribbling", "Crossing", "Pace", "Stamina", "Technique"],
    "AMC": ["Passing", "Vision", "Decisions", "Technique", "Finishing"],
    "AML": ["Dribbling", "Finishing", "Pace", "Technique", "OffTheBall"],
    "AMR": ["Dribbling", "Finishing", "Pace", "Technique", "OffTheBall"],
    "ST": ["Finishing", "Heading", "OffTheBall", "Strength", "Composure"],
    "WBL": ["Marking", "Tackling", "Crossing", "Pace", "Stamina"],
    "WBR": ["Marking", "Tackling", "Crossing", "Pace", "Stamina"]
}

# Spark configuration
SPARK_CONFIG = {
    "app_name": "FM_BigData_Analytics",
    "master": "local[*]",
    "memory": "4g",
    "driver_memory": "2g",
    "executor_memory": "2g"
}

# Web scraping configuration
SCRAPING_CONFIG = {
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "delay_between_requests": 2,  # seconds
    "max_retries": 3,
    "timeout": 30
}

# Analysis thresholds
ANALYSIS_THRESHOLDS = {
    "weak_position_threshold": 1400,  # Elo score below this is considered weak
    "top_performer_threshold": 1600,  # Elo score above this is considered top performer
    "budget_utilization": 0.8,  # Use 80% of budget for recommendations
    "min_recommendations": 5,
    "max_recommendations": 20
}

# Visualization settings
VIZ_CONFIG = {
    "style": "seaborn-v0_8",
    "figure_size": (12, 8),
    "dpi": 300,
    "color_palette": "viridis",
    "save_format": "png"
}

# Database configuration (optional)
DATABASE_CONFIG = {
    "host": "localhost",
    "port": 5432,
    "database": "fm_analytics",
    "user": "postgres",
    "password": "password"
}

# AWS configuration (optional)
AWS_CONFIG = {
    "region": "us-east-1",
    "s3_bucket": "fm-analytics-data",
    "emr_cluster": "fm-analytics-cluster"
} 