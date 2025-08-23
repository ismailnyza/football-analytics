"""
Squad Analysis Visualization - Charts for data-driven squad weakness analysis.
"""

import pandas as pd
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

from config.settings import (
    VIZ_CONFIG,
    RESULTS_PATH,
    REPORTS_DIR,
    TARGET_CLUBS
)

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Set matplotlib style
plt.style.use(VIZ_CONFIG['style'])
sns.set_palette(VIZ_CONFIG['color_palette'])


class SquadAnalysisVisualizer:
    """Generates visualizations for data-driven squad analysis results."""
    
    def __init__(self):
        self.results_path = RESULTS_PATH
        self.reports_path = REPORTS_DIR
        self.squad_analysis = None
        self.priority_recommendations = None
        
    def load_data(self) -> None:
        """Load squad analysis data."""
        logger.info("Loading squad analysis data...")
        
        # Load squad weakness analysis
        analysis_path = self.results_path / "squad_weakness_analysis.json"
        if analysis_path.exists():
            with open(analysis_path, 'r') as f:
                self.squad_analysis = json.load(f)
            logger.info(f"Loaded squad analysis for {len(self.squad_analysis)} clubs")
        
        # Load priority recommendations
        recs_path = self.results_path / "priority_recommendations.json"
        if recs_path.exists():
            with open(recs_path, 'r') as f:
                self.priority_recommendations = json.load(f)
            logger.info(f"Loaded priority recommendations for {len(self.priority_recommendations)} clubs")
    
    def create_squad_heatmap(self) -> None:
        """Create squad strength heatmap visualization."""
        logger.info("Creating squad strength heatmap...")
        
        if not self.squad_analysis:
            logger.warning("No squad analysis data available")
            return
        
        # Prepare data for heatmap
        heatmap_data = []
        for club_id, club_analysis in self.squad_analysis.items():
            if not club_analysis or 'role_analysis' not in club_analysis:
                continue
            
            club_name = TARGET_CLUBS.get(club_id, {}).get('name', club_id)
            for role in club_analysis['role_analysis']:
                relative_strength = role['avg_elo'] / role.get('league_avg_elo', role['avg_elo'])
                heatmap_data.append({
                    'Club': club_name,
                    'Role': role['data_driven_role'],
                    'Relative Strength': relative_strength,
                    'Player Count': role['player_count']
                })
        
        if not heatmap_data:
            logger.warning("No valid heatmap data to visualize")
            return
        
        df = pd.DataFrame(heatmap_data)
        
        # Create pivot table for heatmap
        pivot_data = df.pivot(index='Role', columns='Club', values='Relative Strength')
        
        # Create heatmap
        plt.figure(figsize=(12, 8))
        sns.heatmap(
            pivot_data,
            annot=True,
            fmt='.2f',
            cmap='RdYlGn',
            center=1.0,
            cbar_kws={'label': 'Relative Strength (Club vs League)'},
            linewidths=0.5
        )
        
        plt.title('Data-Driven Squad Strength Analysis', fontsize=16, fontweight='bold')
        plt.xlabel('Club', fontsize=12)
        plt.ylabel('Data-Driven Role', fontsize=12)
        plt.xticks(rotation=45)
        plt.yticks(rotation=0)
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "data_driven_squad_heatmap.png"
        plt.savefig(chart_path, dpi=VIZ_CONFIG['dpi'], bbox_inches='tight')
        logger.info(f"Saved squad heatmap to {chart_path}")
        plt.close()
    
    def create_priority_chart(self) -> None:
        """Create priority-based recommendations chart."""
        logger.info("Creating priority recommendations chart...")
        
        if not self.priority_recommendations:
            logger.warning("No priority recommendations available")
            return
        
        # Prepare data for visualization
        all_priorities = []
        for club_id, club_recs in self.priority_recommendations.items():
            if not club_recs or 'priorities' not in club_recs:
                continue
            
            club_name = TARGET_CLUBS.get(club_id, {}).get('name', club_id)
            for priority in club_recs['priorities']:
                priority['club_name'] = club_name
                all_priorities.append(priority)
        
        if not all_priorities:
            logger.warning("No priority data to visualize")
            return
        
        df = pd.DataFrame(all_priorities)
        
        # Create subplots
        fig, axes = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle('Data-Driven Squad Priority Analysis', fontsize=16, fontweight='bold')
        
        # Top priorities by club
        top_priorities = df.nlargest(10, 'priority_score')
        colors = ['red' if x == 'Critical' else 'orange' if x == 'High' else 'yellow' 
                 for x in top_priorities['gap_severity']]
        axes[0, 0].barh(range(len(top_priorities)), top_priorities['priority_score'], color=colors)
        axes[0, 0].set_yticks(range(len(top_priorities)))
        axes[0, 0].set_yticklabels([f"{row['club_name']} - {row['role']}" for _, row in top_priorities.iterrows()], fontsize=8)
        axes[0, 0].set_title('Top 10 Data-Driven Priorities')
        axes[0, 0].set_xlabel('Priority Score')
        
        # Priority distribution by severity
        severity_counts = df['gap_severity'].value_counts()
        axes[0, 1].pie(severity_counts.values, labels=severity_counts.index, autopct='%1.1f%%', startangle=90)
        axes[0, 1].set_title('Priority Distribution by Severity')
        
        # Role-wise priority analysis
        role_priorities = df.groupby('role')['priority_score'].mean().sort_values(ascending=False)
        axes[1, 0].bar(range(len(role_priorities)), role_priorities.values, color='lightblue')
        axes[1, 0].set_title('Average Priority Score by Role')
        axes[1, 0].set_xlabel('Role')
        axes[1, 0].set_ylabel('Average Priority Score')
        axes[1, 0].set_xticks(range(len(role_priorities)))
        axes[1, 0].set_xticklabels(role_priorities.index, rotation=45, fontsize=8)
        
        # Club-wise priority count
        club_priorities = df.groupby('club_name')['priority_score'].count()
        axes[1, 1].bar(range(len(club_priorities)), club_priorities.values, color='lightgreen')
        axes[1, 1].set_title('Number of Priorities by Club')
        axes[1, 1].set_xlabel('Club')
        axes[1, 1].set_ylabel('Number of Priorities')
        axes[1, 1].set_xticks(range(len(club_priorities)))
        axes[1, 1].set_xticklabels(club_priorities.index, rotation=45)
        
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "data_driven_priority_analysis.png"
        plt.savefig(chart_path, dpi=VIZ_CONFIG['dpi'], bbox_inches='tight')
        logger.info(f"Saved priority chart to {chart_path}")
        plt.close()
    
    def create_summary_report(self) -> None:
        """Create summary report for data-driven analysis."""
        logger.info("Creating data-driven analysis summary report...")
        
        if not self.squad_analysis or not self.priority_recommendations:
            logger.warning("Insufficient data for summary report")
            return
        
        # Calculate summary statistics
        total_priorities = 0
        critical_priorities = 0
        roles_identified = set()
        
        for club_recs in self.priority_recommendations.values():
            if 'priorities' in club_recs:
                total_priorities += len(club_recs['priorities'])
                for priority in club_recs['priorities']:
                    roles_identified.add(priority['role'])
                    if priority['gap_severity'] == 'Critical':
                        critical_priorities += 1
        
        # Create summary report
        report_content = f"""
# Data-Driven Squad Analysis Summary Report

## Analysis Overview
- **Total Priorities Identified**: {total_priorities}
- **Critical Priorities**: {critical_priorities}
- **Unique Roles Identified**: {len(roles_identified)}
- **Clubs Analyzed**: {len(self.squad_analysis)}

## Data-Driven Approach
This analysis uses machine learning clustering to automatically identify:
1. **Player Roles**: K-means clustering on player attributes
2. **Squad Weaknesses**: Performance gaps vs league averages
3. **Priority Scoring**: Multi-factor priority calculation
4. **Transfer Recommendations**: Role-specific targets

## Key Findings
"""
        
        for club_id, club_recs in self.priority_recommendations.items():
            if 'priorities' in club_recs:
                club_name = TARGET_CLUBS.get(club_id, {}).get('name', club_id)
                report_content += f"\n### {club_name}\n"
                report_content += f"- **Total Priorities**: {len(club_recs['priorities'])}\n"
                
                # Top 3 priorities
                top_priorities = sorted(club_recs['priorities'], key=lambda x: x['priority_score'], reverse=True)[:3]
                for i, priority in enumerate(top_priorities, 1):
                    report_content += f"- **Priority {i}**: {priority['role']} (Score: {priority['priority_score']:.1f}, Severity: {priority['gap_severity']})\n"
        
        report_content += f"""
## Methodology
- **Clustering**: K-means on 19 player attributes
- **Role Mapping**: Attribute pattern matching
- **Weakness Detection**: Performance vs league averages
- **Priority Calculation**: Elo gap + depth gap + tactical importance

Generated on: {pd.Timestamp.now().strftime("%Y-%m-%d %H:%M:%S")}
"""
        
        # Save summary report
        summary_path = self.reports_path / "data_driven_analysis_summary.md"
        with open(summary_path, 'w', encoding='utf-8') as f:
            f.write(report_content)
        
        logger.info(f"Saved data-driven analysis summary to {summary_path}")
    
    def generate_all_visualizations(self) -> None:
        """Generate all squad analysis visualizations."""
        logger.info("Generating all squad analysis visualizations...")
        
        # Load data
        self.load_data()
        
        # Create all charts and reports
        self.create_squad_heatmap()
        self.create_priority_chart()
        self.create_summary_report()
        
        logger.info("All squad analysis visualizations generated successfully!")


def main():
    """Main function to generate squad analysis visualizations."""
    logger.info("Starting squad analysis visualization generation...")
    
    # Initialize visualizer
    visualizer = SquadAnalysisVisualizer()
    
    # Generate all visualizations
    visualizer.generate_all_visualizations()
    
    logger.info("Squad analysis visualization generation completed successfully!")


if __name__ == "__main__":
    main() 