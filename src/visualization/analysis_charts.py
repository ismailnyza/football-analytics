"""
Visualization module for Football Manager Big Data Analytics.
Generates charts, dashboards, and reports for analysis results.
"""

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots
import json
from pathlib import Path
import logging
from typing import Dict, List, Tuple, Optional
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

# Import squad analysis visualizer
from .squad_analysis_charts import SquadAnalysisVisualizer

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Set matplotlib style
plt.style.use(VIZ_CONFIG['style'])
sns.set_palette(VIZ_CONFIG['color_palette'])


class AnalysisVisualizer:
    """Generates visualizations for Football Manager analytics results."""
    
    def __init__(self):
        self.results_path = RESULTS_PATH
        self.reports_path = REPORTS_DIR
        self.elo_data = None
        self.recommendations = None
        self.club_data = None
        
    def load_data(self) -> None:
        """Load analysis results data."""
        logger.info("Loading analysis data for visualization...")
        
        # Load Elo rankings
        elo_csv_path = self.results_path / "elo_rankings.csv"
        if elo_csv_path.exists():
            self.elo_data = pd.read_csv(elo_csv_path)
            logger.info(f"Loaded {len(self.elo_data)} players with Elo ratings")
        
        # Load transfer recommendations
        recs_path = self.results_path / "transfer_recommendations.json"
        if recs_path.exists():
            with open(recs_path, 'r') as f:
                self.recommendations = json.load(f)
            logger.info(f"Loaded transfer recommendations for {len(self.recommendations)} clubs")
        
        # Load Elo summary
        summary_path = self.results_path / "elo_summary.json"
        if summary_path.exists():
            with open(summary_path, 'r') as f:
                self.elo_summary = json.load(f)
            logger.info("Loaded Elo summary statistics")
    
    def create_elo_distribution_chart(self) -> None:
        """Create Elo rating distribution chart."""
        logger.info("Creating Elo distribution chart...")
        
        if self.elo_data is None:
            logger.warning("No Elo data available")
            return
        
        fig, axes = plt.subplots(2, 2, figsize=(15, 12))
        fig.suptitle('Elo Rating Distribution Analysis', fontsize=16, fontweight='bold')
        
        # Overall Elo distribution
        axes[0, 0].hist(self.elo_data['final_elo'].dropna(), bins=50, alpha=0.7, color='skyblue', edgecolor='black')
        axes[0, 0].set_title('Overall Elo Rating Distribution')
        axes[0, 0].set_xlabel('Elo Rating')
        axes[0, 0].set_ylabel('Number of Players')
        axes[0, 0].axvline(self.elo_data['final_elo'].mean(), color='red', linestyle='--', 
                          label=f'Mean: {self.elo_data["final_elo"].mean():.1f}')
        axes[0, 0].legend()
        
        # Position-wise Elo distribution
        position_elo = self.elo_data.groupby('mapped_position')['final_elo'].mean().sort_values(ascending=False)
        axes[0, 1].bar(range(len(position_elo)), position_elo.values, color='lightcoral')
        axes[0, 1].set_title('Average Elo Rating by Position')
        axes[0, 1].set_xlabel('Position')
        axes[0, 1].set_ylabel('Average Elo Rating')
        axes[0, 1].set_xticks(range(len(position_elo)))
        axes[0, 1].set_xticklabels(position_elo.index, rotation=45)
        
        # Age vs Elo scatter plot
        axes[1, 0].scatter(self.elo_data['Age'], self.elo_data['final_elo'], alpha=0.6, color='green')
        axes[1, 0].set_title('Age vs Elo Rating')
        axes[1, 0].set_xlabel('Age')
        axes[1, 0].set_ylabel('Elo Rating')
        
        # Club-wise average Elo (for players with clubs)
        club_elo = self.elo_data[self.elo_data['club_id'].notna()].groupby('club_name')['final_elo'].mean().sort_values(ascending=False)
        axes[1, 1].bar(range(len(club_elo)), club_elo.values, color='gold')
        axes[1, 1].set_title('Average Squad Elo Rating by Club')
        axes[1, 1].set_xlabel('Club')
        axes[1, 1].set_ylabel('Average Elo Rating')
        axes[1, 1].set_xticks(range(len(club_elo)))
        axes[1, 1].set_xticklabels(club_elo.index, rotation=45)
        
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "elo_distribution_analysis.png"
        plt.savefig(chart_path, dpi=VIZ_CONFIG['dpi'], bbox_inches='tight')
        logger.info(f"Saved Elo distribution chart to {chart_path}")
        plt.close()
    
    def create_transfer_recommendations_chart(self) -> None:
        """Create transfer recommendations visualization."""
        logger.info("Creating transfer recommendations chart...")
        
        if not self.recommendations:
            logger.warning("No transfer recommendations available")
            return
        
        # Prepare data for visualization
        all_recommendations = []
        for club_id, club_recs in self.recommendations.items():
            if club_recs and 'recommendations' in club_recs:
                for rec in club_recs['recommendations']:
                    rec['club_id'] = club_id
                    all_recommendations.append(rec)
        
        if not all_recommendations:
            logger.warning("No recommendations to visualize")
            return
        
        recs_df = pd.DataFrame(all_recommendations)
        
        # Create subplots
        fig, axes = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle('Transfer Recommendations Analysis', fontsize=16, fontweight='bold')
        
        # Top recommendations by Elo
        top_recs = recs_df.nlargest(10, 'final_elo')
        axes[0, 0].barh(range(len(top_recs)), top_recs['final_elo'], color='lightblue')
        axes[0, 0].set_yticks(range(len(top_recs)))
        axes[0, 0].set_yticklabels(top_recs['Name'], fontsize=8)
        axes[0, 0].set_title('Top 10 Transfer Targets by Elo Rating')
        axes[0, 0].set_xlabel('Elo Rating')
        
        # Position distribution of recommendations
        pos_counts = recs_df['mapped_position'].value_counts()
        axes[0, 1].pie(pos_counts.values, labels=pos_counts.index, autopct='%1.1f%%', startangle=90)
        axes[0, 1].set_title('Recommendations by Position')
        
        # Club-wise recommendation count
        club_counts = recs_df['club_id'].value_counts()
        axes[1, 0].bar(range(len(club_counts)), club_counts.values, color='lightgreen')
        axes[1, 0].set_title('Number of Recommendations by Club')
        axes[1, 0].set_xlabel('Club')
        axes[1, 0].set_ylabel('Number of Recommendations')
        axes[1, 0].set_xticks(range(len(club_counts)))
        axes[1, 0].set_xticklabels(club_counts.index, rotation=45)
        
        # Age distribution of recommendations
        axes[1, 1].hist(recs_df['Age'].dropna(), bins=15, alpha=0.7, color='orange', edgecolor='black')
        axes[1, 1].set_title('Age Distribution of Recommendations')
        axes[1, 1].set_xlabel('Age')
        axes[1, 1].set_ylabel('Number of Players')
        
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "transfer_recommendations_analysis.png"
        plt.savefig(chart_path, dpi=VIZ_CONFIG['dpi'], bbox_inches='tight')
        logger.info(f"Saved transfer recommendations chart to {chart_path}")
        plt.close()
    
    def create_squad_analysis_chart(self) -> None:
        """Create squad analysis visualization."""
        logger.info("Creating squad analysis chart...")
        
        if self.elo_data is None:
            logger.warning("No Elo data available for squad analysis")
            return
        
        # Filter players with clubs
        club_players = self.elo_data[self.elo_data['club_id'].notna()]
        
        if len(club_players) == 0:
            logger.warning("No club players found")
            return
        
        # Create subplots for each club
        clubs = club_players['club_id'].unique()
        n_clubs = len(clubs)
        
        fig, axes = plt.subplots(2, 2, figsize=(16, 12))
        fig.suptitle('Squad Analysis by Club', fontsize=16, fontweight='bold')
        
        for i, club_id in enumerate(clubs[:4]):  # Limit to 4 clubs
            club_data = club_players[club_players['club_id'] == club_id]
            club_name = club_data['club_name'].iloc[0] if 'club_name' in club_data.columns else club_id
            
            row, col = i // 2, i % 2
            
            # Position-wise average Elo for this club
            pos_elo = club_data.groupby('mapped_position')['final_elo'].mean().sort_values(ascending=False)
            
            if len(pos_elo) > 0:
                axes[row, col].bar(range(len(pos_elo)), pos_elo.values, color='lightcoral')
                axes[row, col].set_title(f'{club_name} - Position Strength')
                axes[row, col].set_xlabel('Position')
                axes[row, col].set_ylabel('Average Elo Rating')
                axes[row, col].set_xticks(range(len(pos_elo)))
                axes[row, col].set_xticklabels(pos_elo.index, rotation=45, fontsize=8)
                
                # Add threshold line
                axes[row, col].axhline(y=1400, color='red', linestyle='--', alpha=0.7, label='Weak Position Threshold')
                axes[row, col].legend()
        
        plt.tight_layout()
        
        # Save chart
        chart_path = self.reports_path / "squad_analysis.png"
        plt.savefig(chart_path, dpi=VIZ_CONFIG['dpi'], bbox_inches='tight')
        logger.info(f"Saved squad analysis chart to {chart_path}")
        plt.close()
    
    def create_interactive_dashboard(self) -> None:
        """Create interactive Plotly dashboard."""
        logger.info("Creating interactive dashboard...")
        
        if self.elo_data is None:
            logger.warning("No Elo data available for dashboard")
            return
        
        # Create subplots
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=('Elo Rating Distribution', 'Top Players by Position', 
                          'Age vs Elo Rating', 'Club Squad Strength'),
            specs=[[{"type": "histogram"}, {"type": "bar"}],
                   [{"type": "scatter"}, {"type": "bar"}]]
        )
        
        # Elo distribution histogram
        fig.add_trace(
            go.Histogram(x=self.elo_data['final_elo'].dropna(), name='Elo Distribution'),
            row=1, col=1
        )
        
        # Top players by position
        top_players = self.elo_data.groupby('mapped_position').apply(
            lambda x: x.nlargest(1, 'final_elo')
        ).reset_index(drop=True)
        
        fig.add_trace(
            go.Bar(x=top_players['mapped_position'], y=top_players['final_elo'], 
                  text=top_players['Name'], name='Top Player per Position'),
            row=1, col=2
        )
        
        # Age vs Elo scatter
        fig.add_trace(
            go.Scatter(x=self.elo_data['Age'], y=self.elo_data['final_elo'], 
                      mode='markers', name='Age vs Elo'),
            row=2, col=1
        )
        
        # Club squad strength
        club_elo = self.elo_data[self.elo_data['club_id'].notna()].groupby('club_name')['final_elo'].mean().sort_values(ascending=False)
        
        fig.add_trace(
            go.Bar(x=club_elo.index, y=club_elo.values, name='Club Squad Strength'),
            row=2, col=2
        )
        
        # Update layout
        fig.update_layout(
            title_text="Football Manager Analytics Dashboard",
            showlegend=False,
            height=800
        )
        
        # Save interactive dashboard
        dashboard_path = self.reports_path / "interactive_dashboard.html"
        fig.write_html(str(dashboard_path))
        logger.info(f"Saved interactive dashboard to {dashboard_path}")
    
    def create_transfer_recommendations_table(self) -> None:
        """Create transfer recommendations table."""
        logger.info("Creating transfer recommendations table...")
        
        if not self.recommendations:
            logger.warning("No transfer recommendations available")
            return
        
        # Create HTML table for each club
        html_content = """
        <html>
        <head>
            <title>Transfer Recommendations Report</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 20px; }
                table { border-collapse: collapse; width: 100%; margin: 20px 0; }
                th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
                th { background-color: #f2f2f2; font-weight: bold; }
                .club-section { margin: 30px 0; }
                .club-title { color: #333; font-size: 24px; margin-bottom: 10px; }
                .budget-info { color: #666; margin-bottom: 15px; }
            </style>
        </head>
        <body>
            <h1>Transfer Recommendations Report</h1>
            <p>Generated on: {timestamp}</p>
        """.format(timestamp=pd.Timestamp.now().strftime("%Y-%m-%d %H:%M:%S"))
        
        for club_id, club_recs in self.recommendations.items():
            if not club_recs or 'recommendations' not in club_recs:
                continue
            
            club_name = TARGET_CLUBS.get(club_id, {}).get('name', club_id)
            budget = club_recs.get('club_budget', 0)
            target_positions = club_recs.get('target_positions', [])
            
            html_content += f"""
            <div class="club-section">
                <h2 class="club-title">{club_name}</h2>
                <p class="budget-info">Budget: £{budget:,} | Target Positions: {', '.join(target_positions)}</p>
                <table>
                    <tr>
                        <th>Rank</th>
                        <th>Player Name</th>
                        <th>Position</th>
                        <th>Elo Rating</th>
                        <th>Age</th>
                        <th>Technical</th>
                        <th>Mental</th>
                        <th>Physical</th>
                        <th>Estimated Value</th>
                        <th>Budget Fit</th>
                    </tr>
            """
            
            for i, rec in enumerate(club_recs['recommendations'][:10], 1):
                html_content += f"""
                    <tr>
                        <td>{i}</td>
                        <td>{rec.get('Name', 'N/A')}</td>
                        <td>{rec.get('mapped_position', 'N/A')}</td>
                        <td>{rec.get('final_elo', 0):.1f}</td>
                        <td>{rec.get('Age', 'N/A')}</td>
                        <td>{rec.get('avg_technical', 0):.1f}</td>
                        <td>{rec.get('avg_mental', 0):.1f}</td>
                        <td>{rec.get('avg_physical', 0):.1f}</td>
                        <td>£{rec.get('estimated_value', 0):,.0f}</td>
                        <td>{'✓' if rec.get('budget_fit', False) else '✗'}</td>
                    </tr>
                """
            
            html_content += "</table></div>"
        
        html_content += "</body></html>"
        
        # Save HTML report
        report_path = self.reports_path / "transfer_recommendations_report.html"
        with open(report_path, 'w', encoding='utf-8') as f:
            f.write(html_content)
        
        logger.info(f"Saved transfer recommendations table to {report_path}")
    
    def create_summary_report(self) -> None:
        """Create comprehensive summary report."""
        logger.info("Creating comprehensive summary report...")
        
        if self.elo_data is None:
            logger.warning("No data available for summary report")
            return
        
        # Calculate summary statistics
        total_players = len(self.elo_data)
        players_with_clubs = len(self.elo_data[self.elo_data['club_id'].notna()])
        avg_elo = self.elo_data['final_elo'].mean()
        top_player = self.elo_data.loc[self.elo_data['final_elo'].idxmax()]
        
        # Position statistics
        position_stats = self.elo_data.groupby('mapped_position').agg({
            'final_elo': ['count', 'mean', 'std'],
            'Age': 'mean'
        }).round(2)
        
        # Club statistics
        club_stats = self.elo_data[self.elo_data['club_id'].notna()].groupby('club_name').agg({
            'final_elo': ['count', 'mean'],
            'Age': 'mean'
        }).round(2)
        
        # Create summary report
        report_content = f"""
# Football Manager Big Data Analytics - Summary Report

## Executive Summary
- **Total Players Analyzed**: {total_players:,}
- **Players with Club Affiliations**: {players_with_clubs:,} ({players_with_clubs/total_players*100:.1f}%)
- **Average Elo Rating**: {avg_elo:.1f}
- **Top Rated Player**: {top_player['Name']} ({top_player['mapped_position']}) - Elo: {top_player['final_elo']:.1f}

## Position Analysis
{position_stats.to_string()}

## Club Squad Analysis
{club_stats.to_string()}

## Transfer Recommendations Summary
"""
        
        if self.recommendations:
            for club_id, club_recs in self.recommendations.items():
                if club_recs:
                    club_name = TARGET_CLUBS.get(club_id, {}).get('name', club_id)
                    rec_count = len(club_recs.get('recommendations', []))
                    report_content += f"- **{club_name}**: {rec_count} recommendations\n"
        
        report_content += f"""
## Analysis Methodology
- **Elo Rating System**: Position-specific ratings based on FM attributes
- **Data Sources**: Football Manager dataset + Club roster data
- **Processing**: Apache Spark for scalable Big Data analytics
- **Recommendations**: Budget-constrained, position-specific transfer targets

Generated on: {pd.Timestamp.now().strftime("%Y-%m-%d %H:%M:%S")}
"""
        
        # Save summary report
        summary_path = self.reports_path / "summary_report.md"
        with open(summary_path, 'w', encoding='utf-8') as f:
            f.write(report_content)
        
        logger.info(f"Saved summary report to {summary_path}")
    
    def generate_all_visualizations(self) -> None:
        """Generate all visualizations and reports."""
        logger.info("Generating all visualizations and reports...")
        
        # Load data
        self.load_data()
        
        # Create all charts and reports
        self.create_elo_distribution_chart()
        self.create_transfer_recommendations_chart()
        self.create_squad_analysis_chart()
        self.create_interactive_dashboard()
        self.create_transfer_recommendations_table()
        self.create_summary_report()
        
        # Generate data-driven squad analysis visualizations
        logger.info("Generating data-driven squad analysis visualizations...")
        squad_visualizer = SquadAnalysisVisualizer()
        squad_visualizer.generate_all_visualizations()
        
        logger.info("All visualizations and reports generated successfully!")


def main():
    """Main function to generate all visualizations."""
    logger.info("Starting visualization generation...")
    
    # Initialize visualizer
    visualizer = AnalysisVisualizer()
    
    # Generate all visualizations
    visualizer.generate_all_visualizations()
    
    logger.info("Visualization generation completed successfully!")


if __name__ == "__main__":
    main() 