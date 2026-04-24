#!/usr/bin/env python3
import csv
import json
import math
from datetime import datetime, timezone
from pathlib import Path

DATA_GLOB_DIR = Path('data/raw/football-data')
OUT_PATH = Path('docs/validation/iteration-002-draw-search.json')


def load_matches():
    matches = []
    for path in sorted(DATA_GLOB_DIR.glob('E0_*.csv')):
        season = path.stem.split('_')[-1]
        with path.open(encoding='utf-8-sig', newline='') as f:
            reader = csv.DictReader(f)
            for row in reader:
                matches.append(
                    {
                        'season': season,
                        'date': datetime.strptime(row['Date'], '%d/%m/%Y'),
                        'home': row['HomeTeam'].strip(),
                        'away': row['AwayTeam'].strip(),
                        'result': row['FTR'].strip(),
                    }
                )
    matches.sort(key=lambda m: (m['date'], m['season'], m['home'], m['away']))
    return matches


def actual_score(result):
    return {'H': 1.0, 'D': 0.5, 'A': 0.0}[result]


def classify(probs):
    return max(probs, key=probs.get)


def run_davidson(matches, cfg, ratings=None):
    ratings = dict(ratings or {})
    correct = 0
    pred = {'H': 0, 'D': 0, 'A': 0}
    for m in matches:
        hr = ratings.get(m['home'], 1500.0)
        ar = ratings.get(m['away'], 1500.0)
        hs = 10 ** ((hr + cfg['home_advantage']) / 400.0)
        aw = 10 ** (ar / 400.0)
        dr = cfg['draw_factor'] * (hs * aw) ** 0.5
        denom = hs + aw + dr
        probs = {'H': hs / denom, 'D': dr / denom, 'A': aw / denom}
        label = classify(probs)
        pred[label] += 1
        correct += label == m['result']
        expected = probs['H'] + 0.5 * probs['D']
        delta = cfg['k_factor'] * (actual_score(m['result']) - expected)
        ratings[m['home']] = hr + delta
        ratings[m['away']] = ar - delta
    n = len(matches)
    return {
        'accuracy': correct / n,
        'predicted_draw_rate': pred['D'] / n,
        'predicted_distribution': {k: v / n for k, v in pred.items()},
    }, ratings


def run_draw_decay(matches, cfg, ratings=None):
    ratings = dict(ratings or {})
    correct = 0
    pred = {'H': 0, 'D': 0, 'A': 0}
    for m in matches:
        hr = ratings.get(m['home'], 1500.0)
        ar = ratings.get(m['away'], 1500.0)
        diff = (hr + cfg['home_advantage']) - ar
        p_draw = cfg['base_draw'] * math.exp(-abs(diff) / cfg['draw_scale'])
        p_draw = max(0.01, min(0.6, p_draw))
        p_home_no_draw = 1.0 / (1.0 + 10 ** (-diff / 400.0))
        probs = {
            'D': p_draw,
            'H': (1.0 - p_draw) * p_home_no_draw,
            'A': (1.0 - p_draw) * (1.0 - p_home_no_draw),
        }
        label = classify(probs)
        pred[label] += 1
        correct += label == m['result']
        expected = probs['H'] + 0.5 * probs['D']
        delta = cfg['k_factor'] * (actual_score(m['result']) - expected)
        ratings[m['home']] = hr + delta
        ratings[m['away']] = ar - delta
    n = len(matches)
    return {
        'accuracy': correct / n,
        'predicted_draw_rate': pred['D'] / n,
        'predicted_distribution': {k: v / n for k, v in pred.items()},
    }, ratings


def main():
    matches = load_matches()
    pretrain = [m for m in matches if m['season'] in {'1920', '2021', '2122', '2223'}]
    validation = [m for m in matches if m['season'] == '2324']
    holdout = [m for m in matches if m['season'] == '2425']

    davidson_rows = []
    for k in [12, 16, 20, 24, 28, 32]:
        for ha in [40, 50, 60, 70, 80, 90]:
            for draw in [0.4, 0.6, 0.8, 1.0, 1.2, 1.5, 2.0, 3.0]:
                cfg = {'k_factor': k, 'home_advantage': ha, 'draw_factor': draw}
                _, ratings = run_davidson(pretrain, cfg)
                val_metrics, ratings = run_davidson(validation, cfg, ratings)
                hold_metrics, _ = run_davidson(holdout, cfg, ratings)
                davidson_rows.append({'family': 'davidson', 'config': cfg, 'validation': val_metrics, 'holdout': hold_metrics})

    decay_rows = []
    for k in [12, 16, 20, 24, 28, 32]:
        for ha in [40, 50, 60, 70, 80, 90]:
            for base_draw in [0.15, 0.20, 0.25, 0.30, 0.35, 0.40]:
                for draw_scale in [50, 75, 100, 125, 150, 200, 250]:
                    cfg = {
                        'k_factor': k,
                        'home_advantage': ha,
                        'base_draw': base_draw,
                        'draw_scale': draw_scale,
                    }
                    _, ratings = run_draw_decay(pretrain, cfg)
                    val_metrics, ratings = run_draw_decay(validation, cfg, ratings)
                    hold_metrics, _ = run_draw_decay(holdout, cfg, ratings)
                    decay_rows.append({'family': 'draw_decay', 'config': cfg, 'validation': val_metrics, 'holdout': hold_metrics})

    artifact = {
        'generated_at': datetime.now(timezone.utc).isoformat(),
        'validation_season': '2324',
        'holdout_season': '2425',
        'top_validation_davidson': sorted(
            davidson_rows,
            key=lambda row: (row['validation']['accuracy'], row['holdout']['accuracy']),
            reverse=True,
        )[:5],
        'top_validation_draw_decay': sorted(
            decay_rows,
            key=lambda row: (row['validation']['accuracy'], row['holdout']['accuracy']),
            reverse=True,
        )[:5],
        'best_holdout_nonzero_draw': sorted(
            [row for row in davidson_rows + decay_rows if row['holdout']['predicted_draw_rate'] > 0],
            key=lambda row: (row['holdout']['accuracy'], row['holdout']['predicted_draw_rate']),
            reverse=True,
        )[:5],
    }

    OUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUT_PATH.write_text(json.dumps(artifact, indent=2) + '\n', encoding='utf-8')
    print(f'wrote {OUT_PATH}')
    print('best davidson validation:', artifact['top_validation_davidson'][0])
    print('best draw_decay validation:', artifact['top_validation_draw_decay'][0])
    print('best nonzero-draw holdout:', artifact['best_holdout_nonzero_draw'][0] if artifact['best_holdout_nonzero_draw'] else 'none')


if __name__ == '__main__':
    main()
