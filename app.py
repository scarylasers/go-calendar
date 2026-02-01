"""
Game Over Pop1 War Team Calendar - Flask Backend
For Render deployment with PostgreSQL
"""

from flask import Flask, jsonify, request, send_from_directory
import os
from datetime import datetime
import random
import string
import requests
import json

# Try to import psycopg2, fall back to JSON if not available (local dev)
try:
    import psycopg2
    from psycopg2.extras import RealDictCursor
    HAS_POSTGRES = True
except ImportError:
    HAS_POSTGRES = False

app = Flask(__name__, static_folder='.', static_url_path='')

# Database URL from environment (Render provides this)
DATABASE_URL = os.environ.get('DATABASE_URL')

# ==================== DATABASE HELPERS ====================

def get_db_connection():
    """Get PostgreSQL connection"""
    if not DATABASE_URL or not HAS_POSTGRES:
        return None
    conn = psycopg2.connect(DATABASE_URL)
    return conn

def init_db():
    """Initialize database tables"""
    conn = get_db_connection()
    if not conn:
        return

    cur = conn.cursor()

    # Games table
    cur.execute('''
        CREATE TABLE IF NOT EXISTS games (
            id TEXT PRIMARY KEY,
            date TEXT NOT NULL,
            time TEXT NOT NULL,
            opponent TEXT NOT NULL,
            notes TEXT DEFAULT '',
            available TEXT DEFAULT '[]',
            unavailable TEXT DEFAULT '[]',
            roster TEXT DEFAULT '[]',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')

    # Player preferences table
    cur.execute('''
        CREATE TABLE IF NOT EXISTS player_preferences (
            player_id TEXT PRIMARY KEY,
            preference TEXT DEFAULT 'starter'
        )
    ''')

    # Settings table (for webhook, etc.)
    cur.execute('''
        CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT
        )
    ''')

    conn.commit()
    cur.close()
    conn.close()

# Initialize on startup
if DATABASE_URL and HAS_POSTGRES:
    try:
        init_db()
        print("Database initialized successfully")
    except Exception as e:
        print(f"Database init error: {e}")

# ==================== DATA FUNCTIONS ====================

def get_all_games():
    """Get all games from database"""
    conn = get_db_connection()
    if not conn:
        return []

    cur = conn.cursor(cursor_factory=RealDictCursor)
    cur.execute('SELECT * FROM games ORDER BY date, time')
    rows = cur.fetchall()
    cur.close()
    conn.close()

    games = []
    for row in rows:
        game = dict(row)
        game['available'] = json.loads(game['available'] or '[]')
        game['unavailable'] = json.loads(game['unavailable'] or '[]')
        game['roster'] = json.loads(game['roster'] or '[]')
        games.append(game)

    return games

def get_game_by_id(game_id):
    """Get a single game by ID"""
    conn = get_db_connection()
    if not conn:
        return None

    cur = conn.cursor(cursor_factory=RealDictCursor)
    cur.execute('SELECT * FROM games WHERE id = %s', (game_id,))
    row = cur.fetchone()
    cur.close()
    conn.close()

    if row:
        game = dict(row)
        game['available'] = json.loads(game['available'] or '[]')
        game['unavailable'] = json.loads(game['unavailable'] or '[]')
        game['roster'] = json.loads(game['roster'] or '[]')
        return game
    return None

def create_game_db(game_data):
    """Create a new game"""
    conn = get_db_connection()
    if not conn:
        return None

    game_id = 'game_' + str(int(datetime.now().timestamp() * 1000)) + '_' + ''.join(random.choices(string.ascii_lowercase + string.digits, k=9))

    cur = conn.cursor()
    cur.execute('''
        INSERT INTO games (id, date, time, opponent, notes, available, unavailable, roster)
        VALUES (%s, %s, %s, %s, %s, '[]', '[]', '[]')
    ''', (game_id, game_data['date'], game_data['time'], game_data['opponent'], game_data.get('notes', '')))

    conn.commit()
    cur.close()
    conn.close()

    return get_game_by_id(game_id)

def update_game_db(game_id, updates):
    """Update a game"""
    conn = get_db_connection()
    if not conn:
        return None

    cur = conn.cursor()

    for key, value in updates.items():
        if key in ['available', 'unavailable', 'roster']:
            value = json.dumps(value)
        cur.execute(f'UPDATE games SET {key} = %s WHERE id = %s', (value, game_id))

    conn.commit()
    cur.close()
    conn.close()

    return get_game_by_id(game_id)

def delete_game_db(game_id):
    """Delete a game"""
    conn = get_db_connection()
    if not conn:
        return False

    cur = conn.cursor()
    cur.execute('DELETE FROM games WHERE id = %s', (game_id,))
    conn.commit()
    cur.close()
    conn.close()
    return True

def get_all_preferences():
    """Get all player preferences"""
    conn = get_db_connection()
    if not conn:
        return {}

    cur = conn.cursor(cursor_factory=RealDictCursor)
    cur.execute('SELECT * FROM player_preferences')
    rows = cur.fetchall()
    cur.close()
    conn.close()

    return {row['player_id']: row['preference'] for row in rows}

def set_preference_db(player_id, preference):
    """Set a player's preference"""
    conn = get_db_connection()
    if not conn:
        return None

    cur = conn.cursor()
    cur.execute('''
        INSERT INTO player_preferences (player_id, preference)
        VALUES (%s, %s)
        ON CONFLICT (player_id) DO UPDATE SET preference = %s
    ''', (player_id, preference, preference))

    conn.commit()
    cur.close()
    conn.close()

    return {'player_id': player_id, 'preference': preference}

def get_setting(key):
    """Get a setting value"""
    conn = get_db_connection()
    if not conn:
        return None

    cur = conn.cursor()
    cur.execute('SELECT value FROM settings WHERE key = %s', (key,))
    row = cur.fetchone()
    cur.close()
    conn.close()

    return row[0] if row else None

def set_setting(key, value):
    """Set a setting value"""
    conn = get_db_connection()
    if not conn:
        return

    cur = conn.cursor()
    cur.execute('''
        INSERT INTO settings (key, value)
        VALUES (%s, %s)
        ON CONFLICT (key) DO UPDATE SET value = %s
    ''', (key, value, value))

    conn.commit()
    cur.close()
    conn.close()

# ==================== MEMBER DATA ====================

ACTIVE_MEMBERS = [
    {'id': 'alock', 'name': 'GO-Alock4', 'year': 2021},
    {'id': 'cronides', 'name': 'GO-Cronides', 'year': 2021, 'note': 'Former Head of GO'},
    {'id': 'ghostxrp', 'name': 'GO_GhostXRP', 'year': 2021},
    {'id': 'bramrianne', 'name': 'GO_ BramRianne', 'year': 2022, 'region': 'EU'},
    {'id': 'thesean', 'name': 'GO_The.Sean', 'year': 2023, 'region': 'EU', 'note': 'Former Head of GO Europe'},
    {'id': 'babs', 'name': 'GO_BABs', 'year': 2023},
    {'id': 'honeyluvv', 'name': 'GO_HoneyLuvv', 'year': 2023},
    {'id': 'zoloto', 'name': 'GO_Zoloto', 'year': 2023, 'region': 'EU'},
    {'id': 'humanoid', 'name': 'GO_Humanoid', 'year': 2023, 'region': 'EU', 'note': 'Head of GO Europe'},
    {'id': 'jesshawk', 'name': 'GO_JessHawk3', 'year': 2023, 'note': 'Head of GO'},
    {'id': 'deathraider', 'name': 'GO_Deathraider255', 'year': 2024, 'region': 'EU'},
    {'id': 'pinkpwnage', 'name': 'GOxPinkPWNAGE5', 'year': 2024},
    {'id': 'sami', 'name': 'Sami_008', 'year': 2024},
    {'id': 'silverlining', 'name': 'GO_SilverLining23', 'year': 2024, 'region': 'EU'},
    {'id': 'colinthe5', 'name': 'Go_Colinthe5', 'year': 2024, 'note': 'Head of GO League'},
    {'id': 'thedon', 'name': 'GO_The_Don', 'year': 2024},
    {'id': 'headhelper', 'name': 'GO HeadHelper', 'year': 2024},
    {'id': 'cosmo', 'name': 'GO_Cosmo', 'year': 2025},
    {'id': 'fxyz', 'name': 'f(x,y,z)', 'year': 2025},
    {'id': 'amberloaf', 'name': 'AmberLoaf', 'year': 2025},
    {'id': 'scarylasers', 'name': 'GO_ScaryLasers', 'year': 2025},
    {'id': 'glotones', 'name': 'GLOTONES', 'year': 2025},
    {'id': 'smich', 'name': 'GO_SMICH1989', 'year': 2025, 'region': 'EU'},
    {'id': 'flexx', 'name': 'GO_FlexX', 'year': 2025},
    {'id': 'kygelli', 'name': 'GO_KyGelli', 'year': 2025},
    {'id': 'chr1sp', 'name': 'GO_Chr1sP', 'year': 2025}
]

SUB_MEMBERS = [
    {'id': 'shark', 'name': 'Shark', 'year': 2021},
    {'id': 'docbutler', 'name': 'GO_DocButler', 'year': 2021},
    {'id': 'drsmartazz', 'name': 'GO_DrSmartAzz', 'year': 2021},
    {'id': 'pbandc', 'name': 'PBandC-GO', 'year': 2021},
    {'id': 'maverick', 'name': 'GO_Maverick', 'year': 2021},
    {'id': 'honeygun', 'name': 'HoneyGUN', 'year': 2022},
    {'id': 'loki', 'name': 'GO_Loki714', 'year': 2022},
    {'id': 'lizlow', 'name': 'LizLow91', 'year': 2023},
    {'id': 'kc', 'name': 'GO_KC', 'year': 2023},
    {'id': 'lester', 'name': 'GO_Lester', 'year': 2024},
    {'id': 'kingslayer', 'name': 'GO_Kingslayer', 'year': 2024, 'region': 'EU'},
    {'id': 'bacon', 'name': 'AllTheBaconAndEGGz', 'year': 2025},
    {'id': 'stooobe', 'name': 'GO_STOOOBE', 'year': 2025, 'note': 'Former Head of GO'},
    {'id': 'johnharple', 'name': 'GO_JohnHarple', 'year': 2025}
]

ALL_MEMBERS = ACTIVE_MEMBERS + SUB_MEMBERS

def get_member_name(member_id):
    """Get member name by ID"""
    for member in ALL_MEMBERS:
        if member['id'] == member_id:
            return member['name']
    return member_id

# ==================== STATIC FILES ====================

@app.route('/')
def serve_index():
    return send_from_directory('.', 'index.html')

@app.route('/<path:path>')
def serve_static(path):
    return send_from_directory('.', path)

# ==================== API ROUTES ====================

@app.route('/api/data', methods=['GET'])
def get_all_data():
    """Get all data (games + preferences)"""
    return jsonify({
        'games': get_all_games(),
        'playerPreferences': get_all_preferences(),
        'discordWebhook': ''  # Don't expose webhook to frontend
    })

# ==================== GAMES ====================

@app.route('/api/games', methods=['GET'])
def get_games():
    """Get all games"""
    return jsonify(get_all_games())

@app.route('/api/games', methods=['POST'])
def create_game():
    """Create a new game"""
    body = request.json

    date = body.get('date')
    time = body.get('time')
    opponent = body.get('opponent')
    notes = body.get('notes', '')

    if not all([date, time, opponent]):
        return jsonify({'error': 'Missing required fields'}), 400

    game = create_game_db({
        'date': date,
        'time': time,
        'opponent': opponent,
        'notes': notes
    })

    if game:
        return jsonify(game), 201
    return jsonify({'error': 'Failed to create game'}), 500

@app.route('/api/games/<game_id>', methods=['DELETE'])
def delete_game(game_id):
    """Delete a game"""
    if delete_game_db(game_id):
        return jsonify({'success': True})
    return jsonify({'error': 'Game not found'}), 404

@app.route('/api/games/<game_id>/roster', methods=['PUT'])
def update_roster(game_id):
    """Update game roster"""
    game = get_game_by_id(game_id)
    if not game:
        return jsonify({'error': 'Game not found'}), 404

    body = request.json
    roster = body.get('roster', [])

    if not isinstance(roster, list):
        return jsonify({'error': 'Roster must be an array'}), 400

    updated = update_game_db(game_id, {'roster': roster})
    return jsonify(updated)

@app.route('/api/games/<game_id>/availability', methods=['POST'])
def set_availability(game_id):
    """Set player availability for a game"""
    game = get_game_by_id(game_id)
    if not game:
        return jsonify({'error': 'Game not found'}), 404

    body = request.json
    player_id = body.get('playerId')
    is_available = body.get('isAvailable')

    if not player_id:
        return jsonify({'error': 'Player ID required'}), 400

    available = [p for p in game['available'] if p != player_id]
    unavailable = [p for p in game['unavailable'] if p != player_id]

    if is_available:
        available.append(player_id)
    else:
        unavailable.append(player_id)

    updated = update_game_db(game_id, {'available': available, 'unavailable': unavailable})
    return jsonify(updated)

# ==================== PLAYER PREFERENCES ====================

@app.route('/api/preferences', methods=['GET'])
def get_preferences():
    """Get all player preferences"""
    return jsonify(get_all_preferences())

@app.route('/api/preferences/<player_id>', methods=['PUT'])
def set_preference(player_id):
    """Set a player's preference"""
    body = request.json
    preference = body.get('preference')

    if preference not in ['starter', 'sub']:
        return jsonify({'error': 'Invalid preference. Must be "starter" or "sub"'}), 400

    result = set_preference_db(player_id, preference)
    return jsonify(result)

# ==================== DISCORD WEBHOOK ====================

@app.route('/api/webhook', methods=['GET'])
def get_webhook():
    """Get webhook status (masked for security)"""
    webhook = get_setting('discord_webhook') or ''
    return jsonify({
        'configured': bool(webhook),
        'preview': ('****' + webhook[-10:]) if webhook else None
    })

@app.route('/api/webhook', methods=['PUT'])
def set_webhook():
    """Set webhook URL"""
    body = request.json
    webhook = body.get('webhook', '')
    set_setting('discord_webhook', webhook)
    return jsonify({'success': True})

@app.route('/api/discord/post/<game_id>', methods=['POST'])
def post_to_discord(game_id):
    """Post game announcement to Discord"""
    webhook = get_setting('discord_webhook')
    if not webhook:
        return jsonify({'error': 'Discord webhook not configured'}), 400

    game = get_game_by_id(game_id)
    if not game:
        return jsonify({'error': 'Game not found'}), 404

    # Format date and time
    try:
        date_obj = datetime.strptime(game['date'], '%Y-%m-%d')
        formatted_date = date_obj.strftime('%A, %b %d, %Y')
    except:
        formatted_date = game['date']

    try:
        time_parts = game['time'].split(':')
        hour = int(time_parts[0])
        minute = time_parts[1]
        ampm = 'PM' if hour >= 12 else 'AM'
        hour12 = hour % 12 or 12
        formatted_time = f'{hour12}:{minute} {ampm} ET'
    except:
        formatted_time = game['time']

    roster_names = [get_member_name(pid) for pid in game.get('roster', [])]

    embed = {
        'title': f"🎮 Game Day: {formatted_date}",
        'color': 0x00f0ff,
        'fields': [
            {'name': '⏰ Time', 'value': formatted_time, 'inline': True},
            {'name': '⚔️ Opponent', 'value': game['opponent'], 'inline': True},
            {
                'name': f"👥 Roster ({len(roster_names)}/10)",
                'value': '\n'.join(roster_names) if roster_names else 'TBD',
                'inline': False
            }
        ],
        'footer': {'text': 'Game Over Pop1 War Team'},
        'timestamp': datetime.utcnow().isoformat()
    }

    if game.get('notes'):
        embed['fields'].append({
            'name': '📝 Notes',
            'value': game['notes'],
            'inline': False
        })

    try:
        response = requests.post(webhook, json={
            'username': 'Game Over Bot',
            'embeds': [embed]
        })

        if response.ok:
            return jsonify({'success': True})
        else:
            return jsonify({'error': 'Discord API error', 'details': response.text}), 500
    except Exception as e:
        return jsonify({'error': 'Failed to post to Discord', 'details': str(e)}), 500

# ==================== RUN SERVER ====================

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 3000))
    app.run(host='0.0.0.0', port=port, debug=False)
