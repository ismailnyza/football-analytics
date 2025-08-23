"""
Football Manager to Club data mapping module.
Maps FM player data with current club affiliations using fuzzy matching.
"""

import pandas as pd
import numpy as np
from pyspark.sql import SparkSession
from pyspark.sql.functions import udf, col, when, lit, array_contains, explode
from pyspark.sql.types import StringType, FloatType, StructType, StructField
import json
import logging
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from fuzzywuzzy import fuzz
import sys
import os

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from config.settings import (
    FM_DATASET_PATH,
    CLUB_DATA_PATH,
    MAPPED_DATA_PATH,
    POSITION_MAPPINGS,
    SPARK_CONFIG
)

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class FMClubMapper:
    """Maps Football Manager data with club information."""
    
    def __init__(self):
        self.spark = self._initialize_spark()
        self.club_data = self._load_club_data()
        self.fm_df = None
        self.mapped_df = None
        
    def _initialize_spark(self) -> SparkSession:
        """Initialize Spark session for Big Data processing."""
        logger.info("Initializing Spark session...")
        
        spark = SparkSession.builder \
            .appName(SPARK_CONFIG['app_name']) \
            .config("spark.driver.memory", SPARK_CONFIG['driver_memory']) \
            .config("spark.executor.memory", SPARK_CONFIG['executor_memory']) \
            .getOrCreate()
        
        logger.info("Spark session initialized successfully")
        return spark
    
    def _load_club_data(self) -> Dict:
        """Load club data from JSON files."""
        logger.info("Loading club data...")
        
        club_data = {}
        club_files = list(CLUB_DATA_PATH.glob("*_roster.json"))
        
        for file_path in club_files:
            with open(file_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                club_id = data['club_id']
                club_data[club_id] = data
        
        logger.info(f"Loaded data for {len(club_data)} clubs")
        return club_data
    
    def load_fm_data(self) -> None:
        """Load Football Manager dataset into Spark DataFrame."""
        logger.info("Loading FM dataset...")
        
        # Read CSV file
        self.fm_df = self.spark.read.csv(
            str(FM_DATASET_PATH),
            header=True,
            inferSchema=True
        )
        
        # Clean column names (remove spaces and special characters)
        for col_name in self.fm_df.columns:
            clean_name = col_name.replace(' ', '').replace('-', '')
            self.fm_df = self.fm_df.withColumnRenamed(col_name, clean_name)
        
        logger.info(f"Loaded {self.fm_df.count()} players from FM dataset")
        logger.info(f"Columns: {self.fm_df.columns}")
    
    def create_club_players_df(self) -> None:
        """Create Spark DataFrame from club player data."""
        logger.info("Creating club players DataFrame...")
        
        club_players_data = []
        
        for club_id, club_info in self.club_data.items():
            for player in club_info['players']:
                club_players_data.append({
                    'club_id': club_id,
                    'club_name': club_info['club_name'],
                    'player_name': player['name'],
                    'position': player['position'],
                    'age': player['age'],
                    'nationality': player['nationality'],
                    'market_value': player['market_value'],
                    'budget': club_info['budget'],
                    'formation': club_info['formation'],
                    'positions_needed': club_info['positions_needed']
                })
        
        # Create DataFrame
        schema = StructType([
            StructField("club_id", StringType(), True),
            StructField("club_name", StringType(), True),
            StructField("player_name", StringType(), True),
            StructField("position", StringType(), True),
            StructField("age", StringType(), True),
            StructField("nationality", StringType(), True),
            StructField("market_value", StringType(), True),
            StructField("budget", StringType(), True),
            StructField("formation", StringType(), True),
            StructField("positions_needed", StringType(), True)
        ])
        
        self.club_players_df = self.spark.createDataFrame(club_players_data, schema)
        logger.info(f"Created club players DataFrame with {self.club_players_df.count()} records")
    
    def map_positions(self) -> None:
        """Map FM positions to standard positions."""
        logger.info("Mapping FM positions to standard positions...")
        
        # Create position mapping UDF
        @udf(StringType())
        def map_position(positions_desc, position_columns):
            """Map FM position description to standard position."""
            if not positions_desc:
                return "Unknown"
            
            # Check which positions the player can play
            positions = []
            for pos, fm_positions in POSITION_MAPPINGS.items():
                for fm_pos in fm_positions:
                    if fm_pos in position_columns and position_columns[fm_pos] == 20:
                        positions.append(pos)
            
            if positions:
                return positions[0]  # Return primary position
            else:
                return "Unknown"
        
        # Apply position mapping
        self.fm_df = self.fm_df.withColumn(
            "mapped_position",
            map_position(col("PositionsDesc"), self.fm_df.columns)
        )
        
        logger.info("Position mapping completed")
    
    def fuzzy_match_players(self, threshold: float = 85) -> None:
        """Perform fuzzy matching between FM players and club players."""
        logger.info("Performing fuzzy matching...")
        
        # Create fuzzy matching UDF
        @udf(StringType())
        def fuzzy_match(fm_name, club_players_str):
            """Fuzzy match FM player name with club players."""
            if not fm_name or not club_players_str:
                return None
            
            club_players = json.loads(club_players_str)
            best_match = None
            best_score = 0
            
            for player in club_players:
                score = fuzz.ratio(fm_name.lower(), player['name'].lower())
                if score > best_score and score >= threshold:
                    best_score = score
                    best_match = player['name']
            
            return best_match
        
        # Prepare club players data for matching
        club_players_json = {}
        for club_id, club_info in self.club_data.items():
            club_players_json[club_id] = json.dumps(club_info['players'])
        
        # Broadcast club players data
        club_players_broadcast = self.spark.sparkContext.broadcast(club_players_json)
        
        # Apply fuzzy matching
        self.fm_df = self.fm_df.withColumn(
            "matched_club_player",
            fuzzy_match(col("Name"), lit(json.dumps(club_players_broadcast.value)))
        )
        
        logger.info("Fuzzy matching completed")
    
    def join_fm_club_data(self) -> None:
        """Join FM data with club data."""
        logger.info("Joining FM data with club data...")
        
        # Create club players DataFrame if not exists
        if not hasattr(self, 'club_players_df'):
            self.create_club_players_df()
        
        # Join on matched player names
        self.mapped_df = self.fm_df.join(
            self.club_players_df,
            self.fm_df.matched_club_player == self.club_players_df.player_name,
            "left"
        )
        
        # Add club_id to all FM players (null for unmatched)
        self.mapped_df = self.mapped_df.withColumn(
            "club_id",
            when(col("club_id").isNotNull(), col("club_id")).otherwise(lit(None))
        )
        
        logger.info(f"Mapped DataFrame created with {self.mapped_df.count()} records")
    
    def calculate_player_attributes(self) -> None:
        """Calculate composite player attributes for analysis."""
        logger.info("Calculating player attributes...")
        
        # Technical attributes
        technical_cols = ["Passing", "Dribbling", "Finishing", "FirstTouch", "Technique", 
                         "Crossing", "LongShots", "Marking", "Tackling", "PenaltyTaking"]
        
        # Mental attributes
        mental_cols = ["Decisions", "Vision", "Anticipation", "Composure", "Concentration",
                      "Determination", "Leadership", "OffTheBall", "Positioning", "Teamwork"]
        
        # Physical attributes
        physical_cols = ["Acceleration", "Pace", "Stamina", "Strength", "Agility", 
                        "Balance", "Jumping", "NaturalFitness"]
        
        # Calculate averages
        for attr_type, cols in [("technical", technical_cols), 
                               ("mental", mental_cols), 
                               ("physical", physical_cols)]:
            available_cols = [col for col in cols if col in self.mapped_df.columns]
            if available_cols:
                avg_expr = sum(col(c) for c in available_cols) / len(available_cols)
                self.mapped_df = self.mapped_df.withColumn(f"avg_{attr_type}", avg_expr)
        
        # Calculate overall ability (weighted average)
        self.mapped_df = self.mapped_df.withColumn(
            "overall_ability",
            (col("avg_technical") * 0.4 + col("avg_mental") * 0.4 + col("avg_physical") * 0.2)
        )
        
        logger.info("Player attributes calculated")
    
    def save_mapped_data(self) -> None:
        """Save mapped data to CSV and Parquet formats."""
        logger.info("Saving mapped data...")
        
        # Save as CSV
        csv_path = MAPPED_DATA_PATH / "fm_club_mapped.csv"
        self.mapped_df.toPandas().to_csv(csv_path, index=False)
        logger.info(f"Saved mapped data to {csv_path}")
        
        # Save as Parquet (better for Spark)
        parquet_path = MAPPED_DATA_PATH / "fm_club_mapped.parquet"
        self.mapped_df.write.mode("overwrite").parquet(str(parquet_path))
        logger.info(f"Saved mapped data to {parquet_path}")
        
        # Save summary statistics
        self._save_summary_stats()
    
    def _save_summary_stats(self) -> None:
        """Save mapping summary statistics."""
        logger.info("Generating mapping summary statistics...")
        
        # Count matched vs unmatched players
        total_players = self.mapped_df.count()
        matched_players = self.mapped_df.filter(col("club_id").isNotNull()).count()
        unmatched_players = total_players - matched_players
        
        # Club-wise statistics
        club_stats = self.mapped_df.filter(col("club_id").isNotNull()) \
            .groupBy("club_id", "club_name") \
            .agg(
                {"*": "count", "overall_ability": "avg", "Age": "avg"}
            ) \
            .withColumnRenamed("count(1)", "player_count") \
            .withColumnRenamed("avg(overall_ability)", "avg_ability") \
            .withColumnRenamed("avg(Age)", "avg_age")
        
        # Save statistics
        stats = {
            "mapping_summary": {
                "total_players": total_players,
                "matched_players": matched_players,
                "unmatched_players": unmatched_players,
                "match_rate": round(matched_players / total_players * 100, 2)
            },
            "club_statistics": club_stats.toPandas().to_dict('records')
        }
        
        stats_path = MAPPED_DATA_PATH / "mapping_statistics.json"
        with open(stats_path, 'w') as f:
            json.dump(stats, f, indent=2)
        
        logger.info(f"Saved mapping statistics to {stats_path}")
        logger.info(f"Mapping success rate: {stats['mapping_summary']['match_rate']}%")
    
    def run_complete_mapping(self) -> None:
        """Run the complete mapping pipeline."""
        logger.info("Starting complete FM to Club mapping pipeline...")
        
        # Load FM data
        self.load_fm_data()
        
        # Map positions
        self.map_positions()
        
        # Perform fuzzy matching
        self.fuzzy_match_players()
        
        # Join with club data
        self.join_fm_club_data()
        
        # Calculate attributes
        self.calculate_player_attributes()
        
        # Save results
        self.save_mapped_data()
        
        logger.info("FM to Club mapping pipeline completed successfully!")
        
        return self.mapped_df
    
    def get_mapping_quality_report(self) -> Dict:
        """Generate a quality report for the mapping process."""
        logger.info("Generating mapping quality report...")
        
        if self.mapped_df is None:
            logger.error("No mapped data available. Run mapping pipeline first.")
            return {}
        
        # Basic statistics
        total_players = self.mapped_df.count()
        matched_players = self.mapped_df.filter(col("club_id").isNotNull()).count()
        
        # Position distribution
        position_dist = self.mapped_df.filter(col("club_id").isNotNull()) \
            .groupBy("mapped_position") \
            .count() \
            .orderBy("count", ascending=False)
        
        # Club distribution
        club_dist = self.mapped_df.filter(col("club_id").isNotNull()) \
            .groupBy("club_name") \
            .count() \
            .orderBy("count", ascending=False)
        
        # Ability distribution
        ability_stats = self.mapped_df.filter(col("club_id").isNotNull()) \
            .select("overall_ability") \
            .summary("count", "mean", "stddev", "min", "max")
        
        report = {
            "mapping_quality": {
                "total_players": total_players,
                "matched_players": matched_players,
                "unmatched_players": total_players - matched_players,
                "match_rate": round(matched_players / total_players * 100, 2)
            },
            "position_distribution": position_dist.toPandas().to_dict('records'),
            "club_distribution": club_dist.toPandas().to_dict('records'),
            "ability_statistics": ability_stats.toPandas().to_dict('records')
        }
        
        # Save report
        report_path = MAPPED_DATA_PATH / "mapping_quality_report.json"
        with open(report_path, 'w') as f:
            json.dump(report, f, indent=2)
        
        logger.info(f"Quality report saved to {report_path}")
        return report


def main():
    """Main function to run the mapping process."""
    logger.info("Starting Football Manager to Club mapping process...")
    
    # Initialize mapper
    mapper = FMClubMapper()
    
    # Run complete mapping pipeline
    mapped_df = mapper.run_complete_mapping()
    
    # Generate quality report
    quality_report = mapper.get_mapping_quality_report()
    
    logger.info("Mapping process completed successfully!")
    logger.info(f"Quality report: {quality_report.get('mapping_quality', {})}")
    
    return mapped_df, quality_report


if __name__ == "__main__":
    main() 