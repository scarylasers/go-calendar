"""
Game Over Pop1 War Team Calendar - Flask Backend
For PythonAnywhere deployment
"""

from flask import Flask, jsonify, request, send_from_directory
import json
import os
from datetime import datetime
import random
import string
import requests

app = Flask(__name__, static_folder='.', static_url_path='')

# Data file path - PythonAnywhere has persistent storage
DATA_FILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'data.json')

# ==================== DATA HELPERS ====================

def init_data_file():
    """Create data file if it doesn't exist"""
    if not os.path.exists(DATA_FILE):
        initial_data = {
            'games': [],
            'playerPreferences': {},
            'discordWebhook': ''
        }
        save_data(initial_data)

def load_data():
    """Load data from JSON file"""
    try:
        with open(DATA_FILE, 'r') as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return {'games': [], 'playerPreferences': {}, 'discordWebhook': ''}

def save_data(data):
    """Save data to JSON file"""
    with open(DATA_FILE, 'w') as f:
        json.dump(data, f, indent=2)

def generate_game_id():
    """Generate unique game ID"""
    timestamp = int(datetime.now().timestamp() * 1000)
    random_str = ''.join(random.choices(string.ascii_lowercase + string.digits, k=9))
    return f'game_{timestamp}_{random_str}'

# Initialize on startup
init_data_file()

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
    return jsonify(load_data())

# ==================== GAMES ====================

@app.route('/api/games', methods=['GET'])
def get_games():
    """Get all games"""
    data = load_data()
    return jsonify(data.get('games', []))

@app.route('/api/games', methods=['POST'])
def create_game():
    """Create a new game"""
    data = load_data()
    body = request.json

    date = body.get('date')
    time = body.get('time')
    opponent = body.get('opponent')
    notes = body.get('notes', '')

    if not all([date, time, opponent]):
        return jsonify({'error': 'Missing required fields'}), 400

    game = {
        'id': generate_game_id(),
        'date': date,
        'time': time,
        'opponent': opponent,
        'notes': notes,
        'available': [],
        'unavailable': [],
        'roster': [],
        'createdAt': datetime.now().isoformat()
    }

    data['games'].append(game)
    save_data(data)

    return jsonify(game), 201

@app.route('/api/games/<game_id>', methods=['DELETE'])
def delete_game(game_id):
    """Delete a game"""
    data = load_data()

    game_index = next((i for i, g in enumerate(data['games']) if g['id'] == game_id), None)

    if game_index is None:
        return jsonify({'error': 'Game not found'}), 404

    data['games'].pop(game_index)
    save_data(data)

    return jsonify({'success': True})

@app.route('/api/games/<game_id>/roster', methods=['PUT'])
def update_roster(game_id):
    """Update game roster"""
    data = load_data()

    game = next((g for g in data['games'] if g['id'] == game_id), None)

    if not game:
        return jsonify({'error': 'Game not found'}), 404

    body = request.json
    roster = body.get('roster', [])

    if not isinstance(roster, list):
        return jsonify({'error': 'Roster must be an array'}), 400

    game['roster'] = roster
    save_data(data)

    return jsonify(game)

@app.route('/api/games/<game_id>/availability', methods=['POST'])
def set_availability(game_id):
    """Set player availability for a game"""
    data = load_data()

    game = next((g for g in data['games'] if g['id'] == game_id), None)

    if not game:
        return jsonify({'error': 'Game not found'}), 404

    body = request.json
    player_id = body.get('playerId')
    is_available = body.get('isAvailable')

    if not player_id:
        return jsonify({'error': 'Player ID required'}), 400

    # Initialize arrays if needed
    if 'available' not in game:
        game['available'] = []
    if 'unavailable' not in game:
        game['unavailable'] = []

    # Remove from both lists first
    game['available'] = [p for p in game['available'] if p != player_id]
    game['unavailable'] = [p for p in game['unavailable'] if p != player_id]

    # Add to appropriate list
    if is_available:
        game['available'].append(player_id)
    else:
        game['unavailable'].append(player_id)

    save_data(data)
    return jsonify(game)

# ==================== PLAYER PREFERENCES ====================

@app.route('/api/preferences', methods=['GET'])
def get_preferences():
    """Get all player preferences"""
    data = load_data()
    return jsonify(data.get('playerPreferences', {}))

@app.route('/api/preferences/<player_id>', methods=['PUT'])
def set_preference(player_id):
    """Set a player's preference"""
    data = load_data()
    body = request.json
    preference = body.get('preference')

    if preference not in ['starter', 'sub']:
        return jsonify({'error': 'Invalid preference. Must be "starter" or "sub"'}), 400

    if 'playerPreferences' not in data:
        data['playerPreferences'] = {}

    data['playerPreferences'][player_id] = preference
    save_data(data)

    return jsonify({'playerId': player_id, 'preference': preference})

# ==================== DISCORD WEBHOOK ====================

@app.route('/api/webhook', methods=['GET'])
def get_webhook():
    """Get webhook status (masked for security)"""
    data = load_data()
    webhook = data.get('discordWebhook', '')
    return jsonify({
        'configured': bool(webhook),
        'preview': ('****' + webhook[-10:]) if webhook else None
    })

@app.route('/api/webhook', methods=['PUT'])
def set_webhook():
    """Set webhook URL"""
    data = load_data()
    body = request.json
    webhook = body.get('webhook', '')

    data['discordWebhook'] = webhook
    save_data(data)

    return jsonify({'success': True})

@app.route('/api/discord/post/<game_id>', methods=['POST'])
def post_to_discord(game_id):
    """Post game announcement to Discord"""
    data = load_data()

    webhook = data.get('discordWebhook')
    if not webhook:
        return jsonify({'error': 'Discord webhook not configured'}), 400

    game = next((g for g in data['games'] if g['id'] == game_id), None)
    if not game:
        return jsonify({'error': 'Game not found'}), 404

    # Format date and time
    from datetime import datetime as dt
    try:
        date_obj = dt.strptime(game['date'], '%Y-%m-%d')
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
    app.run(debug=True, port=3000)
