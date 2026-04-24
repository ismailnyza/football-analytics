#!/usr/bin/env python3
import urllib.request
import sys
from pathlib import Path

LEAGUES = {
    "E0":  "Premier League",
    "SP1": "La Liga",
    "D1":  "Bundesliga",
    "I1":  "Serie A",
    "F1":  "Ligue 1",
}

SEASONS = ["1920", "2021", "2122", "2223", "2324", "2425"]

BASE_URL = "https://www.football-data.co.uk/mmz4281"
OUT_DIR = Path("data/raw/football-data")

def download(league_code, season):
    url = f"{BASE_URL}/{season}/{league_code}.csv"
    out_path = OUT_DIR / f"{league_code}_{season}.csv"
    if out_path.exists():
        print(f"  SKIP {out_path} (exists)")
        return
    try:
        with urllib.request.urlopen(url) as resp:
            data = resp.read()
        out_path.write_bytes(data)
        print(f"  OK   {out_path}")
    except Exception as e:
        print(f"  FAIL {out_path}: {e}")

def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    for code, name in LEAGUES.items():
        print(f"\n{name} ({code}):")
        for season in SEASONS:
            download(code, season)

if __name__ == "__main__":
    main()
