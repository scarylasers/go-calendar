// Game Over Pop1 War Team Calendar - Frontend App
// Uses backend API for shared data

// ==================== CONFIG ====================
const API_BASE = window.location.origin + '/api';

// ==================== MEMBER DATA ====================

// Active Members
const activeMembers = [
    { id: 'alock', name: 'GO-Alock4', year: 2021 },
    { id: 'cronides', name: 'GO-Cronides', year: 2021, note: 'Former Head of GO' },
    { id: 'ghostxrp', name: 'GO_GhostXRP', year: 2021 },
    { id: 'bramrianne', name: 'GO_ BramRianne', year: 2022, region: 'EU' },
    { id: 'thesean', name: 'GO_The.Sean', year: 2023, region: 'EU', note: 'Former Head of GO Europe' },
    { id: 'babs', name: 'GO_BABs', year: 2023 },
    { id: 'honeyluvv', name: 'GO_HoneyLuvv', year: 2023 },
    { id: 'zoloto', name: 'GO_Zoloto', year: 2023, region: 'EU' },
    { id: 'humanoid', name: 'GO_Humanoid', year: 2023, region: 'EU', note: 'Head of GO Europe' },
    { id: 'jesshawk', name: 'GO_JessHawk3', year: 2023, note: 'Head of GO' },
    { id: 'deathraider', name: 'GO_Deathraider255', year: 2024, region: 'EU' },
    { id: 'pinkpwnage', name: 'GOxPinkPWNAGE5', year: 2024 },
    { id: 'sami', name: 'Sami_008', year: 2024 },
    { id: 'silverlining', name: 'GO_SilverLining23', year: 2024, region: 'EU' },
    { id: 'colinthe5', name: 'Go_Colinthe5', year: 2024, note: 'Head of GO League' },
    { id: 'thedon', name: 'GO_The_Don', year: 2024 },
    { id: 'headhelper', name: 'GO HeadHelper', year: 2024 },
    { id: 'cosmo', name: 'GO_Cosmo', year: 2025 },
    { id: 'fxyz', name: 'f(x,y,z)', year: 2025 },
    { id: 'amberloaf', name: 'AmberLoaf', year: 2025 },
    { id: 'scarylasers', name: 'GO_ScaryLasers', year: 2025 },
    { id: 'glotones', name: 'GLOTONES', year: 2025 },
    { id: 'smich', name: 'GO_SMICH1989', year: 2025, region: 'EU' },
    { id: 'flexx', name: 'GO_FlexX', year: 2025 },
    { id: 'kygelli', name: 'GO_KyGelli', year: 2025 },
    { id: 'chr1sp', name: 'GO_Chr1sP', year: 2025 }
];

// Substitute Members (Vets)
const subMembers = [
    { id: 'shark', name: 'Shark', year: 2021 },
    { id: 'docbutler', name: 'GO_DocButler', year: 2021 },
    { id: 'drsmartazz', name: 'GO_DrSmartAzz', year: 2021 },
    { id: 'pbandc', name: 'PBandC-GO', year: 2021 },
    { id: 'maverick', name: 'GO_Maverick', year: 2021 },
    { id: 'honeygun', name: 'HoneyGUN', year: 2022 },
    { id: 'loki', name: 'GO_Loki714', year: 2022 },
    { id: 'lizlow', name: 'LizLow91', year: 2023 },
    { id: 'kc', name: 'GO_KC', year: 2023 },
    { id: 'lester', name: 'GO_Lester', year: 2024 },
    { id: 'kingslayer', name: 'GO_Kingslayer', year: 2024, region: 'EU' },
    { id: 'bacon', name: 'AllTheBaconAndEGGz', year: 2025 },
    { id: 'stooobe', name: 'GO_STOOOBE', year: 2025, note: 'Former Head of GO' },
    { id: 'johnharple', name: 'GO_JohnHarple', year: 2025 }
];

// All members combined
const allMembers = [
    ...activeMembers.map(m => ({ ...m, type: 'active' })),
    ...subMembers.map(m => ({ ...m, type: 'sub' }))
];

// ==================== STATE ====================

let state = {
    games: [],
    playerPreferences: {},
    currentPlayer: null,
    isManagerMode: false,
    loading: false
};

// ==================== API FUNCTIONS ====================

async function fetchData() {
    try {
        state.loading = true;
        const response = await fetch(`${API_BASE}/data`);
        const data = await response.json();
        state.games = data.games || [];
        state.playerPreferences = data.playerPreferences || {};
        return data;
    } catch (error) {
        console.error('Failed to fetch data:', error);
        showError('Failed to load data from server');
        return null;
    } finally {
        state.loading = false;
    }
}

async function createGame(gameData) {
    try {
        const response = await fetch(`${API_BASE}/games`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(gameData)
        });
        if (!response.ok) throw new Error('Failed to create game');
        return await response.json();
    } catch (error) {
        console.error('Failed to create game:', error);
        showError('Failed to create game');
        return null;
    }
}

async function deleteGameAPI(gameId) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}`, {
            method: 'DELETE'
        });
        if (!response.ok) throw new Error('Failed to delete game');
        return true;
    } catch (error) {
        console.error('Failed to delete game:', error);
        showError('Failed to delete game');
        return false;
    }
}

async function setAvailabilityAPI(gameId, playerId, isAvailable) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/availability`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ playerId, isAvailable })
        });
        if (!response.ok) throw new Error('Failed to set availability');
        return await response.json();
    } catch (error) {
        console.error('Failed to set availability:', error);
        showError('Failed to update availability');
        return null;
    }
}

async function updateRosterAPI(gameId, roster) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/roster`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ roster })
        });
        if (!response.ok) throw new Error('Failed to update roster');
        return await response.json();
    } catch (error) {
        console.error('Failed to update roster:', error);
        showError('Failed to save roster');
        return null;
    }
}

async function setPlayerPreferenceAPI(playerId, preference) {
    try {
        const response = await fetch(`${API_BASE}/preferences/${playerId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ preference })
        });
        if (!response.ok) throw new Error('Failed to set preference');
        return await response.json();
    } catch (error) {
        console.error('Failed to set preference:', error);
        showError('Failed to save preference');
        return null;
    }
}

async function postToDiscordAPI(gameId) {
    try {
        const response = await fetch(`${API_BASE}/discord/post/${gameId}`, {
            method: 'POST'
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to post');
        }
        return true;
    } catch (error) {
        console.error('Failed to post to Discord:', error);
        throw error;
    }
}

async function saveWebhookAPI(webhook) {
    try {
        const response = await fetch(`${API_BASE}/webhook`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ webhook })
        });
        if (!response.ok) throw new Error('Failed to save webhook');
        return true;
    } catch (error) {
        console.error('Failed to save webhook:', error);
        showError('Failed to save webhook');
        return false;
    }
}

async function getWebhookStatus() {
    try {
        const response = await fetch(`${API_BASE}/webhook`);
        return await response.json();
    } catch (error) {
        return { configured: false };
    }
}

// ==================== LOCAL STORAGE (for user identity only) ====================

function loadLocalSettings() {
    const savedPlayer = localStorage.getItem('go_calendar_player');
    if (savedPlayer) {
        state.currentPlayer = savedPlayer;
    }

    const savedManager = localStorage.getItem('go_calendar_manager');
    state.isManagerMode = savedManager === 'true';
}

function savePlayer(playerId) {
    localStorage.setItem('go_calendar_player', playerId);
}

function saveManagerMode(isManager) {
    localStorage.setItem('go_calendar_manager', isManager.toString());
}

// ==================== HELPERS ====================

function showError(message) {
    // Simple error display - could be improved with a toast notification
    alert(message);
}

function formatDate(dateStr) {
    const date = new Date(dateStr + 'T00:00:00');
    const options = { weekday: 'long', month: 'short', day: 'numeric', year: 'numeric' };
    return date.toLocaleDateString('en-US', options);
}

function formatTime(timeStr) {
    const [hours, minutes] = timeStr.split(':');
    const hour = parseInt(hours);
    const ampm = hour >= 12 ? 'PM' : 'AM';
    const hour12 = hour % 12 || 12;
    return `${hour12}:${minutes} ${ampm} ET`;
}

function getMemberById(id) {
    return allMembers.find(m => m.id === id);
}

function getPlayerPreference(playerId) {
    return state.playerPreferences[playerId] || 'starter';
}

function isSameDay(d1, d2) {
    return d1.getFullYear() === d2.getFullYear() &&
           d1.getMonth() === d2.getMonth() &&
           d1.getDate() === d2.getDate();
}

// ==================== RENDER FUNCTIONS ====================

function renderPlayerSelect() {
    const select = document.getElementById('currentPlayer');
    select.innerHTML = '<option value="">-- Select Your Name --</option>';

    // Active members first
    const activeGroup = document.createElement('optgroup');
    activeGroup.label = 'Active Members';
    activeMembers.forEach(member => {
        const option = document.createElement('option');
        option.value = member.id;
        option.textContent = member.name + (member.region ? ` [${member.region}]` : '');
        if (state.currentPlayer === member.id) option.selected = true;
        activeGroup.appendChild(option);
    });
    select.appendChild(activeGroup);

    // Subs
    const subGroup = document.createElement('optgroup');
    subGroup.label = 'Substitute Players (Vets)';
    subMembers.forEach(member => {
        const option = document.createElement('option');
        option.value = member.id;
        option.textContent = member.name + (member.region ? ` [${member.region}]` : '');
        if (state.currentPlayer === member.id) option.selected = true;
        subGroup.appendChild(option);
    });
    select.appendChild(subGroup);
}

function renderRoster() {
    const activeRoster = document.getElementById('activeRoster');
    const subRoster = document.getElementById('subRoster');

    activeRoster.innerHTML = activeMembers.map(member => {
        const pref = getPlayerPreference(member.id);
        return `
            <div class="roster-member">
                <div class="member-indicator"></div>
                <span class="member-name">${member.name}</span>
                ${pref === 'sub' ? '<span class="member-tag" style="background: var(--warning); color: var(--bg-dark);">Prefers Sub</span>' : ''}
                ${member.region ? `<span class="member-tag">${member.region}</span>` : ''}
                ${member.note ? `<span class="member-tag">${member.note}</span>` : ''}
            </div>
        `;
    }).join('');

    subRoster.innerHTML = subMembers.map(member => `
        <div class="roster-member sub">
            <div class="member-indicator"></div>
            <span class="member-name">${member.name}</span>
            ${member.region ? `<span class="member-tag">${member.region}</span>` : ''}
            ${member.note ? `<span class="member-tag">${member.note}</span>` : ''}
        </div>
    `).join('');
}

function renderGames() {
    const gamesList = document.getElementById('gamesList');

    // Sort games by date
    const sortedGames = [...state.games].sort((a, b) => {
        const dateA = new Date(a.date + 'T' + a.time);
        const dateB = new Date(b.date + 'T' + b.time);
        return dateA - dateB;
    });

    // Filter to upcoming games only
    const now = new Date();
    const upcomingGames = sortedGames.filter(game => {
        const gameDate = new Date(game.date + 'T' + game.time);
        return gameDate >= now || isSameDay(gameDate, now);
    });

    if (upcomingGames.length === 0) {
        gamesList.innerHTML = '<div class="no-games">No upcoming games scheduled. Check back later!</div>';
        return;
    }

    gamesList.innerHTML = upcomingGames.map(game => renderGameCard(game)).join('');
}

function renderGameCard(game) {
    const availableCount = game.available ? game.available.length : 0;
    const rosterCount = game.roster ? game.roster.length : 0;
    const isReady = rosterCount >= 10;

    // Check current player's status
    let playerStatus = 'pending';
    let statusText = 'Not responded';
    if (state.currentPlayer) {
        if (game.available && game.available.includes(state.currentPlayer)) {
            playerStatus = 'confirmed';
            statusText = 'You are available';
        } else if (game.unavailable && game.unavailable.includes(state.currentPlayer)) {
            playerStatus = 'declined';
            statusText = 'You are unavailable';
        }
    }

    // Get names for roster display
    const rosterNames = (game.roster || []).map(id => {
        const member = getMemberById(id);
        return member ? member.name : id;
    });

    return `
        <div class="game-card" data-game-id="${game.id}">
            <div class="game-header">
                <div>
                    <div class="game-date">${formatDate(game.date)}</div>
                    <div class="game-time">${formatTime(game.time)}</div>
                </div>
                <div class="game-status">
                    ${isReady ?
                        '<span class="status-badge status-ready">Roster Full</span>' :
                        `<span class="status-badge status-need-players">Need ${10 - rosterCount} more</span>`
                    }
                </div>
            </div>

            <div class="game-opponent">
                <span>vs</span> ${game.opponent}
            </div>

            ${game.notes ? `<div class="game-notes">${game.notes}</div>` : ''}

            ${rosterCount > 0 ? `
                <div class="availability-section">
                    <h4>Selected Roster (${rosterCount}/10)</h4>
                    <div class="selected-roster">
                        ${rosterNames.map(name => `<span class="player-chip roster">${name}</span>`).join('')}
                    </div>
                </div>
            ` : ''}

            ${availableCount > 0 ? `
                <div class="availability-section">
                    <h4>Available Players (${availableCount})</h4>
                    <div class="available-players">
                        ${(game.available || []).map(id => {
                            const member = getMemberById(id);
                            if (!member) return '';
                            const isVet = member.type === 'sub';
                            const prefersSub = getPlayerPreference(id) === 'sub';
                            return `<span class="player-chip ${prefersSub ? 'sub' : ''}">${member.name}${prefersSub ? ' (sub)' : ''}${isVet ? ' [vet]' : ''}</span>`;
                        }).join('')}
                    </div>
                </div>
            ` : ''}

            ${state.currentPlayer ? `
                <div class="availability-section">
                    <div class="your-status ${playerStatus}">
                        <strong>Your Status:</strong> ${statusText}
                    </div>
                    <div class="availability-buttons">
                        <button class="btn btn-success btn-small" onclick="setAvailability('${game.id}', true)">
                            I Can Play
                        </button>
                        <button class="btn btn-danger btn-small" onclick="setAvailability('${game.id}', false)">
                            Can't Make It
                        </button>
                    </div>
                </div>
            ` : `
                <div class="availability-section">
                    <p style="color: var(--warning);">Select your name above to confirm availability</p>
                </div>
            `}

            ${state.isManagerMode ? `
                <div class="availability-section">
                    <button class="btn btn-primary btn-small" onclick="openRosterModal('${game.id}')">
                        Select Roster
                    </button>
                </div>
            ` : ''}
        </div>
    `;
}

function renderManageGames() {
    const manageList = document.getElementById('manageGamesList');

    const sortedGames = [...state.games].sort((a, b) => {
        const dateA = new Date(a.date + 'T' + a.time);
        const dateB = new Date(b.date + 'T' + b.time);
        return dateA - dateB;
    });

    if (sortedGames.length === 0) {
        manageList.innerHTML = '<p style="color: var(--text-secondary);">No games created yet. Create one above!</p>';
        return;
    }

    manageList.innerHTML = sortedGames.map(game => `
        <div class="manage-game-item">
            <div class="manage-game-info">
                <span class="game-date">${formatDate(game.date)}</span>
                <span class="game-time">${formatTime(game.time)}</span>
                <span>vs ${game.opponent}</span>
            </div>
            <div class="manage-game-actions">
                <button class="btn btn-primary btn-small" onclick="openRosterModal('${game.id}')">
                    Set Roster
                </button>
                <button class="btn btn-warning btn-small" onclick="postToDiscord('${game.id}')">
                    Post to Discord
                </button>
                <button class="btn btn-danger btn-small" onclick="deleteGame('${game.id}')">
                    Delete
                </button>
            </div>
        </div>
    `).join('');
}

// ==================== ACTIONS ====================

async function setAvailability(gameId, isAvailable) {
    if (!state.currentPlayer) {
        alert('Please select your name first!');
        return;
    }

    const result = await setAvailabilityAPI(gameId, state.currentPlayer, isAvailable);
    if (result) {
        // Update local state
        const game = state.games.find(g => g.id === gameId);
        if (game) {
            game.available = result.available;
            game.unavailable = result.unavailable;
        }
        renderGames();
    }
}

async function handleCreateGame(event) {
    event.preventDefault();

    const date = document.getElementById('gameDate').value;
    const time = document.getElementById('gameTime').value;
    const opponent = document.getElementById('opponent').value;
    const notes = document.getElementById('gameNotes').value;

    const game = await createGame({ date, time, opponent, notes });
    if (game) {
        state.games.push(game);
        event.target.reset();
        renderGames();
        renderManageGames();
        alert('Game created successfully!');
    }
}

async function deleteGame(gameId) {
    if (!confirm('Are you sure you want to delete this game?')) return;

    const success = await deleteGameAPI(gameId);
    if (success) {
        state.games = state.games.filter(g => g.id !== gameId);
        renderGames();
        renderManageGames();
    }
}

function openRosterModal(gameId) {
    const game = state.games.find(g => g.id === gameId);
    if (!game) return;

    const modal = document.getElementById('rosterModal');
    const content = document.getElementById('rosterModalContent');

    const availableIds = game.available || [];
    const currentRoster = game.roster || [];

    // Separate available into starters and subs by preference
    const availableStarters = availableIds.filter(id => getPlayerPreference(id) === 'starter');
    const availableSubs = availableIds.filter(id => getPlayerPreference(id) === 'sub');

    content.innerHTML = `
        <p style="margin-bottom: 20px;">
            <strong>Game:</strong> ${formatDate(game.date)} at ${formatTime(game.time)} vs ${game.opponent}
        </p>

        <div class="roster-selection">
            <div class="roster-pool">
                <h4>Available Starters (confirmed)</h4>
                ${availableStarters.length > 0 ? availableStarters.map(id => {
                    const member = getMemberById(id);
                    if (!member) return '';
                    const isSelected = currentRoster.includes(id);
                    return `
                        <div class="player-select-item ${isSelected ? 'selected' : ''}" data-player-id="${id}">
                            <input type="checkbox" class="player-checkbox" ${isSelected ? 'checked' : ''}>
                            <span>${member.name}</span>
                            ${member.type === 'sub' ? '<span class="member-tag">Vet</span>' : ''}
                        </div>
                    `;
                }).join('') : '<p style="color: var(--text-secondary);">No starters confirmed</p>'}

                <h4 style="margin-top: 15px;">Available Subs (confirmed)</h4>
                ${availableSubs.length > 0 ? availableSubs.map(id => {
                    const member = getMemberById(id);
                    if (!member) return '';
                    const isSelected = currentRoster.includes(id);
                    return `
                        <div class="player-select-item ${isSelected ? 'selected' : ''}" data-player-id="${id}">
                            <input type="checkbox" class="player-checkbox" ${isSelected ? 'checked' : ''}>
                            <span>${member.name}</span>
                            <span class="member-tag" style="background: var(--warning); color: var(--bg-dark);">Prefers Sub</span>
                        </div>
                    `;
                }).join('') : '<p style="color: var(--text-secondary);">No subs confirmed</p>'}
            </div>

            <div class="roster-pool">
                <h4>All Members (not confirmed)</h4>
                ${allMembers.filter(m => !availableIds.includes(m.id)).map(member => {
                    const isSelected = currentRoster.includes(member.id);
                    const pref = getPlayerPreference(member.id);
                    return `
                        <div class="player-select-item unavailable ${isSelected ? 'selected' : ''}" data-player-id="${member.id}">
                            <input type="checkbox" class="player-checkbox" ${isSelected ? 'checked' : ''}>
                            <span>${member.name}</span>
                            ${pref === 'sub' ? '<span class="member-tag" style="background: var(--warning); color: var(--bg-dark);">Prefers Sub</span>' : ''}
                            ${member.type === 'sub' ? '<span class="member-tag">Vet</span>' : ''}
                            ${member.region ? `<span class="member-tag">${member.region}</span>` : ''}
                        </div>
                    `;
                }).join('')}
            </div>
        </div>

        <div class="roster-count" id="rosterCount">Selected: ${currentRoster.length}/10</div>

        <button class="btn btn-primary save-roster-btn" onclick="saveRoster('${gameId}')">
            Save Roster
        </button>
    `;

    // Add click handlers
    content.querySelectorAll('.player-select-item').forEach(item => {
        item.addEventListener('click', function(e) {
            if (e.target.type === 'checkbox') return;
            const checkbox = this.querySelector('.player-checkbox');
            checkbox.checked = !checkbox.checked;
            this.classList.toggle('selected', checkbox.checked);
            updateRosterCount();
        });

        item.querySelector('.player-checkbox').addEventListener('change', function() {
            item.classList.toggle('selected', this.checked);
            updateRosterCount();
        });
    });

    modal.classList.add('active');
}

function updateRosterCount() {
    const checked = document.querySelectorAll('#rosterModalContent .player-checkbox:checked').length;
    const countEl = document.getElementById('rosterCount');
    countEl.textContent = `Selected: ${checked}/10`;
    countEl.className = 'roster-count';
    if (checked === 10) countEl.classList.add('full');
    if (checked > 10) countEl.classList.add('over');
}

async function saveRoster(gameId) {
    const selectedPlayers = [];
    document.querySelectorAll('#rosterModalContent .player-checkbox:checked').forEach(checkbox => {
        const item = checkbox.closest('.player-select-item');
        const playerId = item.dataset.playerId;
        selectedPlayers.push(playerId);
    });

    if (selectedPlayers.length > 10) {
        alert('You can only select up to 10 players!');
        return;
    }

    const result = await updateRosterAPI(gameId, selectedPlayers);
    if (result) {
        const game = state.games.find(g => g.id === gameId);
        if (game) {
            game.roster = result.roster;
        }
        closeModals();
        renderGames();
        renderManageGames();
    }
}

async function postToDiscord(gameId) {
    try {
        await postToDiscordAPI(gameId);
        alert('Posted to Discord successfully!');
    } catch (error) {
        if (error.message.includes('not configured')) {
            alert('Discord webhook not configured. Set it in the Discord Settings section below.');
        } else {
            alert('Failed to post to Discord: ' + error.message);
        }
    }
}

async function saveWebhookSetting() {
    const url = document.getElementById('discordWebhook').value;
    const success = await saveWebhookAPI(url);
    if (success) {
        alert('Discord webhook saved!');
    }
}

async function loadWebhookSetting() {
    const status = await getWebhookStatus();
    if (status.configured && status.preview) {
        document.getElementById('discordWebhook').placeholder = `Configured (ends with ${status.preview})`;
    }
}

function updatePlayerPreferenceUI() {
    const prefDiv = document.getElementById('playerPreference');
    const prefSelect = document.getElementById('rolePreference');

    if (state.currentPlayer) {
        prefDiv.style.display = 'flex';
        prefSelect.value = getPlayerPreference(state.currentPlayer);
    } else {
        prefDiv.style.display = 'none';
    }
}

async function handlePreferenceChange(preference) {
    if (!state.currentPlayer) return;

    const result = await setPlayerPreferenceAPI(state.currentPlayer, preference);
    if (result) {
        state.playerPreferences[state.currentPlayer] = preference;
        renderGames();
        renderRoster();
    }
}

function closeModals() {
    document.querySelectorAll('.modal').forEach(modal => {
        modal.classList.remove('active');
    });
}

// ==================== INITIALIZATION ====================

async function init() {
    // Load local settings (who the user is)
    loadLocalSettings();

    // Fetch shared data from server
    await fetchData();

    // Render initial content
    renderPlayerSelect();
    renderRoster();
    renderGames();
    renderManageGames();

    // Update manager mode UI
    if (state.isManagerMode) {
        document.body.classList.add('manager-mode');
        document.getElementById('managerToggle').classList.add('active');
    }

    // Event listeners
    document.getElementById('currentPlayer').addEventListener('change', function() {
        state.currentPlayer = this.value;
        savePlayer(this.value);
        updatePlayerPreferenceUI();
        renderGames();
    });

    document.getElementById('rolePreference').addEventListener('change', function() {
        handlePreferenceChange(this.value);
    });

    document.getElementById('managerToggle').addEventListener('click', function() {
        state.isManagerMode = !state.isManagerMode;
        saveManagerMode(state.isManagerMode);
        document.body.classList.toggle('manager-mode', state.isManagerMode);
        this.classList.toggle('active', state.isManagerMode);
    });

    document.getElementById('createGameForm').addEventListener('submit', handleCreateGame);

    // Tab switching
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', function() {
            const tabId = this.dataset.tab;

            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

            this.classList.add('active');
            document.getElementById(tabId).classList.add('active');
        });
    });

    // Modal close buttons
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.addEventListener('click', closeModals);
    });

    // Close modal on outside click
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', function(e) {
            if (e.target === this) closeModals();
        });
    });

    // Load webhook status
    loadWebhookSetting();

    // Show preference if player already selected
    updatePlayerPreferenceUI();

    // Auto-refresh data every 30 seconds
    setInterval(async () => {
        await fetchData();
        renderGames();
        renderManageGames();
        renderRoster();
    }, 30000);
}

// Start the app
document.addEventListener('DOMContentLoaded', init);
