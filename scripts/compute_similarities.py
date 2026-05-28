#!/usr/bin/env python3
"""
compute_similarities.py
=======================
Reads movies from PostgreSQL, computes TF-IDF cosine similarity,
and stores top-N similar movies per movie in movie_similarities table.

Usage:
    pip install -r requirements.txt
    python compute_similarities.py
    python compute_similarities.py --env ../../.env --top-n 10
"""

import os
import re
import sys
import argparse

import numpy as np
import psycopg2
import psycopg2.extras
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity


# ── Helpers ────────────────────────────────────────────────────

def load_env(path: str):
    if not os.path.exists(path):
        return
    with open(path, encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith('#'):
                continue
            if '=' in line:
                k, v = line.split('=', 1)
                os.environ.setdefault(k.strip(), v.strip())


def parse_dsn(dsn: str) -> dict:
    """Parse libpq key=value DSN into a dict."""
    return dict(re.findall(r'(\w+)=([^\s]+)', dsn))


def connect(dsn: str):
    p = parse_dsn(dsn)
    return psycopg2.connect(
        host=p.get('host', 'localhost'),
        port=int(p.get('port', 5432)),
        user=p.get('user', 'postgres'),
        password=p.get('password', ''),
        dbname=p.get('dbname', ''),
    )


# ── Main ───────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description='Compute movie similarities')
    parser.add_argument('--env',   default='../.env', help='Path to .env file')
    parser.add_argument('--top-n', type=int, default=10, help='Top-N similar movies to store per movie')
    args = parser.parse_args()

    load_env(args.env)
    dsn = os.environ.get('DB_DSN', '')
    if not dsn:
        print('ERROR: DB_DSN not set in environment or .env file', file=sys.stderr)
        sys.exit(1)

    conn = connect(dsn)
    cur  = conn.cursor(cursor_factory=psycopg2.extras.DictCursor)
    print('Connected to database.')

    # ── Fetch movies with genres and tags ──────────────────────
    cur.execute("""
        SELECT
            m.id,
            m.title,
            COALESCE(m.description, '')                                      AS description,
            COALESCE(STRING_AGG(DISTINCT REPLACE(g.name, ' ', ''), ' '), '') AS genres,
            COALESCE(STRING_AGG(DISTINCT REPLACE(t.name, ' ', ''), ' '), '') AS tags
        FROM movies m
        LEFT JOIN movie_genres mg ON mg.movie_id = m.id
        LEFT JOIN genres       g  ON g.id = mg.genre_id
        LEFT JOIN movie_tags   mt ON mt.movie_id = m.id
        LEFT JOIN tags         t  ON t.id = mt.tag_id
        GROUP BY m.id, m.title, m.description
        ORDER BY m.created_at DESC
    """)
    rows = cur.fetchall()

    if len(rows) < 2:
        print(f'Only {len(rows)} movie(s) — need at least 2 to compute similarities.')
        conn.close()
        return

    print(f'Found {len(rows)} movies. Building feature vectors...')

    ids      = [r['id']    for r in rows]
    features = []
    for r in rows:
        # Weight: genres 3×, tags 2×, description 1×
        text = ' '.join([
            r['genres']      * 3,
            r['tags']        * 2,
            r['description'] * 1,
        ])
        features.append(text.strip() or r['title'])   # fallback to title

    # ── TF-IDF ────────────────────────────────────────────────
    vectorizer  = TfidfVectorizer(max_features=5000, stop_words='english', lowercase=True)
    tfidf       = vectorizer.fit_transform(features)
    print(f'TF-IDF matrix: {tfidf.shape[0]} movies × {tfidf.shape[1]} features')

    # ── Ensure table exists ────────────────────────────────────
    cur.execute("""
        CREATE TABLE IF NOT EXISTS movie_similarities (
            movie_id         UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
            similar_movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
            score            NUMERIC(6,5) NOT NULL,
            rank             INT NOT NULL,
            PRIMARY KEY (movie_id, similar_movie_id)
        );
        CREATE INDEX IF NOT EXISTS idx_movie_sim_lookup
            ON movie_similarities(movie_id, rank);
    """)
    conn.commit()

    cur.execute("TRUNCATE TABLE movie_similarities")
    conn.commit()
    print('Cleared old similarities. Computing new ones...')

    # ── Compute and insert in batches ─────────────────────────
    n         = len(ids)
    BATCH     = 200
    inserted  = 0

    for start in range(0, n, BATCH):
        end         = min(start + BATCH, n)
        sim_scores  = cosine_similarity(tfidf[start:end], tfidf)
        to_insert   = []

        for i, scores in enumerate(sim_scores):
            movie_idx = start + i
            movie_id  = ids[movie_idx]

            top = [j for j in np.argsort(scores)[::-1] if j != movie_idx][:args.top_n]
            for rank, j in enumerate(top, start=1):
                score = float(scores[j])
                if score > 0:
                    to_insert.append((movie_id, ids[j], round(score, 5), rank))

        if to_insert:
            psycopg2.extras.execute_values(
                cur,
                """
                INSERT INTO movie_similarities (movie_id, similar_movie_id, score, rank)
                VALUES %s
                ON CONFLICT (movie_id, similar_movie_id)
                DO UPDATE SET score = EXCLUDED.score, rank = EXCLUDED.rank
                """,
                to_insert,
            )
            conn.commit()
            inserted += len(to_insert)

        print(f'  {end}/{n} movies processed  ({inserted} rows inserted)')

    print(f'\nDone. {inserted} similarity pairs stored for {n} movies.')
    conn.close()


if __name__ == '__main__':
    main()
