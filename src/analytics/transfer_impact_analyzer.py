"""
Transfer Impact Analyzer for Football Manager Analytics.
Analyzes the impact of transfers on team performance and Premier League position.
"""

import pandas as pd
import numpy as np
import json
import logging
from pathlib import Path
from typing import Dict, List, Tuple, Optional
import sys
import os

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from config.settings import RESULTS_PATH, DATA_DIR

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TransferImpactAnalyzer:
    """Analyzes transfer impact on team performance and Premier League position."""
    
    def __init__(self):
        self.results_path = RESULTS_PATH
        self.data_path = DATA_DIR
        self.premier_league_data = None
        self.transfer_recommendations = None
        self.load_data()  # Load data on initialization
        
    def load_data(self) -> None:
        """Load Premier League data and transfer recommendations."""
        logger.info("Loading data for transfer impact analysis...")
        
        # Load Premier League table
        pl_table_path = self.data_path / "premier_league" / "premier_league_table_2023_24.csv"
        if pl_table_path.exists():
            self.premier_league_data = pd.read_csv(pl_table_path)
            logger.info(f"Loaded Premier League data for {len(self.premier_league_data)} clubs")
        else:
            logger.warning(f"Premier League data not found at {pl_table_path}")
            # Try alternative path
            alt_path = Path("data/premier_league/premier_league_table_2023_24.csv")
            if alt_path.exists():
                self.premier_league_data = pd.read_csv(alt_path)
                logger.info(f"Loaded Premier League data from alternative path: {alt_path}")
                logger.info(f"Loaded data for {len(self.premier_league_data)} clubs")
            else:
                logger.error("Premier League data not found in any location")
        
        # Load transfer recommendations
        recs_path = self.results_path / "transfer_recommendations.json"
        if recs_path.exists():
            with open(recs_path, 'r') as f:
                self.transfer_recommendations = json.load(f)
            logger.info(f"Loaded transfer recommendations for {len(self.transfer_recommendations)} clubs")
    
    def calculate_transfer_impact(self, club_name: str, transfers: List[Dict]) -> Dict:
        """Calculate the impact of transfers on a club's performance based on Elo ratings."""
        logger.info(f"Calculating transfer impact for {club_name}...")
        
        if self.premier_league_data is None:
            logger.warning("No Premier League data available")
            return {}
        
        # Get current club performance
        club_data = self.premier_league_data[self.premier_league_data['club'] == club_name]
        if club_data.empty:
            logger.warning(f"No data found for {club_name}")
            return {}
        
        current_performance = club_data.iloc[0].to_dict()
        
        # Calculate Elo-based transfer impact
        squad_elo_improvement = 0
        position_elo_improvements = {
            'attack': 0,
            'midfield': 0,
            'defense': 0,
            'goalkeeper': 0
        }
        
        for transfer in transfers:
            player_elo = transfer.get('elo_rating', 1500)
            position = transfer.get('position', 'Unknown')
            
            # Calculate Elo improvement based on position
            if position in ['ST', 'CF', 'RW', 'LW', 'AM']:
                position_elo_improvements['attack'] += (player_elo - 1500) * 0.8
                squad_elo_improvement += (player_elo - 1500) * 0.8
            elif position in ['CM', 'CDM', 'CAM']:
                position_elo_improvements['midfield'] += (player_elo - 1500) * 0.7
                squad_elo_improvement += (player_elo - 1500) * 0.7
            elif position in ['CB', 'LB', 'RB', 'WB']:
                position_elo_improvements['defense'] += (player_elo - 1500) * 0.6
                squad_elo_improvement += (player_elo - 1500) * 0.6
            elif position == 'GK':
                position_elo_improvements['goalkeeper'] += (player_elo - 1500) * 0.5
                squad_elo_improvement += (player_elo - 1500) * 0.5
        
        # Calculate projected performance based on Elo improvement
        # Use Elo to estimate points improvement
        elo_points_ratio = 0.4  # Points gained per Elo point improvement
        points_boost = squad_elo_improvement * elo_points_ratio
        projected_points = current_performance['points'] + points_boost
        
        # Calculate projected position
        projected_position = self._calculate_projected_position(club_name, projected_points)
        
        # Calculate squad strength improvement
        current_squad_strength = self._estimate_squad_strength(current_performance)
        projected_squad_strength = current_squad_strength + (squad_elo_improvement / 100)
        
        return {
            "club": club_name,
            "current_performance": current_performance,
            "transfers": transfers,
            "elo_analysis": {
                "squad_elo_improvement": round(squad_elo_improvement, 1),
                "attack_elo_improvement": round(position_elo_improvements['attack'], 1),
                "midfield_elo_improvement": round(position_elo_improvements['midfield'], 1),
                "defense_elo_improvement": round(position_elo_improvements['defense'], 1),
                "goalkeeper_elo_improvement": round(position_elo_improvements['goalkeeper'], 1),
                "current_squad_strength": round(current_squad_strength, 2),
                "projected_squad_strength": round(projected_squad_strength, 2)
            },
            "projected_performance": {
                "position": projected_position,
                "points": round(projected_points, 1),
                "points_improvement": round(points_boost, 1)
            },
            "improvement": {
                "position_change": current_performance['position'] - projected_position,
                "points_gain": round(points_boost, 1),
                "squad_strength_gain": round(projected_squad_strength - current_squad_strength, 2)
            }
        }
    
    def _calculate_projected_position(self, club_name: str, projected_points: float) -> int:
        """Calculate projected position based on points."""
        if self.premier_league_data is None:
            return 0
        
        table_copy = self.premier_league_data.copy()
        club_mask = table_copy['club'] == club_name
        table_copy.loc[club_mask, 'points'] = projected_points
        table_copy = table_copy.sort_values('points', ascending=False).reset_index(drop=True)
        table_copy['position'] = range(1, len(table_copy) + 1)
        new_position = table_copy[table_copy['club'] == club_name]['position'].iloc[0]
        return int(new_position)
    
    def _estimate_squad_strength(self, performance: Dict) -> float:
        """Estimate squad strength based on Premier League performance."""
        # Base squad strength from league position and points
        base_strength = (21 - performance['position']) * 0.05  # Higher position = higher strength
        
        # Adjust based on points
        points_factor = performance['points'] / 100  # Normalize to 0-1 scale
        
        # Adjust based on goal difference
        goal_diff_factor = max(0, performance['goal_difference'] + 50) / 100  # Normalize
        
        # Calculate overall strength
        squad_strength = base_strength + (points_factor * 0.6) + (goal_diff_factor * 0.4)
        
        return min(1.0, max(0.0, squad_strength))  # Clamp between 0 and 1
    
    def generate_transfer_impact_report(self) -> Dict:
        """Generate comprehensive transfer impact report."""
        logger.info("Generating transfer impact report...")
        
        if self.transfer_recommendations is None:
            logger.warning("No transfer recommendations available")
            return {}
        
        impact_analysis = {}
        
        for club_name, recommendations in self.transfer_recommendations.items():
            if 'recommendations' in recommendations:
                transfers = recommendations['recommendations']
                impact = self.calculate_transfer_impact(club_name, transfers)
                if impact:
                    impact_analysis[club_name] = impact
        
        # Save impact analysis
        impact_path = self.results_path / "transfer_impact_analysis.json"
        with open(impact_path, 'w') as f:
            json.dump(impact_analysis, f, indent=2)
        
        logger.info(f"Transfer impact analysis saved to {impact_path}")
        
        return impact_analysis
    
    def run_complete_analysis(self) -> Dict:
        """Run complete transfer impact analysis."""
        logger.info("=" * 60)
        logger.info("TRANSFER IMPACT ANALYSIS")
        logger.info("=" * 60)
        
        try:
            # Load data
            self.load_data()
            
            # Generate impact analysis
            impact_analysis = self.generate_transfer_impact_report()
            
            logger.info("Transfer impact analysis completed successfully!")
            
            return {
                "impact_analysis": impact_analysis
            }
            
        except Exception as e:
            logger.error(f"Transfer impact analysis failed: {str(e)}")
            raise


def main():
    """Main function to run transfer impact analysis."""
    analyzer = TransferImpactAnalyzer()
    results = analyzer.run_complete_analysis()
    print("Transfer impact analysis completed!")
    return results


if __name__ == "__main__":
    main() 