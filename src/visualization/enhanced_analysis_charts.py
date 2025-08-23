"""
Enhanced Analysis Charts for Football Manager Analytics.
Shows Elo-based transfer impact analysis with before/after comparisons.
"""

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
import json
from pathlib import Path
import logging
from typing import Dict, List
import sys
import os

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from config.settings import RESULTS_PATH, REPORTS_DIR

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Set matplotlib style
plt.style.use('seaborn-v0_8')
sns.set_palette("husl")


class EnhancedAnalysisVisualizer:
    """Generates enhanced visualizations for Elo-based transfer impact analysis."""
    
    def __init__(self):
        self.results_path = RESULTS_PATH
        self.reports_path = REPORTS_DIR
        self.transfer_impact_data = None
        
    def load_data(self) -> None:
        """Load transfer impact analysis data."""
        logger.info("Loading data for enhanced visualization...")
        
        # Load transfer impact analysis
        impact_path = self.results_path / "transfer_impact_analysis.json"
        if impact_path.exists():
            with open(impact_path, 'r') as f:
                self.transfer_impact_data = json.load(f)
            logger.info(f"Loaded transfer impact data for {len(self.transfer_impact_data)} clubs")
    
    def create_elo_impact_chart(self) -> None:
        """Create Elo impact visualization."""
        logger.info("Creating Elo impact chart...")
        
        if not self.transfer_impact_data:
            logger.warning("No transfer impact data available")
            return
        
        # Prepare data for visualization
        clubs = []
        squad_improvements = []
        attack_improvements = []
        midfield_improvements = []
        defense_improvements = []
        
        for club_name, data in self.transfer_impact_data.items():
            clubs.append(club_name)
            elo_analysis = data.get('elo_analysis', {})
            squad_improvements.append(elo_analysis.get('squad_elo_improvement', 0))
            attack_improvements.append(elo_analysis.get('attack_elo_improvement', 0))
            midfield_improvements.append(elo_analysis.get('midfield_elo_improvement', 0))
            defense_improvements.append(elo_analysis.get('defense_elo_improvement', 0))
        
        # Create visualization
        fig, axes = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle('Transfer Impact Analysis - Elo Rating Improvements', fontsize=16, fontweight='bold')
        
        # Squad Elo improvement
        axes[0, 0].bar(clubs, squad_improvements, color='skyblue', alpha=0.7)
        axes[0, 0].set_title('Squad Elo Improvement')
        axes[0, 0].set_ylabel('Elo Points')
        axes[0, 0].tick_params(axis='x', rotation=45)
        
        # Position-specific improvements
        x = np.arange(len(clubs))
        width = 0.25
        
        axes[0, 1].bar(x - width, attack_improvements, width, label='Attack', alpha=0.7)
        axes[0, 1].bar(x, midfield_improvements, width, label='Midfield', alpha=0.7)
        axes[0, 1].bar(x + width, defense_improvements, width, label='Defense', alpha=0.7)
        
        axes[0, 1].set_title('Position-Specific Elo Improvements')
        axes[0, 1].set_ylabel('Elo Points')
        axes[0, 1].set_xticks(x)
        axes[0, 1].set_xticklabels(clubs, rotation=45)
        axes[0, 1].legend()
        
        # Squad strength comparison
        current_strengths = []
        projected_strengths = []
        
        for club_name, data in self.transfer_impact_data.items():
            elo_analysis = data.get('elo_analysis', {})
            current_strengths.append(elo_analysis.get('current_squad_strength', 0))
            projected_strengths.append(elo_analysis.get('projected_squad_strength', 0))
        
        x = np.arange(len(clubs))
        width = 0.35
        
        axes[1, 0].bar(x - width/2, current_strengths, width, label='Current', alpha=0.7)
        axes[1, 0].bar(x + width/2, projected_strengths, width, label='Projected', alpha=0.7)
        
        axes[1, 0].set_title('Squad Strength Comparison')
        axes[1, 0].set_ylabel('Squad Strength (0-1)')
        axes[1, 0].set_xticks(x)
        axes[1, 0].set_xticklabels(clubs, rotation=45)
        axes[1, 0].legend()
        
        # Points improvement
        points_gains = []
        for club_name, data in self.transfer_impact_data.items():
            improvement = data.get('improvement', {})
            points_gains.append(improvement.get('points_gain', 0))
        
        colors = ['green' if gain > 0 else 'red' for gain in points_gains]
        axes[1, 1].bar(clubs, points_gains, color=colors, alpha=0.7)
        axes[1, 1].set_title('Projected Points Gain')
        axes[1, 1].set_ylabel('Points')
        axes[1, 1].tick_params(axis='x', rotation=45)
        axes[1, 1].axhline(y=0, color='black', linestyle='-', alpha=0.3)
        
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "elo_impact_analysis.png"
        plt.savefig(chart_path, dpi=300, bbox_inches='tight')
        logger.info(f"Elo impact chart saved to {chart_path}")
        plt.close()
    
    def create_before_after_comparison(self) -> None:
        """Create before/after comparison visualization."""
        logger.info("Creating before/after comparison chart...")
        
        if not self.transfer_impact_data:
            logger.warning("No transfer impact data available")
            return
        
        # Prepare data
        comparison_data = []
        
        for club_name, data in self.transfer_impact_data.items():
            current = data.get('current_performance', {})
            projected = data.get('projected_performance', {})
            improvement = data.get('improvement', {})
            
            comparison_data.append({
                'Club': club_name,
                'Current Position': current.get('position', 0),
                'Projected Position': projected.get('position', 0),
                'Current Points': current.get('points', 0),
                'Projected Points': projected.get('points', 0),
                'Position Change': improvement.get('position_change', 0),
                'Points Gain': improvement.get('points_gain', 0)
            })
        
        df = pd.DataFrame(comparison_data)
        
        # Create visualization
        fig, axes = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle('Before/After Transfer Impact Analysis', fontsize=16, fontweight='bold')
        
        # Position comparison
        x = np.arange(len(df))
        width = 0.35
        
        axes[0, 0].bar(x - width/2, df['Current Position'], width, label='Current', alpha=0.7, color='lightcoral')
        axes[0, 0].bar(x + width/2, df['Projected Position'], width, label='Projected', alpha=0.7, color='lightgreen')
        
        axes[0, 0].set_title('League Position Comparison')
        axes[0, 0].set_ylabel('Position (Lower is Better)')
        axes[0, 0].set_xticks(x)
        axes[0, 0].set_xticklabels(df['Club'], rotation=45)
        axes[0, 0].legend()
        axes[0, 0].invert_yaxis()
        
        # Points comparison
        axes[0, 1].bar(x - width/2, df['Current Points'], width, label='Current', alpha=0.7, color='lightcoral')
        axes[0, 1].bar(x + width/2, df['Projected Points'], width, label='Projected', alpha=0.7, color='lightgreen')
        
        axes[0, 1].set_title('Points Comparison')
        axes[0, 1].set_ylabel('Points')
        axes[0, 1].set_xticks(x)
        axes[0, 1].set_xticklabels(df['Club'], rotation=45)
        axes[0, 1].legend()
        
        # Position change
        colors = ['green' if change > 0 else 'red' for change in df['Position Change']]
        axes[1, 0].bar(df['Club'], df['Position Change'], color=colors, alpha=0.7)
        axes[1, 0].set_title('Position Change (Positive = Improvement)')
        axes[1, 0].set_ylabel('Position Change')
        axes[1, 0].tick_params(axis='x', rotation=45)
        axes[1, 0].axhline(y=0, color='black', linestyle='-', alpha=0.3)
        
        # Points gain
        colors = ['green' if gain > 0 else 'red' for gain in df['Points Gain']]
        axes[1, 1].bar(df['Club'], df['Points Gain'], color=colors, alpha=0.7)
        axes[1, 1].set_title('Projected Points Gain')
        axes[1, 1].set_ylabel('Points')
        axes[1, 1].tick_params(axis='x', rotation=45)
        axes[1, 1].axhline(y=0, color='black', linestyle='-', alpha=0.3)
        
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "before_after_comparison.png"
        plt.savefig(chart_path, dpi=300, bbox_inches='tight')
        logger.info(f"Before/after comparison chart saved to {chart_path}")
        plt.close()
    
    def generate_all_enhanced_visualizations(self) -> None:
        """Generate all enhanced visualizations."""
        logger.info("=" * 60)
        logger.info("GENERATING ENHANCED VISUALIZATIONS")
        logger.info("=" * 60)
        
        try:
            # Load data
            self.load_data()
            
            # Generate visualizations
            self.create_elo_impact_chart()
            self.create_before_after_comparison()
            
            logger.info("Enhanced visualizations completed successfully!")
            
        except Exception as e:
            logger.error(f"Enhanced visualization generation failed: {str(e)}")
            raise


def main():
    """Main function to run enhanced visualization."""
    visualizer = EnhancedAnalysisVisualizer()
    visualizer.generate_all_enhanced_visualizations()
    print("Enhanced visualizations completed!")
    return True


if __name__ == "__main__":
    main() 