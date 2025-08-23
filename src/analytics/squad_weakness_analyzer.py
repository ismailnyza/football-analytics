"""
Squad Weakness Analyzer - Automated position identification using Big Data analytics.
Identifies squad weaknesses and positions in need of reinforcement through data-driven analysis.
"""

import pandas as pd
import numpy as np
from pyspark.sql import SparkSession
from pyspark.sql.functions import udf, col, when, lit, avg, count, stddev, min as spark_min, max as spark_max, row_number
from pyspark.sql.types import FloatType, StringType, ArrayType
from pyspark.sql.window import Window
import json
import logging
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from sklearn.cluster import KMeans
from sklearn.preprocessing import StandardScaler
import sys
import os

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from config.settings import (
    MAPPED_DATA_PATH,
    RESULTS_PATH,
    POSITION_ATTRIBUTES,
    ANALYSIS_THRESHOLDS
)

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class SquadWeaknessAnalyzer:
    """Automated squad weakness identification using Big Data analytics."""
    
    def __init__(self, spark_session: SparkSession):
        self.spark = spark_session
        self.mapped_df = None
        self.club_analysis = {}
        
    def load_mapped_data(self) -> None:
        """Load mapped FM and club data."""
        logger.info("Loading mapped data for squad analysis...")
        
        # Try to load from Parquet first (faster)
        parquet_path = MAPPED_DATA_PATH / "fm_club_mapped.parquet"
        if parquet_path.exists():
            self.mapped_df = self.spark.read.parquet(str(parquet_path))
        else:
            # Fall back to CSV
            csv_path = MAPPED_DATA_PATH / "fm_club_mapped.csv"
            self.mapped_df = self.spark.read.csv(str(csv_path), header=True, inferSchema=True)
        
        logger.info(f"Loaded {self.mapped_df.count()} players for squad analysis")
    
    def create_player_role_clusters(self) -> None:
        """Create player role clusters using K-means clustering on attributes."""
        logger.info("Creating player role clusters using K-means...")
        
        # Select relevant attributes for clustering
        clustering_attrs = [
            'Passing', 'Dribbling', 'Finishing', 'Crossing', 'LongShots',
            'Marking', 'Tackling', 'Heading', 'Stamina', 'Pace',
            'Strength', 'Agility', 'Balance', 'Jumping', 'Vision',
            'Decisions', 'Positioning', 'OffTheBall', 'Teamwork'
        ]
        
        # Filter available attributes
        available_attrs = [attr for attr in clustering_attrs if attr in self.mapped_df.columns]
        
        if len(available_attrs) < 5:
            logger.warning("Insufficient attributes for clustering")
            return
        
        # Convert to pandas for clustering
        player_attrs = self.mapped_df.select(available_attrs).toPandas()
        
        # Handle missing values
        player_attrs = player_attrs.fillna(player_attrs.mean())
        
        # Standardize features
        scaler = StandardScaler()
        scaled_attrs = scaler.fit_transform(player_attrs)
        
        # Perform K-means clustering
        n_clusters = 12  # Number of player roles
        kmeans = KMeans(n_clusters=n_clusters, random_state=42, n_init=10)
        clusters = kmeans.fit_predict(scaled_attrs)
        
        # Create role mapping
        role_names = [
            'Goalkeeper', 'Center Back', 'Full Back', 'Defensive Midfielder',
            'Central Midfielder', 'Attacking Midfielder', 'Winger', 'Striker',
            'Target Man', 'Playmaker', 'Box-to-Box', 'Defensive Forward'
        ]
        
        # Map clusters to roles based on cluster centers
        cluster_centers = kmeans.cluster_centers_
        role_mapping = self._map_clusters_to_roles(cluster_centers, available_attrs, role_names)
        
        # Add cluster information back to Spark DataFrame
        cluster_rdd = self.spark.sparkContext.parallelize(clusters)
        role_rdd = cluster_rdd.map(lambda x: role_mapping.get(x, 'Unknown'))
        
        # Create temporary DataFrame for joining
        temp_df = self.spark.createDataFrame(
            [(i, clusters[i], role_mapping.get(clusters[i], 'Unknown')) 
             for i in range(len(clusters))],
            ["row_id", "cluster_id", "data_driven_role"]
        )
        
        # Add row numbers to main DataFrame for joining
        window_spec = Window.orderBy("Name")
        self.mapped_df = self.mapped_df.withColumn("row_id", row_number().over(window_spec))
        
        # Join cluster information
        self.mapped_df = self.mapped_df.join(temp_df, "row_id", "left")
        
        logger.info("Player role clustering completed")
    
    def _map_clusters_to_roles(self, cluster_centers: np.ndarray, attributes: List[str], role_names: List[str]) -> Dict[int, str]:
        """Map K-means clusters to football roles based on attribute patterns."""
        
        # Define expected attribute patterns for each role
        role_patterns = {
            'Goalkeeper': {'Handling': 1.0, 'Reflexes': 1.0, 'OneOnOnes': 1.0},
            'Center Back': {'Marking': 1.0, 'Tackling': 1.0, 'Heading': 1.0, 'Strength': 1.0},
            'Full Back': {'Marking': 0.8, 'Tackling': 0.8, 'Crossing': 0.8, 'Pace': 0.8},
            'Defensive Midfielder': {'Marking': 1.0, 'Tackling': 1.0, 'Passing': 0.8, 'Stamina': 0.8},
            'Central Midfielder': {'Passing': 1.0, 'Vision': 0.8, 'Decisions': 0.8, 'Stamina': 0.8},
            'Attacking Midfielder': {'Passing': 1.0, 'Vision': 1.0, 'Dribbling': 0.8, 'Finishing': 0.6},
            'Winger': {'Dribbling': 1.0, 'Crossing': 1.0, 'Pace': 1.0, 'OffTheBall': 0.8},
            'Striker': {'Finishing': 1.0, 'OffTheBall': 1.0, 'Pace': 0.8, 'Dribbling': 0.6},
            'Target Man': {'Heading': 1.0, 'Strength': 1.0, 'Finishing': 0.8, 'OffTheBall': 0.8},
            'Playmaker': {'Passing': 1.0, 'Vision': 1.0, 'Decisions': 1.0, 'Technique': 0.8},
            'Box-to-Box': {'Stamina': 1.0, 'Passing': 0.8, 'Tackling': 0.8, 'OffTheBall': 0.8},
            'Defensive Forward': {'Tackling': 0.8, 'OffTheBall': 1.0, 'Finishing': 0.8, 'Stamina': 0.8}
        }
        
        # Calculate similarity scores for each cluster
        cluster_role_mapping = {}
        
        for cluster_id, center in enumerate(cluster_centers):
            best_role = 'Unknown'
            best_score = -1
            
            for role, pattern in role_patterns.items():
                score = 0
                for attr, weight in pattern.items():
                    if attr in attributes:
                        attr_idx = attributes.index(attr)
                        # Normalize center value and calculate similarity
                        normalized_center = (center[attr_idx] + 3) / 6  # Assuming -3 to 3 range
                        score += weight * normalized_center
                
                if score > best_score:
                    best_score = score
                    best_role = role
            
            cluster_role_mapping[cluster_id] = best_role
        
        return cluster_role_mapping
    
    def analyze_squad_depth(self, club_id: str) -> Dict:
        """Analyze squad depth and identify weak areas."""
        logger.info(f"Analyzing squad depth for club {club_id}...")
        
        # Filter players for the club
        club_players = self.mapped_df.filter(col("club_id") == club_id)
        
        if club_players.count() == 0:
            logger.warning(f"No players found for club {club_id}")
            return {}
        
        # Analyze by data-driven roles
        role_analysis = club_players.groupBy("data_driven_role") \
            .agg(
                avg("final_elo").alias("avg_elo"),
                count("*").alias("player_count"),
                avg("Age").alias("avg_age"),
                stddev("final_elo").alias("elo_std"),
                min("final_elo").alias("min_elo"),
                max("final_elo").alias("max_elo")
            ) \
            .orderBy("avg_elo", ascending=False)
        
        # Calculate league averages for comparison
        league_analysis = self.mapped_df.filter(col("club_id").isNotNull()) \
            .groupBy("data_driven_role") \
            .agg(
                avg("final_elo").alias("league_avg_elo"),
                count("*").alias("league_player_count")
            )
        
        # Join club and league analysis
        comparison = role_analysis.join(league_analysis, "data_driven_role", "left")
        
        # Identify weaknesses
        weaknesses = comparison.filter(
            (col("avg_elo") < col("league_avg_elo") * 0.95) |  # Below 95% of league average
            (col("player_count") < 2)  # Less than 2 players in role
        )
        
        # Calculate squad strength metrics
        squad_stats = club_players.agg(
            avg("final_elo").alias("squad_avg_elo"),
            count("*").alias("total_players"),
            avg("Age").alias("squad_avg_age"),
            stddev("final_elo").alias("squad_elo_std")
        )
        
        # Identify critical gaps (roles with no players)
        all_roles = self.mapped_df.select("data_driven_role").distinct()
        club_roles = club_players.select("data_driven_role").distinct()
        missing_roles = all_roles.subtract(club_roles)
        
        analysis = {
            "club_id": club_id,
            "squad_statistics": squad_stats.toPandas().to_dict('records')[0],
            "role_analysis": comparison.toPandas().to_dict('records'),
            "weaknesses": weaknesses.toPandas().to_dict('records'),
            "missing_roles": missing_roles.toPandas().to_dict('records'),
            "analysis_timestamp": pd.Timestamp.now().isoformat()
        }
        
        return analysis
    
    def calculate_position_importance(self) -> Dict[str, float]:
        """Calculate position importance based on tactical analysis."""
        logger.info("Calculating position importance...")
        
        # Define tactical importance weights (can be adjusted based on formation)
        tactical_importance = {
            'Goalkeeper': 1.0,      # Essential
            'Center Back': 0.9,     # High importance
            'Full Back': 0.7,       # Important for width
            'Defensive Midfielder': 0.8,  # Defensive stability
            'Central Midfielder': 0.9,    # Core position
            'Attacking Midfielder': 0.8,  # Creative outlet
            'Winger': 0.7,          # Width and pace
            'Striker': 0.9,         # Goalscoring
            'Target Man': 0.6,      # Tactical option
            'Playmaker': 0.8,       # Creative hub
            'Box-to-Box': 0.7,      # Energy and workrate
            'Defensive Forward': 0.5  # Specialized role
        }
        
        return tactical_importance
    
    def generate_priority_recommendations(self, club_id: str, max_recommendations: int = 5) -> Dict:
        """Generate priority-based transfer recommendations."""
        logger.info(f"Generating priority recommendations for club {club_id}...")
        
        # Get squad analysis
        squad_analysis = self.analyze_squad_depth(club_id)
        
        if not squad_analysis:
            return {}
        
        # Calculate priority scores for each weakness
        weaknesses = squad_analysis.get('weaknesses', [])
        missing_roles = squad_analysis.get('missing_roles', [])
        
        # Get tactical importance
        tactical_importance = self.calculate_position_importance()
        
        # Create priority list
        priorities = []
        
        # Add missing roles (highest priority)
        for missing in missing_roles:
            role = missing['data_driven_role']
            priorities.append({
                'role': role,
                'priority_score': 10.0,  # Highest priority
                'reason': 'Missing role',
                'current_players': 0,
                'league_avg_elo': 0,
                'gap_severity': 'Critical'
            })
        
        # Add weak roles
        for weakness in weaknesses:
            role = weakness['data_driven_role']
            current_elo = weakness['avg_elo']
            league_avg = weakness.get('league_avg_elo', current_elo)
            player_count = weakness['player_count']
            
            # Calculate priority score
            elo_gap = (league_avg - current_elo) / league_avg if league_avg > 0 else 0
            depth_gap = max(0, 3 - player_count) / 3  # Normalize to 0-1
            tactical_weight = tactical_importance.get(role, 0.5)
            
            priority_score = (elo_gap * 0.5 + depth_gap * 0.3 + tactical_weight * 0.2) * 10
            
            priorities.append({
                'role': role,
                'priority_score': priority_score,
                'reason': 'Weak performance or insufficient depth',
                'current_players': player_count,
                'current_elo': current_elo,
                'league_avg_elo': league_avg,
                'gap_severity': 'High' if priority_score > 7 else 'Medium'
            })
        
        # Sort by priority score
        priorities.sort(key=lambda x: x['priority_score'], reverse=True)
        
        # Generate recommendations for top priorities
        recommendations = []
        for priority in priorities[:max_recommendations]:
            role = priority['role']
            
            # Find available players for this role
            available_players = self.mapped_df.filter(
                (col("data_driven_role") == role) &
                (col("club_id").isNull()) &  # Not currently at a club
                (col("final_elo") > ANALYSIS_THRESHOLDS['weak_position_threshold'])
            ).orderBy(col("final_elo").desc()).limit(3)
            
            role_recommendations = available_players.select(
                "Name", "data_driven_role", "final_elo", "avg_technical", 
                "avg_mental", "avg_physical", "Age", "NationID"
            ).toPandas().to_dict('records')
            
            for rec in role_recommendations:
                rec['priority_score'] = priority['priority_score']
                rec['gap_severity'] = priority['gap_severity']
                rec['reason'] = priority['reason']
                recommendations.append(rec)
        
        return {
            "club_id": club_id,
            "priorities": priorities,
            "recommendations": recommendations,
            "analysis_timestamp": pd.Timestamp.now().isoformat()
        }
    
    def create_squad_heatmap_data(self, club_id: str) -> Dict:
        """Create data for squad strength heatmap visualization."""
        logger.info(f"Creating squad heatmap data for club {club_id}...")
        
        club_players = self.mapped_df.filter(col("club_id") == club_id)
        
        if club_players.count() == 0:
            return {}
        
        # Get role-wise statistics
        role_stats = club_players.groupBy("data_driven_role") \
            .agg(
                avg("final_elo").alias("avg_elo"),
                count("*").alias("player_count"),
                avg("Age").alias("avg_age")
            ).toPandas()
        
        # Calculate league averages for comparison
        league_stats = self.mapped_df.filter(col("club_id").isNotNull()) \
            .groupBy("data_driven_role") \
            .agg(avg("final_elo").alias("league_avg_elo")) \
            .toPandas()
        
        # Merge club and league stats
        heatmap_data = role_stats.merge(league_stats, on="data_driven_role", how="left")
        
        # Calculate relative strength (club vs league)
        heatmap_data['relative_strength'] = heatmap_data['avg_elo'] / heatmap_data['league_avg_elo']
        heatmap_data['strength_category'] = pd.cut(
            heatmap_data['relative_strength'],
            bins=[0, 0.8, 0.95, 1.05, 1.2, float('inf')],
            labels=['Very Weak', 'Weak', 'Average', 'Strong', 'Very Strong']
        )
        
        return heatmap_data.to_dict('records')
    
    def run_complete_analysis(self) -> Dict:
        """Run complete squad weakness analysis."""
        logger.info("Starting complete squad weakness analysis...")
        
        # Load data
        self.load_mapped_data()
        
        # Create player role clusters
        self.create_player_role_clusters()
        
        # Analyze each club
        club_analyses = {}
        priority_recommendations = {}
        
        for club_id in ["MANU985", "ARS11", "CHELSEA", "LIV"]:
            logger.info(f"Analyzing squad for {club_id}...")
            
            # Squad depth analysis
            club_analyses[club_id] = self.analyze_squad_depth(club_id)
            
            # Priority recommendations
            priority_recommendations[club_id] = self.generate_priority_recommendations(club_id)
        
        # Save results
        self._save_analysis_results(club_analyses, priority_recommendations)
        
        logger.info("Squad weakness analysis completed successfully!")
        
        return {
            'club_analyses': club_analyses,
            'priority_recommendations': priority_recommendations
        }
    
    def _save_analysis_results(self, club_analyses: Dict, priority_recommendations: Dict) -> None:
        """Save analysis results to files."""
        logger.info("Saving squad analysis results...")
        
        # Save club analyses
        analyses_path = RESULTS_PATH / "squad_weakness_analysis.json"
        with open(analyses_path, 'w') as f:
            json.dump(club_analyses, f, indent=2)
        
        # Save priority recommendations
        recs_path = RESULTS_PATH / "priority_recommendations.json"
        with open(recs_path, 'w') as f:
            json.dump(priority_recommendations, f, indent=2)
        
        # Save heatmap data
        heatmap_data = {}
        for club_id in club_analyses.keys():
            heatmap_data[club_id] = self.create_squad_heatmap_data(club_id)
        
        heatmap_path = RESULTS_PATH / "squad_heatmap_data.json"
        with open(heatmap_path, 'w') as f:
            json.dump(heatmap_data, f, indent=2)
        
        logger.info(f"Analysis results saved to {RESULTS_PATH}")


def main():
    """Main function to run squad weakness analysis."""
    logger.info("Starting squad weakness analysis...")
    
    # Initialize Spark
    spark = SparkSession.builder \
        .appName("Squad_Weakness_Analysis") \
        .config("spark.driver.memory", "2g") \
        .config("spark.executor.memory", "2g") \
        .getOrCreate()
    
    # Initialize analyzer
    analyzer = SquadWeaknessAnalyzer(spark)
    
    # Run complete analysis
    results = analyzer.run_complete_analysis()
    
    logger.info("Squad weakness analysis completed successfully!")
    
    return results


if __name__ == "__main__":
    main() 