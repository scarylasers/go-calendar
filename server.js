// Game Over Pop1 War Team Calendar - Backend Server
const express = require('express');
const cors = require('cors');
const fs = require('fs');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3000;
const DATA_FILE = path.join(__dirname, 'data.json');

// Middleware
app.use(cors());
app.use(express.json());
app.use(express.static(__dirname)); // Serve frontend files

// Initialize data file if it doesn't exist
function initDataFile() {
    if (!fs.existsSync(DATA_FILE)) {
        const initialData = {
            games: [],
            playerPreferences: {},
            discordWebhook: ''
        };
        fs.writeFileSync(DATA_FILE, JSON.stringify(initialData, null, 2));
    }
}

// Read data from file
function readData() {
    try {
        const data = fs.readFileSync(DATA_FILE, 'utf8');
        return JSON.parse(data);
    } catch (error) {
        console.error('Error reading data:', error);
        return { games: [], playerPreferences: {}, discordWebhook: '' };
    }
}

// Write data to file
function writeData(data) {
    try {
        fs.writeFileSync(DATA_FILE, JSON.stringify(data, null, 2));
        return true;
    } catch (error) {
        console.error('Error writing data:', error);
        return false;
    }
}

// Initialize on startup
initDataFile();

// ==================== API ROUTES ====================

// Get all data (games + preferences)
app.get('/api/data', (req, res) => {
    const data = readData();
    res.json(data);
});

// ==================== GAMES ====================

// Get all games
app.get('/api/games', (req, res) => {
    const data = readData();
    res.json(data.games);
});

// Create a new game
app.post('/api/games', (req, res) => {
    const data = readData();
    const { date, time, opponent, notes } = req.body;

    if (!date || !time || !opponent) {
        return res.status(400).json({ error: 'Missing required fields' });
    }

    const game = {
        id: 'game_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9),
        date,
        time,
        opponent,
        notes: notes || '',
        available: [],
        unavailable: [],
        roster: [],
        createdAt: new Date().toISOString()
    };

    data.games.push(game);
    writeData(data);

    res.status(201).json(game);
});

// Delete a game
app.delete('/api/games/:id', (req, res) => {
    const data = readData();
    const gameIndex = data.games.findIndex(g => g.id === req.params.id);

    if (gameIndex === -1) {
        return res.status(404).json({ error: 'Game not found' });
    }

    data.games.splice(gameIndex, 1);
    writeData(data);

    res.json({ success: true });
});

// Update game roster
app.put('/api/games/:id/roster', (req, res) => {
    const data = readData();
    const game = data.games.find(g => g.id === req.params.id);

    if (!game) {
        return res.status(404).json({ error: 'Game not found' });
    }

    const { roster } = req.body;
    if (!Array.isArray(roster)) {
        return res.status(400).json({ error: 'Roster must be an array' });
    }

    game.roster = roster;
    writeData(data);

    res.json(game);
});

// Set player availability for a game
app.post('/api/games/:id/availability', (req, res) => {
    const data = readData();
    const game = data.games.find(g => g.id === req.params.id);

    if (!game) {
        return res.status(404).json({ error: 'Game not found' });
    }

    const { playerId, isAvailable } = req.body;
    if (!playerId) {
        return res.status(400).json({ error: 'Player ID required' });
    }

    // Initialize arrays if needed
    if (!game.available) game.available = [];
    if (!game.unavailable) game.unavailable = [];

    // Remove from both lists first
    game.available = game.available.filter(id => id !== playerId);
    game.unavailable = game.unavailable.filter(id => id !== playerId);

    // Add to appropriate list
    if (isAvailable) {
        game.available.push(playerId);
    } else {
        game.unavailable.push(playerId);
    }

    writeData(data);
    res.json(game);
});

// ==================== PLAYER PREFERENCES ====================

// Get all player preferences
app.get('/api/preferences', (req, res) => {
    const data = readData();
    res.json(data.playerPreferences);
});

// Set a player's preference
app.put('/api/preferences/:playerId', (req, res) => {
    const data = readData();
    const { preference } = req.body;

    if (!['starter', 'sub'].includes(preference)) {
        return res.status(400).json({ error: 'Invalid preference. Must be "starter" or "sub"' });
    }

    data.playerPreferences[req.params.playerId] = preference;
    writeData(data);

    res.json({ playerId: req.params.playerId, preference });
});

// ==================== DISCORD WEBHOOK ====================

// Get webhook (masked for security)
app.get('/api/webhook', (req, res) => {
    const data = readData();
    res.json({
        configured: !!data.discordWebhook,
        preview: data.discordWebhook ? '****' + data.discordWebhook.slice(-10) : null
    });
});

// Set webhook
app.put('/api/webhook', (req, res) => {
    const data = readData();
    const { webhook } = req.body;

    data.discordWebhook = webhook || '';
    writeData(data);

    res.json({ success: true });
});

// Post to Discord
app.post('/api/discord/post/:gameId', async (req, res) => {
    const data = readData();

    if (!data.discordWebhook) {
        return res.status(400).json({ error: 'Discord webhook not configured' });
    }

    const game = data.games.find(g => g.id === req.params.gameId);
    if (!game) {
        return res.status(404).json({ error: 'Game not found' });
    }

    // Import members list for name lookup
    const { activeMembers, subMembers } = require('./members.js');
    const allMembers = [...activeMembers, ...subMembers];

    const getMemberName = (id) => {
        const member = allMembers.find(m => m.id === id);
        return member ? member.name : id;
    };

    const formatDate = (dateStr) => {
        const date = new Date(dateStr + 'T00:00:00');
        return date.toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric', year: 'numeric' });
    };

    const formatTime = (timeStr) => {
        const [hours, minutes] = timeStr.split(':');
        const hour = parseInt(hours);
        const ampm = hour >= 12 ? 'PM' : 'AM';
        const hour12 = hour % 12 || 12;
        return `${hour12}:${minutes} ${ampm} ET`;
    };

    const rosterNames = (game.roster || []).map(getMemberName);

    const embed = {
        title: `🎮 Game Day: ${formatDate(game.date)}`,
        color: 0x00f0ff,
        fields: [
            { name: '⏰ Time', value: formatTime(game.time), inline: true },
            { name: '⚔️ Opponent', value: game.opponent, inline: true },
            { name: `👥 Roster (${rosterNames.length}/10)`, value: rosterNames.length > 0 ? rosterNames.join('\n') : 'TBD', inline: false }
        ],
        footer: { text: 'Game Over Pop1 War Team' },
        timestamp: new Date().toISOString()
    };

    if (game.notes) {
        embed.fields.push({ name: '📝 Notes', value: game.notes, inline: false });
    }

    try {
        const response = await fetch(data.discordWebhook, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username: 'Game Over Bot',
                embeds: [embed]
            })
        });

        if (response.ok) {
            res.json({ success: true });
        } else {
            const text = await response.text();
            res.status(500).json({ error: 'Discord API error', details: text });
        }
    } catch (error) {
        res.status(500).json({ error: 'Failed to post to Discord', details: error.message });
    }
});

// ==================== START SERVER ====================

app.listen(PORT, () => {
    console.log(`
╔═══════════════════════════════════════════════════════╗
║     GAME OVER - Pop1 War Team Calendar Server         ║
╠═══════════════════════════════════════════════════════╣
║  Server running at: http://localhost:${PORT}             ║
║  API endpoints:     http://localhost:${PORT}/api         ║
╚═══════════════════════════════════════════════════════╝
    `);
});
