"""
Elo-based rating system for Football Manager player evaluation.
Implements position-specific Elo ratings and squad analysis.
"""

import pandas as pd
import numpy as np
from pyspark.sql import SparkSession
from pyspark.sql.functions import udf, col, when, lit, avg, count, sum as spark_sum, row_number
from pyspark.sql.types import FloatType, StringType, IntegerType
from pyspark.sql.window import Window
import json
import logging
from pathlib import Path
from typing import Dict, List, Tuple, Optional
import sys
import os

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from config.settings import (
    ELO_CONFIG,
    ATTRIBUTE_WEIGHTS,
    POSITION_ATTRIBUTES,
    POSITION_MAPPINGS,
    ANALYSIS_THRESHOLDS,
    MAPPED_DATA_PATH,
    RESULTS_PATH
)

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class EloRatingSystem:
    """Elo-based rating system for football player evaluation."""
    
    def __init__(self, spark_session: SparkSession):
        self.spark = spark_session
        self.mapped_df = None
        self.elo_df = None
        self.position_ratings = {}
        
    def load_mapped_data(self) -> None:
        """Load mapped FM and club data."""
        logger.info("Loading mapped data...")
        
        # Try to load from Parquet first (faster)
        parquet_path = MAPPED_DATA_PATH / "fm_club_mapped.parquet"
        if parquet_path.exists():
            self.mapped_df = self.spark.read.parquet(str(parquet_path))
        else:
            # Fall back to CSV
            csv_path = MAPPED_DATA_PATH / "fm_club_mapped.csv"
            self.mapped_df = self.spark.read.csv(str(csv_path), header=True, inferSchema=True)
        
        logger.info(f"Loaded {self.mapped_df.count()} players")
    
    def calculate_position_specific_elo(self) -> None:
        """Calculate Elo ratings based on position-specific attributes."""
        logger.info("Calculating position-specific Elo ratings...")
        
        # Create position-specific Elo calculation UDF
        @udf(FloatType())
        def calculate_position_elo(position, attributes_dict):
            """Calculate Elo rating for a specific position."""
            if not position or position == "Unknown":
                return ELO_CONFIG['initial_rating']
            
            attributes = json.loads(attributes_dict) if isinstance(attributes_dict, str) else attributes_dict
            
            # Get position-specific attributes
            pos_attrs = POSITION_ATTRIBUTES.get(position, [])
            if not pos_attrs:
                return ELO_CONFIG['initial_rating']
            
            # Calculate weighted score
            total_score = 0
            total_weight = 0
            
            for attr in pos_attrs:
                if attr in attributes and attributes[attr] is not None:
                    # Normalize attribute (0-20 scale to 0-1)
                    normalized_attr = attributes[attr] / 20.0
                    total_score += normalized_attr
                    total_weight += 1
            
            if total_weight == 0:
                return ELO_CONFIG['initial_rating']
            
            # Calculate average score
            avg_score = total_score / total_weight
            
            # Convert to Elo scale (1000-2000)
            elo_score = ELO_CONFIG['initial_rating'] + (avg_score * 1000)
            
            # Apply position weight
            position_weight = ELO_CONFIG['position_weights'].get(position, 1.0)
            elo_score = elo_score * position_weight
            
            # Ensure within bounds
            elo_score = max(ELO_CONFIG['rating_range'][0], 
                           min(ELO_CONFIG['rating_range'][1], elo_score))
            
            return float(elo_score)
        
        # Prepare attributes dictionary for each player
        attribute_cols = list(ATTRIBUTE_WEIGHTS['technical'].keys()) + \
                        list(ATTRIBUTE_WEIGHTS['mental'].keys()) + \
                        list(ATTRIBUTE_WEIGHTS['physical'].keys())
        
        # Create attributes dictionary
        attributes_expr = "{" + ", ".join([f'"{col}": {col}' for col in attribute_cols if col in self.mapped_df.columns]) + "}"
        
        self.mapped_df = self.mapped_df.withColumn(
            "attributes_dict",
            attributes_expr
        )
        
        # Calculate position-specific Elo
        self.mapped_df = self.mapped_df.withColumn(
            "position_elo",
            calculate_position_elo(col("mapped_position"), col("attributes_dict"))
        )
        
        logger.info("Position-specific Elo ratings calculated")
    
    def calculate_composite_elo(self) -> None:
        """Calculate composite Elo rating using weighted attributes."""
        logger.info("Calculating composite Elo ratings...")
        
        # Create composite Elo calculation UDF
        @udf(FloatType())
        def calculate_composite_elo(technical_avg, mental_avg, physical_avg, age, consistency):
            """Calculate composite Elo rating."""
            if not all([technical_avg, mental_avg, physical_avg]):
                return ELO_CONFIG['initial_rating']
            
            # Weighted average of attributes
            composite_score = (technical_avg * 0.4 + mental_avg * 0.4 + physical_avg * 0.2)
            
            # Age factor (peak performance around 25-28)
            age_factor = 1.0
            if age:
                age = float(age)
                if 25 <= age <= 28:
                    age_factor = 1.1  # Peak performance
                elif age < 22:
                    age_factor = 0.9  # Young player
                elif age > 32:
                    age_factor = 0.85  # Aging player
            
            # Consistency factor
            consistency_factor = 1.0
            if consistency:
                consistency = float(consistency)
                consistency_factor = consistency / 20.0  # Normalize to 0-1
            
            # Calculate final Elo
            elo_score = ELO_CONFIG['initial_rating'] + (composite_score * 1000)
            elo_score = elo_score * age_factor * consistency_factor
            
            # Ensure within bounds
            elo_score = max(ELO_CONFIG['rating_range'][0], 
                           min(ELO_CONFIG['rating_range'][1], elo_score))
            
            return float(elo_score)
        
        # Calculate composite Elo
        self.mapped_df = self.mapped_df.withColumn(
            "composite_elo",
            calculate_composite_elo(
                col("avg_technical"),
                col("avg_mental"),
                col("avg_physical"),
                col("Age"),
                col("Consistency")
            )
        )
        
        # Calculate final Elo (average of position and composite)
        self.mapped_df = self.mapped_df.withColumn(
            "final_elo",
            (col("position_elo") + col("composite_elo")) / 2
        )
        
        logger.info("Composite Elo ratings calculated")
    
    def analyze_squad_strengths(self, club_id: str) -> Dict:
        """Analyze squad strengths and weaknesses for a specific club."""
        logger.info(f"Analyzing squad strengths for club {club_id}...")
        
        # Filter players for the club
        club_players = self.mapped_df.filter(col("club_id") == club_id)
        
        if club_players.count() == 0:
            logger.warning(f"No players found for club {club_id}")
            return {}
        
        # Position-wise analysis
        position_analysis = club_players.groupBy("mapped_position") \
            .agg(
                avg("final_elo").alias("avg_elo"),
                count("*").alias("player_count"),
                avg("Age").alias("avg_age")
            ) \
            .orderBy("avg_elo", ascending=False)
        
        # Overall squad statistics
        squad_stats = club_players.agg(
            avg("final_elo").alias("squad_avg_elo"),
            count("*").alias("total_players"),
            avg("Age").alias("squad_avg_age")
        )
        
        # Identify weak positions
        weak_positions = position_analysis.filter(
            col("avg_elo") < ANALYSIS_THRESHOLDS['weak_position_threshold']
        )
        
        # Identify strong positions
        strong_positions = position_analysis.filter(
            col("avg_elo") > ANALYSIS_THRESHOLDS['top_performer_threshold']
        )
        
        # Compile analysis results
        analysis = {
            "club_id": club_id,
            "squad_statistics": squad_stats.toPandas().to_dict('records')[0],
            "position_analysis": position_analysis.toPandas().to_dict('records'),
            "weak_positions": weak_positions.toPandas().to_dict('records'),
            "strong_positions": strong_positions.toPandas().to_dict('records'),
            "analysis_timestamp": pd.Timestamp.now().isoformat()
        }
        
        return analysis
    
    def generate_transfer_recommendations(self, club_id: str, max_recommendations: int = 10) -> Dict:
        """Generate transfer recommendations for a specific club."""
        logger.info(f"Generating transfer recommendations for club {club_id}...")
        
        # Get club information
        club_info = self.mapped_df.filter(col("club_id") == club_id).first()
        if not club_info:
            logger.warning(f"No club information found for {club_id}")
            return {}
        
        club_budget = float(club_info['budget']) if club_info['budget'] else 0
        positions_needed = club_info['positions_needed'].split(', ') if club_info['positions_needed'] else []
        
        # Get squad analysis
        squad_analysis = self.analyze_squad_strengths(club_id)
        weak_positions = [pos['mapped_position'] for pos in squad_analysis.get('weak_positions', [])]
        
        # Combine positions needed with weak positions
        target_positions = list(set(positions_needed + weak_positions))
        
        recommendations = []
        
        for position in target_positions:
            # Find available players for this position
            available_players = self.mapped_df.filter(
                (col("mapped_position") == position) &
                (col("club_id").isNull()) &  # Not currently at a club
                (col("final_elo") > ANALYSIS_THRESHOLDS['weak_position_threshold'])
            )
            
            if available_players.count() > 0:
                # Rank by Elo score
                window_spec = Window.partitionBy("mapped_position").orderBy(col("final_elo").desc())
                
                ranked_players = available_players.withColumn(
                    "position_rank",
                    row_number().over(window_spec)
                ).filter(col("position_rank") <= max_recommendations)
                
                # Add to recommendations
                position_recommendations = ranked_players.select(
                    "Name", "mapped_position", "final_elo", "avg_technical", 
                    "avg_mental", "avg_physical", "Age", "NationID"
                ).toPandas().to_dict('records')
                
                recommendations.extend(position_recommendations)
        
        # Sort by Elo score and limit total recommendations
        recommendations.sort(key=lambda x: x['final_elo'], reverse=True)
        recommendations = recommendations[:max_recommendations]
        
        # Calculate estimated transfer values (simplified)
        for rec in recommendations:
            # Simple transfer value estimation based on Elo and age
            base_value = rec['final_elo'] * 1000  # Base value
            age_factor = 1.0
            if rec['Age']:
                age = float(rec['Age'])
                if 22 <= age <= 28:
                    age_factor = 1.2  # Prime age
                elif age < 22:
                    age_factor = 0.8  # Young player
                else:
                    age_factor = 0.6  # Older player
            
            rec['estimated_value'] = base_value * age_factor
            rec['budget_fit'] = rec['estimated_value'] <= club_budget * ANALYSIS_THRESHOLDS['budget_utilization']
        
        return {
            "club_id": club_id,
            "club_budget": club_budget,
            "target_positions": target_positions,
            "recommendations": recommendations,
            "recommendation_count": len(recommendations),
            "generated_at": pd.Timestamp.now().isoformat()
        }
    
    def create_elo_rankings(self) -> None:
        """Create global Elo rankings for all players."""
        logger.info("Creating global Elo rankings...")
        
        # Create rankings by position
        window_spec = Window.partitionBy("mapped_position").orderBy(col("final_elo").desc())
        
        self.elo_df = self.mapped_df.withColumn(
            "position_rank",
            row_number().over(window_spec)
        )
        
        # Create overall rankings
        overall_window = Window.orderBy(col("final_elo").desc())
        self.elo_df = self.elo_df.withColumn(
            "overall_rank",
            row_number().over(overall_window)
        )
        
        logger.info("Elo rankings created")
    
    def save_elo_results(self) -> None:
        """Save Elo analysis results."""
        logger.info("Saving Elo analysis results...")
        
        # Save ranked players
        csv_path = RESULTS_PATH / "elo_rankings.csv"
        self.elo_df.toPandas().to_csv(csv_path, index=False)
        logger.info(f"Saved Elo rankings to {csv_path}")
        
        # Save Parquet version
        parquet_path = RESULTS_PATH / "elo_rankings.parquet"
        self.elo_df.write.mode("overwrite").parquet(str(parquet_path))
        logger.info(f"Saved Elo rankings to {parquet_path}")
        
        # Generate and save summary statistics
        self._save_elo_summary()
    
    def _save_elo_summary(self) -> None:
        """Save Elo analysis summary statistics."""
        logger.info("Generating Elo summary statistics...")
        
        # Position-wise statistics
        position_stats = self.elo_df.groupBy("mapped_position") \
            .agg(
                avg("final_elo").alias("avg_elo"),
                count("*").alias("player_count"),
                avg("Age").alias("avg_age")
            ) \
            .orderBy("avg_elo", ascending=False)
        
        # Club-wise statistics
        club_stats = self.elo_df.filter(col("club_id").isNotNull()) \
            .groupBy("club_id", "club_name") \
            .agg(
                avg("final_elo").alias("squad_avg_elo"),
                count("*").alias("player_count")
            ) \
            .orderBy("squad_avg_elo", ascending=False)
        
        # Top players by position
        top_players = self.elo_df.filter(col("position_rank") <= 10) \
            .select("Name", "mapped_position", "final_elo", "club_name", "position_rank") \
            .orderBy("mapped_position", "position_rank")
        
        summary = {
            "position_statistics": position_stats.toPandas().to_dict('records'),
            "club_statistics": club_stats.toPandas().to_dict('records'),
            "top_players_by_position": top_players.toPandas().to_dict('records'),
            "analysis_timestamp": pd.Timestamp.now().isoformat()
        }
        
        # Save summary
        summary_path = RESULTS_PATH / "elo_summary.json"
        with open(summary_path, 'w') as f:
            json.dump(summary, f, indent=2)
        
        logger.info(f"Saved Elo summary to {summary_path}")
    
    def run_complete_elo_analysis(self) -> Dict:
        """Run complete Elo analysis pipeline."""
        logger.info("Starting complete Elo analysis pipeline...")
        
        # Load data
        self.load_mapped_data()
        
        # Calculate Elo ratings
        self.calculate_position_specific_elo()
        self.calculate_composite_elo()
        
        # Create rankings
        self.create_elo_rankings()
        
        # Save results
        self.save_elo_results()
        
        # Generate transfer recommendations for all clubs
        recommendations = {}
        for club_id in ["MANU985", "ARS11", "CHELSEA", "LIV"]:
            recommendations[club_id] = self.generate_transfer_recommendations(club_id)
        
        # Save recommendations
        recs_path = RESULTS_PATH / "transfer_recommendations.json"
        with open(recs_path, 'w') as f:
            json.dump(recommendations, f, indent=2)
        
        logger.info(f"Saved transfer recommendations to {recs_path}")
        logger.info("Elo analysis pipeline completed successfully!")
        
        return recommendations


def main():
    """Main function to run the Elo analysis."""
    logger.info("Starting Elo rating system analysis...")
    
    # Initialize Spark
    spark = SparkSession.builder \
        .appName("FM_Elo_Analysis") \
        .config("spark.driver.memory", "2g") \
        .config("spark.executor.memory", "2g") \
        .getOrCreate()
    
    # Initialize Elo system
    elo_system = EloRatingSystem(spark)
    
    # Run complete analysis
    recommendations = elo_system.run_complete_elo_analysis()
    
    logger.info("Elo analysis completed successfully!")
    
    return recommendations


if __name__ == "__main__":
    main() 