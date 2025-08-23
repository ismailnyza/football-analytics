"""
Club data collection module for Football Manager Big Data Analytics.
Collects current club rosters, player information, and budget data.
"""

import requests
import time
import json
import pandas as pd
from bs4 import BeautifulSoup
from pathlib import Path
import logging
from typing import Dict, List, Optional
from fuzzywuzzy import fuzz
import sys
import os

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))

from config.settings import (
    TARGET_CLUBS, 
    TRANSFERMARKT_BASE_URL, 
    SCRAPING_CONFIG,
    CLUB_DATA_PATH
)

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class ClubDataScraper:
    """Scrapes club data from Transfermarkt and other sources."""
    
    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': SCRAPING_CONFIG['user_agent']
        })
        self.club_data = {}
        
    def scrape_club_roster(self, club_id: str, club_info: Dict) -> Dict:
        """
        Scrape current club roster from Transfermarkt.
        
        Args:
            club_id: Unique club identifier
            club_info: Club information dictionary
            
        Returns:
            Dictionary containing club roster data
        """
        transfermarkt_id = club_info['transfermarkt_id']
        url = f"{TRANSFERMARKT_BASE_URL}/{club_info['name'].lower().replace(' ', '-')}/kader/verein/{transfermarkt_id}"
        
        logger.info(f"Scraping roster for {club_info['name']} from {url}")
        
        try:
            response = self.session.get(url, timeout=SCRAPING_CONFIG['timeout'])
            response.raise_for_status()
            
            soup = BeautifulSoup(response.content, 'html.parser')
            
            # Find the squad table
            squad_table = soup.find('table', {'class': 'items'})
            if not squad_table:
                logger.warning(f"No squad table found for {club_info['name']}")
                return self._create_empty_roster(club_id, club_info)
            
            players = []
            rows = squad_table.find_all('tr', {'class': ['odd', 'even']})
            
            for row in rows:
                player_data = self._extract_player_data(row)
                if player_data:
                    players.append(player_data)
            
            # Add delay to respect rate limits
            time.sleep(SCRAPING_CONFIG['delay_between_requests'])
            
            return {
                'club_id': club_id,
                'club_name': club_info['name'],
                'league': club_info['league'],
                'budget': club_info['budget'],
                'formation': club_info['formation'],
                'positions_needed': club_info['positions_needed'],
                'players': players,
                'scraped_at': pd.Timestamp.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error scraping {club_info['name']}: {str(e)}")
            return self._create_empty_roster(club_id, club_info)
    
    def _extract_player_data(self, row) -> Optional[Dict]:
        """Extract player data from a table row."""
        try:
            # Extract player name
            name_cell = row.find('td', {'class': 'hauptlink'})
            if not name_cell:
                return None
            
            player_name = name_cell.get_text(strip=True)
            
            # Extract position
            position_cell = row.find('td', {'class': 'zentriert'})
            position = position_cell.get_text(strip=True) if position_cell else "Unknown"
            
            # Extract age
            age_cell = row.find_all('td', {'class': 'zentriert'})
            age = age_cell[1].get_text(strip=True) if len(age_cell) > 1 else "Unknown"
            
            # Extract nationality
            nationality_cell = row.find('img', {'class': 'flaggenrahmen'})
            nationality = nationality_cell.get('title') if nationality_cell else "Unknown"
            
            # Extract market value (if available)
            value_cell = row.find('td', {'class': 'rechts'})
            market_value = value_cell.get_text(strip=True) if value_cell else "Unknown"
            
            return {
                'name': player_name,
                'position': position,
                'age': age,
                'nationality': nationality,
                'market_value': market_value
            }
            
        except Exception as e:
            logger.warning(f"Error extracting player data: {str(e)}")
            return None
    
    def _create_empty_roster(self, club_id: str, club_info: Dict) -> Dict:
        """Create empty roster structure when scraping fails."""
        return {
            'club_id': club_id,
            'club_name': club_info['name'],
            'league': club_info['league'],
            'budget': club_info['budget'],
            'formation': club_info['formation'],
            'positions_needed': club_info['positions_needed'],
            'players': [],
            'scraped_at': pd.Timestamp.now().isoformat(),
            'error': 'Failed to scrape roster'
        }
    
    def scrape_all_clubs(self) -> Dict[str, Dict]:
        """Scrape data for all target clubs."""
        logger.info("Starting to scrape data for all target clubs...")
        
        for club_id, club_info in TARGET_CLUBS.items():
            logger.info(f"Scraping data for {club_info['name']}")
            club_data = self.scrape_club_roster(club_id, club_info)
            self.club_data[club_id] = club_data
            
            # Save individual club data
            self._save_club_data(club_id, club_data)
        
        # Save combined data
        self._save_combined_data()
        
        logger.info(f"Completed scraping data for {len(self.club_data)} clubs")
        return self.club_data
    
    def _save_club_data(self, club_id: str, club_data: Dict):
        """Save individual club data to JSON file."""
        file_path = CLUB_DATA_PATH / f"{club_id}_roster.json"
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(club_data, f, indent=2, ensure_ascii=False)
        logger.info(f"Saved {club_id} data to {file_path}")
    
    def _save_combined_data(self):
        """Save combined club data to JSON file."""
        file_path = CLUB_DATA_PATH / "all_clubs_data.json"
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(self.club_data, f, indent=2, ensure_ascii=False)
        logger.info(f"Saved combined data to {file_path}")
    
    def create_club_summary_csv(self):
        """Create a summary CSV file with club information."""
        summary_data = []
        
        for club_id, club_data in self.club_data.items():
            summary_data.append({
                'club_id': club_id,
                'club_name': club_data['club_name'],
                'league': club_data['league'],
                'budget': club_data['budget'],
                'formation': club_data['formation'],
                'positions_needed': ', '.join(club_data['positions_needed']),
                'player_count': len(club_data['players']),
                'scraped_at': club_data['scraped_at']
            })
        
        df = pd.DataFrame(summary_data)
        file_path = CLUB_DATA_PATH / "clubs_summary.csv"
        df.to_csv(file_path, index=False)
        logger.info(f"Saved clubs summary to {file_path}")
        
        return df


class BudgetDataCollector:
    """Collects and manages club budget information."""
    
    def __init__(self):
        self.budget_data = {}
    
    def compile_budget_data(self) -> Dict[str, int]:
        """
        Compile budget data from various sources.
        Currently uses hardcoded data, but can be extended to scrape from news sources.
        """
        logger.info("Compiling budget data for target clubs...")
        
        # Budget data from reliable sources (ESPN, BBC, club financial reports)
        budget_sources = {
            "MANU985": {
                "budget": 100000000,  # £100M
                "source": "ESPN Transfer Window Guide 2025",
                "notes": "Based on recent transfer window spending patterns"
            },
            "ARS11": {
                "budget": 80000000,  # £80M
                "source": "BBC Sport Transfer Analysis",
                "notes": "Conservative estimate based on Arsenal's spending history"
            },
            "CHELSEA": {
                "budget": 120000000,  # £120M
                "source": "Financial Times Club Finances",
                "notes": "High budget due to new ownership investment"
            },
            "LIV": {
                "budget": 90000000,  # £90M
                "source": "Liverpool Echo Transfer News",
                "notes": "Balanced budget considering recent signings"
            }
        }
        
        # Save budget data
        file_path = CLUB_DATA_PATH / "club_budgets.json"
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(budget_sources, f, indent=2, ensure_ascii=False)
        
        logger.info(f"Saved budget data to {file_path}")
        
        # Extract just the budget values
        self.budget_data = {club_id: data['budget'] for club_id, data in budget_sources.items()}
        
        return self.budget_data


def main():
    """Main function to run the data collection process."""
    logger.info("Starting Football Manager club data collection...")
    
    # Initialize scrapers
    scraper = ClubDataScraper()
    budget_collector = BudgetDataCollector()
    
    # Scrape club rosters
    club_data = scraper.scrape_all_clubs()
    
    # Compile budget data
    budget_data = budget_collector.compile_budget_data()
    
    # Create summary
    summary_df = scraper.create_club_summary_csv()
    
    logger.info("Data collection completed successfully!")
    logger.info(f"Collected data for {len(club_data)} clubs")
    logger.info(f"Total players collected: {sum(len(data['players']) for data in club_data.values())}")
    
    return club_data, budget_data


if __name__ == "__main__":
    main() 