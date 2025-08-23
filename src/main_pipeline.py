"""
Main pipeline for Football Manager Big Data Analytics.
Orchestrates the complete analytics process from data collection to visualization.
"""

import logging
import sys
import os
from pathlib import Path
from datetime import datetime
import json

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(__file__)))

from pyspark.sql import SparkSession
from config.settings import SPARK_CONFIG, RESULTS_PATH, REPORTS_DIR

# Import modules
from data_collection.club_data_scraper import ClubDataScraper, BudgetDataCollector
from data_processing.fm_club_mapper import FMClubMapper
from analytics.elo_rating_system import EloRatingSystem
from analytics.squad_weakness_analyzer import SquadWeaknessAnalyzer
from analytics.transfer_impact_analyzer import TransferImpactAnalyzer
from visualization.analysis_charts import AnalysisVisualizer
from visualization.enhanced_analysis_charts import EnhancedAnalysisVisualizer
from utils.transfer_report_generator import generate_enhanced_transfer_report

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('pipeline.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


class FootballManagerPipeline:
    """Main pipeline for Football Manager Big Data Analytics."""
    
    def __init__(self):
        self.spark = None
        self.start_time = datetime.now()
        self.pipeline_results = {}
        
    def initialize_spark(self):
        """Initialize Apache Spark session."""
        logger.info("Initializing Apache Spark session...")
        
        self.spark = SparkSession.builder \
            .appName(SPARK_CONFIG['app_name']) \
            .config("spark.driver.memory", SPARK_CONFIG['driver_memory']) \
            .config("spark.executor.memory", SPARK_CONFIG['executor_memory']) \
            .config("spark.sql.adaptive.enabled", "true") \
            .config("spark.sql.adaptive.coalescePartitions.enabled", "true") \
            .getOrCreate()
        
        logger.info("Spark session initialized successfully")
        logger.info(f"Spark version: {self.spark.version}")
        
        return self.spark
    
    def run_data_collection(self):
        """Run data collection phase."""
        logger.info("=" * 60)
        logger.info("PHASE 1: DATA COLLECTION")
        logger.info("=" * 60)
        
        try:
            # Initialize scrapers
            scraper = ClubDataScraper()
            budget_collector = BudgetDataCollector()
            
            # Scrape club rosters
            logger.info("Collecting club roster data...")
            club_data = scraper.scrape_all_clubs()
            
            # Compile budget data
            logger.info("Compiling budget data...")
            budget_data = budget_collector.compile_budget_data()
            
            # Create summary
            summary_df = scraper.create_club_summary_csv()
            
            self.pipeline_results['data_collection'] = {
                'clubs_collected': len(club_data),
                'total_players': sum(len(data['players']) for data in club_data.values()),
                'budgets_compiled': len(budget_data),
                'status': 'success'
            }
            
            logger.info(f"Data collection completed: {len(club_data)} clubs, "
                       f"{sum(len(data['players']) for data in club_data.values())} players")
            
            return club_data, budget_data
            
        except Exception as e:
            logger.error(f"Data collection failed: {str(e)}")
            self.pipeline_results['data_collection'] = {
                'status': 'failed',
                'error': str(e)
            }
            raise
    
    def run_data_processing(self):
        """Run data processing and mapping phase."""
        logger.info("=" * 60)
        logger.info("PHASE 2: DATA PROCESSING & MAPPING")
        logger.info("=" * 60)
        
        try:
            # Initialize mapper
            mapper = FMClubMapper()
            
            # Run complete mapping pipeline
            logger.info("Starting FM to Club data mapping...")
            mapped_df = mapper.run_complete_mapping()
            
            # Generate quality report
            logger.info("Generating mapping quality report...")
            quality_report = mapper.get_mapping_quality_report()
            
            self.pipeline_results['data_processing'] = {
                'mapping_success_rate': quality_report.get('mapping_quality', {}).get('match_rate', 0),
                'total_players_processed': quality_report.get('mapping_quality', {}).get('total_players', 0),
                'matched_players': quality_report.get('mapping_quality', {}).get('matched_players', 0),
                'status': 'success'
            }
            
            logger.info(f"Data processing completed: "
                       f"{quality_report.get('mapping_quality', {}).get('match_rate', 0):.1f}% success rate")
            
            return mapped_df, quality_report
            
        except Exception as e:
            logger.error(f"Data processing failed: {str(e)}")
            self.pipeline_results['data_processing'] = {
                'status': 'failed',
                'error': str(e)
            }
            raise
    
    def run_analytics(self):
        """Run analytics and Elo rating phase with enhanced transfer impact analysis."""
        logger.info("=" * 60)
        logger.info("PHASE 3: ANALYTICS & ELO RATING")
        logger.info("=" * 60)
        
        try:
            # Initialize Elo system
            elo_system = EloRatingSystem(self.spark)
            
            # Run complete Elo analysis
            logger.info("Starting Elo rating analysis...")
            recommendations = elo_system.run_complete_elo_analysis()
            
            # Initialize squad weakness analyzer
            weakness_analyzer = SquadWeaknessAnalyzer(self.spark)
            
            # Run data-driven squad analysis
            logger.info("Starting data-driven squad weakness analysis...")
            squad_analysis = weakness_analyzer.run_complete_analysis()
            
            # Initialize transfer impact analyzer
            impact_analyzer = TransferImpactAnalyzer()
            
            # Run enhanced transfer impact analysis
            logger.info("Starting enhanced transfer impact analysis...")
            transfer_impact_analysis = {}
            
            for club_name, club_data in recommendations.items():
                if 'recommendations' in club_data and club_data['recommendations']:
                    impact_analysis = impact_analyzer.calculate_transfer_impact(
                        club_name=club_name,
                        transfers=club_data['recommendations']
                    )
                    transfer_impact_analysis[club_name] = impact_analysis
            
            # Save transfer impact analysis
            impact_path = RESULTS_PATH / "transfer_impact_analysis.json"
            with open(impact_path, 'w') as f:
                json.dump(transfer_impact_analysis, f, indent=2)
            
            logger.info(f"Transfer impact analysis saved to {impact_path}")
            
            # Generate enhanced transfer report
            logger.info("Generating enhanced transfer impact report...")
            generate_enhanced_transfer_report(recommendations, transfer_impact_analysis)
            
            self.pipeline_results['analytics'] = {
                'clubs_analyzed': len(recommendations),
                'total_recommendations': sum(
                    len(rec.get('recommendations', [])) for rec in recommendations.values()
                ),
                'squad_analysis_completed': True,
                'data_driven_positions': True,
                'transfer_impact_analysis': True,
                'status': 'success'
            }
            
            logger.info(f"Analytics completed: {len(recommendations)} clubs analyzed, "
                       f"{sum(len(rec.get('recommendations', [])) for rec in recommendations.values())} recommendations generated")
            logger.info("Data-driven squad weakness analysis completed successfully!")
            logger.info("Enhanced transfer impact analysis completed successfully!")
            
            return recommendations, squad_analysis, transfer_impact_analysis
            
        except Exception as e:
            logger.error(f"Analytics failed: {str(e)}")
            self.pipeline_results['analytics'] = {
                'status': 'failed',
                'error': str(e)
            }
            raise
    
    def run_visualization(self):
        """Run visualization and reporting phase with enhanced analysis."""
        logger.info("=" * 60)
        logger.info("PHASE 4: VISUALIZATION & REPORTING")
        logger.info("=" * 60)
        
        try:
            # Initialize visualizers
            visualizer = AnalysisVisualizer()
            enhanced_visualizer = EnhancedAnalysisVisualizer()
            
            # Generate standard visualizations
            logger.info("Generating standard visualizations and reports...")
            visualizer.generate_all_visualizations()
            
            # Generate enhanced visualizations
            logger.info("Generating enhanced Elo-based visualizations...")
            enhanced_visualizer.generate_all_enhanced_visualizations()
            
            # Count generated files
            report_files = list(REPORTS_DIR.glob("*"))
            
            self.pipeline_results['visualization'] = {
                'reports_generated': len(report_files),
                'report_types': [f.suffix for f in report_files],
                'enhanced_elo_analysis': True,
                'before_after_comparison': True,
                'status': 'success'
            }
            
            logger.info(f"Visualization completed: {len(report_files)} reports generated")
            logger.info("Enhanced Elo-based visualizations completed successfully!")
            
            return report_files
            
        except Exception as e:
            logger.error(f"Visualization failed: {str(e)}")
            self.pipeline_results['visualization'] = {
                'status': 'failed',
                'error': str(e)
            }
            raise
    
    def generate_pipeline_report(self):
        """Generate comprehensive pipeline execution report."""
        logger.info("=" * 60)
        logger.info("GENERATING PIPELINE REPORT")
        logger.info("=" * 60)
        
        end_time = datetime.now()
        execution_time = end_time - self.start_time
        
        # Calculate success rate
        successful_phases = sum(1 for phase in self.pipeline_results.values() 
                              if phase.get('status') == 'success')
        total_phases = len(self.pipeline_results)
        success_rate = (successful_phases / total_phases) * 100 if total_phases > 0 else 0
        
        # Create pipeline report
        pipeline_report = {
            "pipeline_execution": {
                "start_time": self.start_time.isoformat(),
                "end_time": end_time.isoformat(),
                "execution_time_seconds": execution_time.total_seconds(),
                "success_rate_percent": success_rate,
                "total_phases": total_phases,
                "successful_phases": successful_phases
            },
            "phase_results": self.pipeline_results,
            "system_info": {
                "spark_version": self.spark.version if self.spark else "Not initialized",
                "python_version": sys.version,
                "platform": sys.platform
            }
        }
        
        # Save pipeline report
        report_path = RESULTS_PATH / "pipeline_execution_report.json"
        with open(report_path, 'w') as f:
            json.dump(pipeline_report, f, indent=2)
        
        # Create human-readable summary
        summary_content = f"""
# Football Manager Big Data Analytics Pipeline Report

## Execution Summary
- **Start Time**: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}
- **End Time**: {end_time.strftime('%Y-%m-%d %H:%M:%S')}
- **Execution Time**: {execution_time.total_seconds():.2f} seconds
- **Success Rate**: {success_rate:.1f}% ({successful_phases}/{total_phases} phases)

## Phase Results
"""
        
        for phase_name, phase_result in self.pipeline_results.items():
            status = phase_result.get('status', 'unknown')
            summary_content += f"\n### {phase_name.replace('_', ' ').title()}\n"
            summary_content += f"- **Status**: {status}\n"
            
            if status == 'success':
                for key, value in phase_result.items():
                    if key != 'status':
                        summary_content += f"- **{key.replace('_', ' ').title()}**: {value}\n"
            elif status == 'failed':
                summary_content += f"- **Error**: {phase_result.get('error', 'Unknown error')}\n"
        
        summary_content += f"""
## System Information
- **Spark Version**: {pipeline_report['system_info']['spark_version']}
- **Python Version**: {pipeline_report['system_info']['python_version']}
- **Platform**: {pipeline_report['system_info']['platform']}

## Generated Outputs
- **Results Directory**: {RESULTS_PATH}
- **Reports Directory**: {REPORTS_DIR}
- **Log File**: pipeline.log

Generated on: {end_time.strftime('%Y-%m-%d %H:%M:%S')}
"""
        
        # Save summary report
        summary_path = REPORTS_DIR / "pipeline_summary_report.md"
        with open(summary_path, 'w') as f:
            f.write(summary_content)
        
        logger.info(f"Pipeline report saved to {report_path}")
        logger.info(f"Pipeline summary saved to {summary_path}")
        
        return pipeline_report
    
    def run_complete_pipeline(self):
        """Run the complete Big Data analytics pipeline."""
        logger.info("=" * 80)
        logger.info("FOOTBALL MANAGER BIG DATA ANALYTICS PIPELINE")
        logger.info("=" * 80)
        logger.info(f"Pipeline started at: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        try:
            # Phase 1: Data Collection
            club_data, budget_data = self.run_data_collection()
            
            # Phase 2: Data Processing
            mapped_df, quality_report = self.run_data_processing()
            
            # Phase 3: Analytics
            recommendations, squad_analysis, transfer_impact_analysis = self.run_analytics()
            
            # Phase 4: Visualization
            report_files = self.run_visualization()
            
            # Generate final report
            pipeline_report = self.generate_pipeline_report()
            
            logger.info("=" * 80)
            logger.info("PIPELINE COMPLETED SUCCESSFULLY!")
            logger.info("=" * 80)
            
            # Print summary
            execution_time = datetime.now() - self.start_time
            logger.info(f"Total execution time: {execution_time.total_seconds():.2f} seconds")
            logger.info(f"Success rate: {pipeline_report['pipeline_execution']['success_rate_percent']:.1f}%")
            logger.info(f"Generated {len(report_files)} reports")
            
            return {
                'club_data': club_data,
                'mapped_data': mapped_df,
                'recommendations': recommendations,
                'squad_analysis': squad_analysis,
                'transfer_impact_analysis': transfer_impact_analysis,
                'pipeline_report': pipeline_report
            }
            
        except Exception as e:
            logger.error(f"Pipeline failed: {str(e)}")
            self.generate_pipeline_report()  # Generate report even if failed
            raise
    
    def cleanup(self):
        """Clean up resources."""
        if self.spark:
            logger.info("Stopping Spark session...")
            self.spark.stop()
            logger.info("Spark session stopped")


def main():
    """Main function to run the complete pipeline."""
    pipeline = FootballManagerPipeline()
    
    try:
        # Initialize Spark
        pipeline.initialize_spark()
        
        # Run complete pipeline
        results = pipeline.run_complete_pipeline()
        
        logger.info("Pipeline execution completed successfully!")
        return results
        
    except Exception as e:
        logger.error(f"Pipeline execution failed: {str(e)}")
        raise
    
    finally:
        # Cleanup
        pipeline.cleanup()


if __name__ == "__main__":
    main() 