// Game Over Pop1 War Team Calendar - Frontend App
// Uses backend API with Discord OAuth

// ==================== CONFIG ====================
const API_BASE = window.location.origin + '/api';
const AUTH_BASE = window.location.origin + '/auth';

// ==================== MEMBER DATA ====================

let activeMembers = [
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

let subMembers = [
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

let allMembers = [
    ...activeMembers.map(m => ({ ...m, type: 'active' })),
    ...subMembers.map(m => ({ ...m, type: 'sub' }))
];

// ==================== STATE ====================

let state = {
    games: [],
    playerPreferences: {},
    currentPlayer: null,
    isManager: false,
    user: null, // Discord user info
    linkedUsers: {}, // Maps player IDs to their Discord info (avatar, etc.)
    leagues: [],
    divisions: [],
    editingGameId: null, // ID of game being edited, null if creating new
    loading: false
};

// ==================== AUTH FUNCTIONS ====================

async function checkAuth() {
    try {
        const response = await fetch(`${AUTH_BASE}/me`, { credentials: 'include' });
        const data = await response.json();

        if (data.authenticated) {
            state.user = {
                discordId: data.discordId,
                username: data.username,
                displayName: data.displayName,
                avatar: data.avatar,
                isManager: data.isManager,
                playerId: data.playerId,
                email: data.email || '',
                phone: data.phone || ''
            };
            state.isManager = data.isManager;
            state.currentPlayer = data.playerId || null;
            return true;
        }
        return false;
    } catch (error) {
        console.error('Auth check failed:', error);
        return false;
    }
}

async function logout() {
    try {
        await fetch(`${AUTH_BASE}/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        state.user = null;
        state.isManager = false;
        state.currentPlayer = null;
        updateAuthUI();
        renderAll();
    } catch (error) {
        console.error('Logout failed:', error);
    }
}

async function linkPlayer(playerId) {
    try {
        const response = await fetch(`${AUTH_BASE}/link`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ playerId })
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to link account');
        }

        state.currentPlayer = playerId;
        if (state.user) state.user.playerId = playerId;
        closeLinkModal();
        updateAuthUI();
        renderAll();
        return true;
    } catch (error) {
        console.error('Link player failed:', error);
        showError(error.message);
        return false;
    }
}

// ==================== API FUNCTIONS ====================

async function fetchData() {
    try {
        state.loading = true;
        const response = await fetch(`${API_BASE}/data`, { credentials: 'include' });
        const data = await response.json();
        state.games = data.games || [];
        state.playerPreferences = data.playerPreferences || {};

        // Fetch linked users for avatars
        await fetchLinkedUsers();

        return data;
    } catch (error) {
        console.error('Failed to fetch data:', error);
        showError('Failed to load data from server');
        return null;
    } finally {
        state.loading = false;
    }
}

async function fetchLinkedUsers() {
    try {
        const response = await fetch(`${API_BASE}/users/linked`, { credentials: 'include' });
        if (response.ok) {
            const data = await response.json();
            state.linkedUsers = data || {};
        }
    } catch (error) {
        console.error('Failed to fetch linked users:', error);
        state.linkedUsers = {};
    }
}

async function fetchMembers() {
    try {
        const response = await fetch(`${API_BASE}/members`, { credentials: 'include' });
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch members:', error);
        return { active: activeMembers, subs: subMembers };
    }
}

async function createGame(gameData) {
    try {
        const response = await fetch(`${API_BASE}/games`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(gameData)
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to create game');
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to create game:', error);
        showError(error.message);
        return null;
    }
}

async function updateGame(gameId, gameData) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(gameData)
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to update game');
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to update game:', error);
        showError(error.message);
        return null;
    }
}

async function deleteGameAPI(gameId) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}`, {
            method: 'DELETE',
            credentials: 'include'
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to delete game');
        }
        return true;
    } catch (error) {
        console.error('Failed to delete game:', error);
        showError(error.message);
        return false;
    }
}

async function setAvailabilityAPI(gameId, playerId, isAvailable) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/availability`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ playerId, isAvailable })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to set availability');
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to set availability:', error);
        showError(error.message);
        return null;
    }
}

async function updateRosterAPI(gameId, roster, subs = []) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/roster`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ roster, subs })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to update roster');
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to update roster:', error);
        showError(error.message);
        return null;
    }
}

async function withdrawFromRosterAPI(gameId) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/withdraw`, {
            method: 'POST',
            credentials: 'include'
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to withdraw from roster');
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to withdraw from roster:', error);
        showError(error.message);
        return null;
    }
}

async function setPlayerPreferenceAPI(playerId, preference) {
    try {
        const response = await fetch(`${API_BASE}/preferences/${playerId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ preference })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to set preference');
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to set preference:', error);
        showError(error.message);
        return null;
    }
}

async function postToDiscordAPI(gameId, mentionPlayers = false) {
    try {
        const url = `${API_BASE}/discord/post/${gameId}${mentionPlayers ? '?mention=true' : ''}`;
        const response = await fetch(url, {
            method: 'POST',
            credentials: 'include'
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
            credentials: 'include',
            body: JSON.stringify({ webhook })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to save webhook');
        }
        return true;
    } catch (error) {
        console.error('Failed to save webhook:', error);
        showError(error.message);
        return false;
    }
}

// ==================== LEAGUES & DIVISIONS ====================

async function fetchLeagues() {
    try {
        const response = await fetch(`${API_BASE}/leagues`);
        state.leagues = await response.json();
        populateLeagueDropdown();
        renderLeaguesList();
    } catch (error) {
        console.error('Failed to fetch leagues:', error);
    }
}

async function fetchDivisions() {
    try {
        const response = await fetch(`${API_BASE}/divisions`);
        state.divisions = await response.json();
        populateDivisionDropdown();
        renderDivisionsList();
    } catch (error) {
        console.error('Failed to fetch divisions:', error);
    }
}

function populateLeagueDropdown() {
    const select = document.getElementById('gameLeague');
    if (!select) return;
    select.innerHTML = '<option value="">Select League</option>';
    state.leagues.forEach(league => {
        const opt = document.createElement('option');
        opt.value = league;
        opt.textContent = league;
        select.appendChild(opt);
    });
}

function populateDivisionDropdown() {
    const select = document.getElementById('gameDivision');
    if (!select) return;
    select.innerHTML = '<option value="">Select Division</option>';
    state.divisions.forEach(div => {
        const opt = document.createElement('option');
        opt.value = div;
        opt.textContent = div;
        select.appendChild(opt);
    });
}

function renderLeaguesList() {
    const container = document.getElementById('leaguesList');
    if (!container) return;
    container.innerHTML = state.leagues.map(league => `
        <div class="item-row">
            <span>${league}</span>
            <button class="btn-remove" onclick="removeLeague('${league}')" title="Remove">×</button>
        </div>
    `).join('') || '<p class="no-items">No leagues added yet</p>';
}

function renderDivisionsList() {
    const container = document.getElementById('divisionsList');
    if (!container) return;
    container.innerHTML = state.divisions.map(div => `
        <div class="item-row">
            <span>${div}</span>
            <button class="btn-remove" onclick="removeDivision('${div}')" title="Remove">×</button>
        </div>
    `).join('') || '<p class="no-items">No divisions added yet</p>';
}

async function addLeague() {
    const input = document.getElementById('newLeagueName');
    const name = input.value.trim();
    if (!name) return;

    try {
        const response = await fetch(`${API_BASE}/leagues`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ name })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to add league');
        }
        state.leagues = await response.json();
        input.value = '';
        populateLeagueDropdown();
        renderLeaguesList();
    } catch (error) {
        showError(error.message);
    }
}

async function removeLeague(name) {
    if (!confirm(`Remove league "${name}"?`)) return;

    try {
        const response = await fetch(`${API_BASE}/leagues/${encodeURIComponent(name)}`, {
            method: 'DELETE',
            credentials: 'include'
        });
        if (!response.ok) throw new Error('Failed to remove league');
        state.leagues = await response.json();
        populateLeagueDropdown();
        renderLeaguesList();
    } catch (error) {
        showError(error.message);
    }
}

async function addDivision() {
    const input = document.getElementById('newDivisionName');
    const name = input.value.trim();
    if (!name) return;

    try {
        const response = await fetch(`${API_BASE}/divisions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ name })
        });
        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Failed to add division');
        }
        state.divisions = await response.json();
        input.value = '';
        populateDivisionDropdown();
        renderDivisionsList();
    } catch (error) {
        showError(error.message);
    }
}

async function removeDivision(name) {
    if (!confirm(`Remove division "${name}"?`)) return;

    try {
        const response = await fetch(`${API_BASE}/divisions/${encodeURIComponent(name)}`, {
            method: 'DELETE',
            credentials: 'include'
        });
        if (!response.ok) throw new Error('Failed to remove division');
        state.divisions = await response.json();
        populateDivisionDropdown();
        renderDivisionsList();
    } catch (error) {
        showError(error.message);
    }
}

// ==================== UI FUNCTIONS ====================

function updateAuthUI() {
    const loginSection = document.getElementById('loginSection');
    const userSection = document.getElementById('userSection');
    const managerTab = document.querySelector('.tab[data-tab="manage"]');
    const rosterLoginRequired = document.getElementById('rosterLoginRequired');
    const rosterContent = document.getElementById('rosterContent');
    const addMemberSection = document.getElementById('addMemberSection');

    if (state.user) {
        loginSection.style.display = 'none';
        userSection.style.display = 'flex';

        // Show roster content for logged-in users
        if (rosterLoginRequired) rosterLoginRequired.style.display = 'none';
        if (rosterContent) rosterContent.style.display = 'block';

        // Update user info
        const avatar = document.getElementById('userAvatar');
        const userName = document.getElementById('userName');
        const managerBadge = document.getElementById('managerBadge');
        const linkedPlayer = document.getElementById('linkedPlayer');
        const playerPreference = document.getElementById('playerPreference');

        if (state.user.avatar) {
            avatar.src = state.user.avatar;
            avatar.style.display = 'block';
        } else {
            avatar.style.display = 'none';
        }

        userName.textContent = state.user.displayName || state.user.username;
        managerBadge.style.display = state.isManager ? 'inline-block' : 'none';
        if (managerTab) managerTab.style.display = state.isManager ? 'inline-block' : 'none';
        if (addMemberSection) addMemberSection.style.display = state.isManager ? 'block' : 'none';

        // Show linked player or prompt to link
        const myAccountBtn = document.getElementById('myAccountBtn');
        if (state.currentPlayer) {
            const member = allMembers.find(m => m.id === state.currentPlayer);
            if (linkedPlayer) linkedPlayer.textContent = member ? `Playing as: ${member.name}` : '';
            if (playerPreference) playerPreference.style.display = 'flex';
            if (myAccountBtn) myAccountBtn.style.display = 'inline-block';

            // Set current preference
            const roleSelect = document.getElementById('rolePreference');
            if (roleSelect) roleSelect.value = state.playerPreferences[state.currentPlayer] || 'starter';
        } else {
            if (linkedPlayer) linkedPlayer.innerHTML = '<a href="#" onclick="showLinkModal(); return false;">Link your player profile</a>';
            if (playerPreference) playerPreference.style.display = 'none';
            if (myAccountBtn) myAccountBtn.style.display = 'none';
        }
    } else {
        loginSection.style.display = 'flex';
        userSection.style.display = 'none';
        if (managerTab) managerTab.style.display = 'none';
        if (addMemberSection) addMemberSection.style.display = 'none';

        // Hide roster content for logged-out users
        if (rosterLoginRequired) rosterLoginRequired.style.display = 'block';
        if (rosterContent) rosterContent.style.display = 'none';
    }
}

function showLinkModal() {
    const modal = document.getElementById('linkModal');
    const select = document.getElementById('linkPlayerSelect');

    // Populate select with unlinked members
    fetchMembers().then(data => {
        select.innerHTML = '<option value="">-- Select Your Name --</option>';

        const addOptions = (members, groupLabel) => {
            const group = document.createElement('optgroup');
            group.label = groupLabel;
            members.forEach(m => {
                if (!m.linked) {
                    const option = document.createElement('option');
                    option.value = m.id;
                    option.textContent = m.name + (m.region ? ` [${m.region}]` : '');
                    group.appendChild(option);
                }
            });
            if (group.children.length > 0) {
                select.appendChild(group);
            }
        };

        addOptions(data.active || [], 'Active Members');
        addOptions(data.subs || [], 'Substitute Players');
    });

    modal.classList.add('active');
}

function closeLinkModal() {
    document.getElementById('linkModal').classList.remove('active');
}

function showError(message) {
    // Simple error display
    alert(message);
}

function getMemberName(id) {
    const member = allMembers.find(m => m.id === id);
    return member ? member.name : id;
}

function getMemberType(id) {
    const member = allMembers.find(m => m.id === id);
    return member?.type || 'active';
}

function formatDate(dateStr) {
    const date = new Date(dateStr + 'T00:00:00');
    const options = { weekday: 'long', month: 'short', day: 'numeric' };
    return date.toLocaleDateString('en-US', options);
}

function formatTime(timeStr) {
    const [hours, minutes] = timeStr.split(':');
    const hour = parseInt(hours);
    const ampm = hour >= 12 ? 'PM' : 'AM';
    const hour12 = hour % 12 || 12;
    return `${hour12}:${minutes} ${ampm} ET`;
}

// ==================== RENDER FUNCTIONS ====================

function renderAll() {
    renderGames();
    renderRoster();
    if (state.isManager) {
        renderManageGames();
        renderManageRoster();
    }
}

function renderGames() {
    const container = document.getElementById('gamesList');
    if (!container) return;

    const now = new Date();
    const todayStr = now.toISOString().split('T')[0];

    // Filter future games
    const futureGames = state.games.filter(g => g.date >= todayStr);

    if (futureGames.length === 0) {
        container.innerHTML = '<p class="no-games">No upcoming games scheduled</p>';
        return;
    }

    container.innerHTML = futureGames.map(game => renderGameCard(game)).join('');
}

function renderGameCard(game) {
    const isOnRoster = game.roster?.includes(state.currentPlayer);
    const isAvailable = game.available?.includes(state.currentPlayer);
    const isUnavailable = game.unavailable?.includes(state.currentPlayer);
    const hasWithdrawn = game.withdrawals?.includes(state.currentPlayer);

    const availableCount = game.available?.length || 0;
    const rosterCount = game.roster?.length || 0;
    const teamSize = game.teamSize || 10;
    const needsPlayers = rosterCount < teamSize;
    const withdrawalCount = game.withdrawals?.length || 0;

    let availabilityButtons = '';
    if (state.user && state.currentPlayer) {
        if (isOnRoster) {
            // Show withdraw button if on roster
            availabilityButtons = `
                <div class="availability-buttons">
                    <span class="on-roster-badge">✓ You're on the roster!</span>
                    <button class="btn btn-warning" onclick="withdrawFromRoster('${game.id}')">
                        I need a sub to cover for me
                    </button>
                </div>
            `;
        } else if (hasWithdrawn) {
            availabilityButtons = `
                <div class="availability-buttons">
                    <span class="withdrawn-badge">You requested a sub for this game</span>
                </div>
            `;
        } else {
            availabilityButtons = `
                <div class="availability-buttons">
                    <button class="btn btn-available ${isAvailable ? 'active' : ''}"
                            onclick="setAvailability('${game.id}', true)">
                        I Can Play
                    </button>
                    <button class="btn btn-unavailable ${isUnavailable ? 'active' : ''}"
                            onclick="setAvailability('${game.id}', false)">
                        Can't Make It
                    </button>
                </div>
            `;
        }
    } else if (!state.user) {
        availabilityButtons = '<p class="login-prompt">Login to mark availability</p>';
    } else {
        availabilityButtons = '<p class="login-prompt">Link your profile to mark availability</p>';
    }

    // Available players list
    const availablePlayers = (game.available || []).map(id => {
        const member = allMembers.find(m => m.id === id);
        const isSub = member?.type === 'sub';
        const pref = state.playerPreferences[id];
        return `<span class="player-chip ${isSub ? 'sub' : ''}">${getMemberName(id)}${pref === 'sub' ? ' (sub)' : ''}</span>`;
    }).join('');

    // Roster list - vertical with large chips
    const rosterPlayers = (game.roster || []).map(id => {
        return `<div class="roster-chip-large">${getMemberName(id)}</div>`;
    }).join('');

    // Subs list (backup players) - vertical with large chips
    const subPlayers = (game.subs || []).map(id => {
        return `<div class="roster-chip-large sub-chip-large">${getMemberName(id)}</div>`;
    }).join('');

    // Calculate countdown
    const countdown = getGameCountdown(game.date, game.time);
    const countdownClass = countdown.urgent ? 'urgent' : '';

    // Game mode badge (always show)
    const gameMode = game.gameMode || 'War';
    const gameModeDisplay = `<span class="game-mode-badge">${gameMode}</span>`;

    // Status badges
    const rosterStatusClass = needsPlayers ? 'status-need-players' : 'status-ready';

    return `
        <div class="game-card ${isOnRoster ? 'on-roster' : ''}">
            <div class="game-header-centered">
                <div class="game-datetime">
                    <span class="game-date">${formatDate(game.date)}</span>
                    <span class="datetime-separator">•</span>
                    <span class="game-time">${formatTime(game.time)}</span>
                </div>
                <div class="game-countdown ${countdownClass}">
                    <span class="countdown-label">Gametime in:</span>
                    <span class="countdown-time ${countdownClass}">${countdown.text}</span>
                </div>
            </div>
            <div class="game-mode-badge">${gameMode}</div>
            <div class="game-opponent">vs ${game.opponent}</div>

            ${rosterPlayers ? `
                <div class="final-roster-section">
                    <h4>Final Roster (${rosterCount}/${teamSize})</h4>
                    <div class="roster-list-vertical">${rosterPlayers}</div>
                </div>
            ` : ''}

            ${subPlayers ? `
                <div class="final-roster-section subs-section">
                    <h4>Backup Subs</h4>
                    <div class="roster-list-vertical">${subPlayers}</div>
                </div>
            ` : ''}

            <div class="game-meta">
                ${game.league ? `<span class="game-league">${game.league}${game.division ? ` - ${game.division}` : ''}</span>` : ''}
            </div>
            ${game.notes ? `<div class="game-notes">${game.notes}</div>` : ''}

            <div class="game-status">
                <span class="status-badge ${rosterStatusClass}">
                    Roster: ${rosterCount}/${teamSize}${needsPlayers ? ` <span class="needs-more">(${teamSize - rosterCount} more needed)</span>` : ''}
                </span>
                <span class="status-badge available-badge">${availableCount} Available</span>
                ${withdrawalCount > 0 && state.isManager ? `<span class="status-badge withdrawal-badge clickable" onclick="openRosterModal('${game.id}')" title="Click to assign subs">⚠️ ${withdrawalCount} Need${withdrawalCount > 1 ? '' : 's'} Sub</span>` : ''}
                ${withdrawalCount > 0 && !state.isManager ? `<span class="status-badge withdrawal-badge">⚠️ ${withdrawalCount} Need${withdrawalCount > 1 ? '' : 's'} Sub</span>` : ''}
            </div>

            ${availabilityButtons}

            ${state.isManager ? `
                <div class="manager-actions">
                    <button class="btn btn-secondary" onclick="openRosterModal('${game.id}')">Select Roster</button>
                    <div class="discord-post-group">
                        <button class="btn btn-discord-small" onclick="postToDiscord('${game.id}', false)">Post to Discord</button>
                        <button class="btn btn-discord-mention" onclick="postToDiscord('${game.id}', true)" title="Post with @mentions">Post to Discord and Mention</button>
                    </div>
                </div>
            ` : ''}

            ${availablePlayers ? `
                <div class="available-section">
                    <h4>Available Players</h4>
                    <div class="player-chips">${availablePlayers}</div>
                </div>
            ` : ''}
        </div>
    `;
}

function renderRoster() {
    const naPlayersContainer = document.getElementById('naPlayers');
    const euPlayersContainer = document.getElementById('euPlayers');
    const vetsContainer = document.getElementById('vetsRoster');

    if (!naPlayersContainer) return;

    // Sort by year joined (oldest first)
    const sortByYear = (a, b) => a.year - b.year;

    // Active members (not retired/vets)
    const active = activeMembers.slice().sort(sortByYear);

    // Vets (substitute members are considered "vets" or retired)
    const vets = subMembers.slice().sort(sortByYear);

    // Split active by region
    const naPlayers = active.filter(m => !m.region || m.region === 'NA');
    const euPlayers = active.filter(m => m.region === 'EU');

    // Render each section (simplified - no starter/sub distinction)
    naPlayersContainer.innerHTML = naPlayers.map(m => renderSimpleMemberCard(m)).join('') || '<p class="empty-list">No NA players</p>';
    euPlayersContainer.innerHTML = euPlayers.map(m => renderSimpleMemberCard(m)).join('') || '<p class="empty-list">No EU players</p>';
    vetsContainer.innerHTML = vets.map(m => renderSimpleMemberCard(m, true)).join('') || '<p class="empty-list">No vets</p>';
}

function renderSimpleMemberCard(member, isVet = false) {
    const linkedUser = state.linkedUsers ? state.linkedUsers[member.id] : null;
    const avatarUrl = linkedUser?.avatar || '';
    const hasAvatar = avatarUrl && avatarUrl.length > 0;
    const isNA = !member.region || member.region === 'NA';
    const isEU = member.region === 'EU';

    // Manager edit controls
    const managerControls = state.isManager ? `
        <div class="member-edit-controls">
            <label class="edit-checkbox" title="NA Region">
                <input type="checkbox" ${isNA ? 'checked' : ''} onchange="updateMemberRegion('${member.id}', this.checked ? 'NA' : 'EU')">
                <span>NA</span>
            </label>
            <label class="edit-checkbox" title="EU Region">
                <input type="checkbox" ${isEU ? 'checked' : ''} onchange="updateMemberRegion('${member.id}', this.checked ? 'EU' : 'NA')">
                <span>EU</span>
            </label>
            <label class="edit-checkbox" title="Active Member">
                <input type="checkbox" ${!member.isVet ? 'checked' : ''} onchange="updateMemberStatus('${member.id}', !this.checked)">
                <span>Active</span>
            </label>
            <label class="edit-checkbox" title="Vet (Retired)">
                <input type="checkbox" ${member.isVet ? 'checked' : ''} onchange="updateMemberStatus('${member.id}', this.checked)">
                <span>Vet</span>
            </label>
            <button class="btn-remove-member" onclick="removeMember('${member.id}')" title="Remove">✕</button>
        </div>
    ` : '';

    return `
        <div class="member-card-simple ${isVet ? 'vet-member' : ''}" data-member-id="${member.id}">
            <div class="member-info-row">
                ${hasAvatar
                    ? `<img class="member-avatar-small linked" src="${avatarUrl}" alt="" onerror="this.style.display='none'">`
                    : `<div class="member-avatar-small">${member.name.charAt(0).toUpperCase()}</div>`
                }
                <span class="member-name-small">${member.name}</span>
                <span class="member-year-small">${member.year}</span>
            </div>
            ${managerControls}
        </div>
    `;
}

function renderMemberCard(member, isVet = false) {
    const pref = state.playerPreferences[member.id] || 'starter';
    const linkedUser = state.linkedUsers ? state.linkedUsers[member.id] : null;
    const avatarUrl = linkedUser?.avatar || '';
    const hasAvatar = avatarUrl && avatarUrl.length > 0;

    const isCurrentUser = state.currentPlayer === member.id;

    return `
        <div class="member-card ${isVet ? 'vet-member' : ''}" data-member-id="${member.id}">
            ${hasAvatar
                ? `<img class="member-avatar linked" src="${avatarUrl}" alt="" onerror="this.style.display='none'">`
                : `<div class="member-avatar">${member.name.charAt(0).toUpperCase()}</div>`
            }
            <div class="member-details">
                <div class="member-name">${member.name}</div>
                <div class="member-year">Since ${member.year}</div>
            </div>
            ${isCurrentUser ? `
                <div class="member-actions">
                    ${!isVet ? `
                        <button onclick="togglePreference('${member.id}')" title="Toggle Starter/Sub">
                            ${pref === 'sub' ? '⬆️ Starter' : '⬇️ Sub'}
                        </button>
                        <button class="retire-btn" onclick="toggleRetire('${member.id}', true)" title="Retire to Vets">
                            🎖️ Retire
                        </button>
                    ` : `
                        <button onclick="toggleRetire('${member.id}', false)" title="Return to Active">
                            ↩️ Unretire
                        </button>
                    `}
                </div>
            ` : ''}
        </div>
    `;
}

async function togglePreference(memberId) {
    const currentPref = state.playerPreferences[memberId] || 'starter';
    const newPref = currentPref === 'starter' ? 'sub' : 'starter';

    state.playerPreferences[memberId] = newPref;
    renderRoster();
    showSaveStatus('Unsaved changes');
}

async function toggleRetire(memberId, retire) {
    // Update member's isVet status (retire = move to vets)
    const result = await updateMemberAPI(memberId, { isVet: retire });
    if (result) {
        await fetchData();
        renderAll();
        showSaveStatus(retire ? 'Moved to Vets' : 'Returned to Active');
    }
}

async function saveRosterPreferences() {
    showSaveStatus('Saving...', true);

    // Save all preferences
    const promises = Object.entries(state.playerPreferences).map(([playerId, pref]) => {
        return fetch(`/api/preferences/${playerId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ preference: pref })
        });
    });

    try {
        await Promise.all(promises);
        showSaveStatus('Saved!');
        setTimeout(() => showSaveStatus(''), 2000);
    } catch (error) {
        showSaveStatus('Error saving');
    }
}

function showSaveStatus(message, isSaving = false) {
    const status = document.getElementById('rosterSaveStatus');
    if (status) {
        status.textContent = message;
        status.className = 'save-status' + (isSaving ? ' saving' : '');
    }
}

function renderManageGames() {
    const container = document.getElementById('manageGamesList');
    if (!container) return;

    if (state.games.length === 0) {
        container.innerHTML = '<p class="no-games">No games created yet</p>';
        return;
    }

    container.innerHTML = state.games.map(game => `
        <div class="manage-game-card">
            <div class="game-info">
                <strong>${formatDate(game.date)}</strong> at ${formatTime(game.time)}
                <br>vs ${game.opponent}
                ${game.league ? `<span class="game-tag">${game.league}${game.division ? ` - ${game.division}` : ''}</span>` : ''}
                ${game.notes ? `<br><em>${game.notes}</em>` : ''}
            </div>
            <div class="game-actions">
                <button class="btn btn-secondary btn-small" onclick="editGame('${game.id}')">Edit</button>
                <button class="btn btn-danger btn-small" onclick="deleteGame('${game.id}')">Delete</button>
            </div>
        </div>
    `).join('');
}

// ==================== ACTION FUNCTIONS ====================

async function setAvailability(gameId, isAvailable) {
    if (!state.currentPlayer) {
        showError('Please link your profile first');
        return;
    }

    const result = await setAvailabilityAPI(gameId, state.currentPlayer, isAvailable);
    if (result) {
        // Update local state
        const game = state.games.find(g => g.id === gameId);
        if (game) {
            game.available = result.available;
            game.unavailable = result.unavailable;
            renderGames();
        }
    }
}

async function deleteGame(gameId) {
    if (!confirm('Are you sure you want to delete this game?')) return;

    const success = await deleteGameAPI(gameId);
    if (success) {
        state.games = state.games.filter(g => g.id !== gameId);
        renderAll();
    }
}

function editGame(gameId) {
    const game = state.games.find(g => g.id === gameId);
    if (!game) return;

    state.editingGameId = gameId;

    // Populate form fields
    document.getElementById('gameDate').value = game.date;
    document.getElementById('gameTime').value = game.time;
    document.getElementById('opponent').value = game.opponent;
    document.getElementById('gameLeague').value = game.league || '';
    document.getElementById('gameDivision').value = game.division || '';
    document.getElementById('gameNotes').value = game.notes || '';

    // Handle game mode
    const gameModeSelect = document.getElementById('gameMode');
    const gameModeCustom = document.getElementById('gameModeCustom');
    const standardModes = ['War', 'Squads', 'Legions'];

    if (standardModes.includes(game.gameMode)) {
        gameModeSelect.value = game.gameMode;
        gameModeCustom.style.display = 'none';
    } else {
        gameModeSelect.value = 'Other';
        gameModeCustom.value = game.gameMode || '';
        gameModeCustom.style.display = 'block';
    }

    // Team size
    const teamSizeSelect = document.getElementById('teamSize');
    if (teamSizeSelect) {
        teamSizeSelect.value = game.teamSize || '10';
    }

    // Update form UI for edit mode
    updateFormMode(true);

    // Scroll to form
    document.getElementById('createGameForm').scrollIntoView({ behavior: 'smooth' });
}

function cancelEdit() {
    state.editingGameId = null;
    document.getElementById('createGameForm').reset();

    // Reset custom game mode
    const gameModeCustom = document.getElementById('gameModeCustom');
    if (gameModeCustom) {
        gameModeCustom.style.display = 'none';
        gameModeCustom.value = '';
    }

    // Reset team size
    const teamSizeSelect = document.getElementById('teamSize');
    if (teamSizeSelect) {
        teamSizeSelect.value = '10';
    }

    updateFormMode(false);
}

function updateFormMode(isEditing) {
    const submitBtn = document.querySelector('#createGameForm button[type="submit"]');
    const formTitle = document.querySelector('.create-game-form h3') || document.querySelector('.manage-section h3');
    const cancelBtn = document.getElementById('cancelEditBtn');

    if (submitBtn) {
        submitBtn.textContent = isEditing ? 'Update Game' : 'Create Game';
    }

    if (cancelBtn) {
        cancelBtn.style.display = isEditing ? 'inline-block' : 'none';
    }
}

async function withdrawFromRoster(gameId) {
    if (!confirm('Need a sub? This will remove you from the roster and notify managers to find a replacement.')) {
        return;
    }

    const result = await withdrawFromRosterAPI(gameId);
    if (result) {
        // Update local state
        const game = state.games.find(g => g.id === gameId);
        if (game) {
            game.roster = result.roster;
            game.withdrawals = result.withdrawals;
            game.available = result.available;
            game.unavailable = result.unavailable;
            renderGames();
        }
        alert('Managers have been notified to find a sub for you.');
    }
}

async function postToDiscord(gameId, mentionPlayers = false) {
    try {
        await postToDiscordAPI(gameId, mentionPlayers);
        alert(mentionPlayers ? 'Posted to Discord with @mentions!' : 'Posted to Discord successfully!');
    } catch (error) {
        showError(error.message);
    }
}

function openRosterModal(gameId) {
    const modal = document.getElementById('rosterModal');
    const content = document.getElementById('rosterModalContent');
    const game = state.games.find(g => g.id === gameId);

    if (!game) return;

    const currentRoster = game.roster || [];
    const currentSubs = game.subs || [];
    const teamSize = game.teamSize || 10;
    const withdrawals = game.withdrawals || [];

    const renderPlayerCheckbox = (member, isAvailable, section) => {
        const isOnRoster = currentRoster.includes(member.id);
        const isOnSubs = currentSubs.includes(member.id);
        const hasWithdrawn = withdrawals.includes(member.id);
        const pref = state.playerPreferences[member.id];
        const isMemberSub = member.type === 'sub';

        return `
            <div class="roster-player-row ${hasWithdrawn ? 'withdrawn' : ''}">
                <span class="player-name">
                    ${member.name}
                    ${member.region ? `<span class="tag">[${member.region}]</span>` : ''}
                    ${pref === 'sub' ? '<span class="tag">(prefers sub)</span>' : ''}
                    ${isAvailable ? '<span class="tag available">Available</span>' : ''}
                    ${hasWithdrawn ? '<span class="tag withdrawn">Needs Sub</span>' : ''}
                </span>
                <div class="roster-player-actions">
                    <button class="roster-btn ${isOnRoster ? 'active' : ''}"
                            onclick="togglePlayerRoster('${gameId}', '${member.id}', 'roster')"
                            ${isOnRoster ? '' : ''}>
                        Roster
                    </button>
                    <button class="roster-btn sub-btn ${isOnSubs ? 'active' : ''}"
                            onclick="togglePlayerRoster('${gameId}', '${member.id}', 'sub')">
                        Sub
                    </button>
                </div>
            </div>
        `;
    };

    // Get all available players
    const availablePlayers = allMembers.filter(m => game.available?.includes(m.id));
    const otherPlayers = allMembers.filter(m => !game.available?.includes(m.id) && !withdrawals.includes(m.id));

    // Show players needing subs section if any
    const withdrawnPlayers = allMembers.filter(m => withdrawals.includes(m.id));
    const withdrawnSection = withdrawnPlayers.length > 0 ? `
        <div class="roster-modal-section withdrawn-section">
            <h4>⚠️ Need Subs (${withdrawnPlayers.length})</h4>
            <div class="withdrawn-list">
                ${withdrawnPlayers.map(m => `<span class="withdrawn-player">${m.name}</span>`).join('')}
            </div>
        </div>
    ` : '';

    content.innerHTML = `
        <div class="roster-summary">
            <div class="roster-counter" id="rosterCounter">
                <span class="counter-label">Roster:</span>
                <span id="rosterCount" class="counter-value">${currentRoster.length}</span>/${teamSize}
            </div>
            <div class="subs-counter">
                <span class="counter-label">Backup Subs:</span>
                <span id="subsCount" class="counter-value">${currentSubs.length}</span>
            </div>
        </div>

        ${withdrawnSection}

        <div class="roster-modal-section">
            <h4>Available Players (${availablePlayers.length})</h4>
            <div class="roster-player-list">
                ${availablePlayers.map(m => renderPlayerCheckbox(m, true, 'available')).join('') || '<p class="empty-list">No one marked available yet</p>'}
            </div>
        </div>

        <div class="roster-modal-section">
            <h4>Other Members</h4>
            <div class="roster-player-list">
                ${otherPlayers.map(m => renderPlayerCheckbox(m, false, 'other')).join('') || '<p class="empty-list">None</p>'}
            </div>
        </div>

        <button class="btn btn-primary" onclick="saveRoster('${gameId}')" style="width: 100%; margin-top: 15px;">Save Roster</button>
    `;

    // Store game data for updates
    content.dataset.gameId = gameId;
    content.dataset.teamSize = teamSize;

    // Initialize temp arrays from current game state
    tempRoster = [...currentRoster];
    tempSubs = [...currentSubs];

    modal.classList.add('active');
}

// Temporary state for roster modal
let tempRoster = [];
let tempSubs = [];

function togglePlayerRoster(gameId, playerId, type) {
    const game = state.games.find(g => g.id === gameId);
    if (!game) return;

    const teamSize = game.teamSize || 10;

    // Initialize temp arrays from current game state if not set
    if (!tempRoster.length && !tempSubs.length) {
        tempRoster = [...(game.roster || [])];
        tempSubs = [...(game.subs || [])];
    }

    // Remove from both arrays first
    tempRoster = tempRoster.filter(id => id !== playerId);
    tempSubs = tempSubs.filter(id => id !== playerId);

    if (type === 'roster') {
        // Check if already at limit
        if (tempRoster.length >= teamSize) {
            showError(`Roster is full (${teamSize} players max). Remove someone first or add as backup sub.`);
            return;
        }
        // Check if was already on roster (toggle off)
        if (game.roster?.includes(playerId) && !tempRoster.includes(playerId)) {
            // Already removed above, do nothing
        } else {
            tempRoster.push(playerId);
        }
    } else if (type === 'sub') {
        // Check if was already a sub (toggle off)
        if (game.subs?.includes(playerId) && !tempSubs.includes(playerId)) {
            // Already removed above, do nothing
        } else {
            tempSubs.push(playerId);
        }
    }

    // Update UI
    updateRosterModalUI(gameId);
}

function updateRosterModalUI(gameId) {
    const game = state.games.find(g => g.id === gameId);
    if (!game) return;

    // Update counters
    const rosterCountEl = document.getElementById('rosterCount');
    const subsCountEl = document.getElementById('subsCount');
    const rosterCounter = document.getElementById('rosterCounter');

    if (rosterCountEl) rosterCountEl.textContent = tempRoster.length;
    if (subsCountEl) subsCountEl.textContent = tempSubs.length;

    // Update counter styling
    const teamSize = game.teamSize || 10;
    if (rosterCounter) {
        rosterCounter.classList.remove('roster-full', 'roster-over');
        if (tempRoster.length === teamSize) {
            rosterCounter.classList.add('roster-full');
        } else if (tempRoster.length > teamSize) {
            rosterCounter.classList.add('roster-over');
        }
    }

    // Update button states
    document.querySelectorAll('.roster-player-row').forEach(row => {
        const buttons = row.querySelectorAll('.roster-btn');
        const rosterBtn = buttons[0];
        const subBtn = buttons[1];

        if (rosterBtn && subBtn) {
            const playerId = rosterBtn.getAttribute('onclick').match(/'([^']+)'/g)[1].replace(/'/g, '');

            rosterBtn.classList.toggle('active', tempRoster.includes(playerId));
            subBtn.classList.toggle('active', tempSubs.includes(playerId));
        }
    });
}

// updateRosterCount is now handled by updateRosterModalUI

async function saveRoster(gameId) {
    const game = state.games.find(g => g.id === gameId);
    if (!game) return;

    // Use temp arrays if they have been modified, otherwise use current game state
    const roster = tempRoster.length > 0 || tempSubs.length > 0 ? tempRoster : (game.roster || []);
    const subs = tempRoster.length > 0 || tempSubs.length > 0 ? tempSubs : (game.subs || []);

    // Enforce team size
    const teamSize = game.teamSize || 10;
    if (roster.length > teamSize) {
        showError(`Roster cannot exceed ${teamSize} players`);
        return;
    }

    const result = await updateRosterAPI(gameId, roster, subs);
    if (result) {
        game.roster = result.roster;
        game.subs = result.subs;
        document.getElementById('rosterModal').classList.remove('active');
        // Clear temp arrays
        tempRoster = [];
        tempSubs = [];
        renderAll();
    }
}

async function saveWebhookSetting() {
    const input = document.getElementById('discordWebhook');
    const success = await saveWebhookAPI(input.value);
    if (success) {
        alert('Webhook saved!');
    }
}

// ==================== DISCORD TEST FUNCTIONS ====================

async function testDiscordDM() {
    const btn = document.getElementById('testDmBtn');
    const resultDiv = document.getElementById('testResult');

    btn.disabled = true;
    btn.textContent = 'Sending...';

    try {
        const response = await fetch('/api/test/dm', {
            method: 'POST',
            credentials: 'include'
        });

        const data = await response.json();

        if (response.ok) {
            showTestResult('success', 'Test DM sent! Check your Discord direct messages.');
        } else {
            showTestResult('error', data.error || 'Failed to send test DM');
        }
    } catch (error) {
        showTestResult('error', 'Network error: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.textContent = 'Send Test DM';
    }
}

async function testDiscordWebhook() {
    const btn = document.getElementById('testWebhookBtn');

    btn.disabled = true;
    btn.textContent = 'Posting...';

    try {
        const response = await fetch('/api/test/webhook', {
            method: 'POST',
            credentials: 'include'
        });

        const data = await response.json();

        if (response.ok) {
            showTestResult('success', 'Test message posted! Check your Discord channel.');
        } else {
            showTestResult('error', data.error || 'Failed to post test message');
        }
    } catch (error) {
        showTestResult('error', 'Network error: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.textContent = 'Send Test Post';
    }
}

async function testSubNotification() {
    const btn = document.getElementById('testSubBtn');

    btn.disabled = true;
    btn.textContent = 'Sending...';

    try {
        const response = await fetch('/api/test/sub-notification', {
            method: 'POST',
            credentials: 'include'
        });

        const data = await response.json();

        if (response.ok) {
            const count = data.sentCount || 0;
            showTestResult('success', `Sub alert test sent to ${count} manager(s). Check your Discord DMs.`);
        } else {
            showTestResult('error', data.error || 'Failed to send sub notification');
        }
    } catch (error) {
        showTestResult('error', 'Network error: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.textContent = 'Test Sub Alert';
    }
}

function showTestResult(type, message) {
    const resultDiv = document.getElementById('testResult');
    if (!resultDiv) return;

    resultDiv.style.display = 'block';
    resultDiv.className = `test-result ${type}`;
    resultDiv.innerHTML = `
        <span class="test-result-icon">${type === 'success' ? '✓' : '✕'}</span>
        <span class="test-result-message">${message}</span>
    `;

    // Auto-hide after 5 seconds
    setTimeout(() => {
        resultDiv.style.display = 'none';
    }, 5000);
}

// ==================== ROSTER MANAGEMENT ====================

function renderManageRoster() {
    const activeContainer = document.getElementById('manageActiveRoster');
    const subContainer = document.getElementById('manageSubRoster');

    if (!activeContainer || !subContainer) return;

    // Sort alphabetically by name
    const sortAlpha = (a, b) => a.name.localeCompare(b.name);

    const activeMembersList = allMembers.filter(m => !m.isVet).sort(sortAlpha);
    const vetMembersList = allMembers.filter(m => m.isVet).sort(sortAlpha);

    activeContainer.innerHTML = activeMembersList.map(m => createRosterItemHTML(m)).join('') || '<p style="color: var(--text-secondary); padding: 10px;">No active members</p>';
    subContainer.innerHTML = vetMembersList.map(m => createRosterItemHTML(m)).join('') || '<p style="color: var(--text-secondary); padding: 10px;">No vets</p>';

    // Initialize drag and drop
    initDragAndDrop();
}

function createRosterItemHTML(member) {
    const regionTag = member.region ? `<span class="roster-item-region">${member.region}</span>` : '';
    return `
        <div class="roster-item" data-id="${member.id}" draggable="true">
            <div class="roster-item-info">
                <span class="roster-item-name">${member.name}</span>
                ${regionTag}
            </div>
            <div class="roster-item-actions">
                <button class="move-btn" onclick="moveMember('${member.id}')" title="Move to ${member.isVet ? 'Active' : 'Vets'}">
                    ${member.isVet ? '↑' : '↓'}
                </button>
                <button onclick="removeMember('${member.id}')" title="Remove">✕</button>
            </div>
        </div>
    `;
}

function initDragAndDrop() {
    const containers = document.querySelectorAll('.sortable-roster');
    const items = document.querySelectorAll('.roster-item');

    items.forEach(item => {
        item.addEventListener('dragstart', handleDragStart);
        item.addEventListener('dragend', handleDragEnd);
    });

    containers.forEach(container => {
        container.addEventListener('dragover', handleDragOver);
        container.addEventListener('drop', handleDrop);
    });
}

let draggedItem = null;

function handleDragStart(e) {
    draggedItem = this;
    this.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
}

function handleDragEnd(e) {
    this.classList.remove('dragging');
    draggedItem = null;
}

function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';

    const afterElement = getDragAfterElement(this, e.clientY);
    if (draggedItem) {
        if (afterElement) {
            this.insertBefore(draggedItem, afterElement);
        } else {
            this.appendChild(draggedItem);
        }
    }
}

function getDragAfterElement(container, y) {
    const draggableElements = [...container.querySelectorAll('.roster-item:not(.dragging)')];

    return draggableElements.reduce((closest, child) => {
        const box = child.getBoundingClientRect();
        const offset = y - box.top - box.height / 2;
        if (offset < 0 && offset > closest.offset) {
            return { offset: offset, element: child };
        } else {
            return closest;
        }
    }, { offset: Number.NEGATIVE_INFINITY }).element;
}

async function handleDrop(e) {
    e.preventDefault();
    if (!draggedItem) return;

    const memberId = draggedItem.dataset.id;
    const targetContainer = e.currentTarget;
    const isVet = targetContainer.id === 'manageSubRoster';

    // Get new order
    const items = targetContainer.querySelectorAll('.roster-item');
    const newOrder = Array.from(items).map(item => item.dataset.id);

    // Update member type if moved between lists
    await updateMemberAPI(memberId, { isVet });

    // Save new order
    await saveRosterOrderAPI(isVet ? 'subs' : 'active', newOrder);

    // Refresh data
    await fetchData();
    renderManageRoster();
}

async function addNewMember() {
    const nameInput = document.getElementById('newMemberName');

    const name = nameInput.value.trim();
    if (!name) {
        alert('Please enter a player name');
        return;
    }

    // Default to Active NA member - manager can change via checkboxes
    const member = {
        name: name,
        isVet: false,
        region: 'NA'
    };

    const result = await addMemberAPI(member);
    if (result) {
        nameInput.value = '';
        await fetchData();
        renderAll();
    }
}

async function removeMember(memberId) {
    const member = allMembers.find(m => m.id === memberId);
    if (!member) return;

    if (!confirm(`Remove ${member.name} from the roster?`)) return;

    const success = await removeMemberAPI(memberId);
    if (success) {
        await fetchData();
        renderAll();
        renderManageRoster();
    }
}

async function moveMember(memberId) {
    const member = allMembers.find(m => m.id === memberId);
    if (!member) return;

    await updateMemberAPI(memberId, { isVet: !member.isVet });
    await fetchData();
    renderAll();
    renderManageRoster();
}

// API functions for roster management
async function addMemberAPI(member) {
    try {
        const response = await fetch('/api/members', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(member)
        });
        if (!response.ok) throw new Error('Failed to add member');
        return await response.json();
    } catch (error) {
        console.error('Error adding member:', error);
        alert('Failed to add member');
        return null;
    }
}

async function updateMemberAPI(memberId, updates) {
    try {
        const response = await fetch(`/api/members/${memberId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(updates)
        });
        if (!response.ok) throw new Error('Failed to update member');
        return await response.json();
    } catch (error) {
        console.error('Error updating member:', error);
        return null;
    }
}

async function removeMemberAPI(memberId) {
    try {
        const response = await fetch(`/api/members/${memberId}`, {
            method: 'DELETE',
            credentials: 'include'
        });
        if (!response.ok) throw new Error('Failed to remove member');
        return true;
    } catch (error) {
        console.error('Error removing member:', error);
        alert('Failed to remove member');
        return false;
    }
}

async function saveRosterOrderAPI(type, order) {
    try {
        const response = await fetch(`/api/members/order`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ type, order })
        });
        if (!response.ok) throw new Error('Failed to save order');
        return true;
    } catch (error) {
        console.error('Error saving roster order:', error);
        return false;
    }
}

async function updateMemberRegion(memberId, region) {
    const result = await updateMemberAPI(memberId, { region });
    if (result) {
        // Refresh members list and re-render
        const members = await fetchMembers();
        if (members.active) activeMembers = members.active;
        if (members.subs) subMembers = members.subs;
        allMembers = [
            ...activeMembers.map(m => ({ ...m, type: 'active' })),
            ...subMembers.map(m => ({ ...m, type: 'sub' }))
        ];
        renderRoster();
    }
}

async function updateMemberStatus(memberId, isVet) {
    const result = await updateMemberAPI(memberId, { isVet: isVet });
    if (result) {
        // Refresh members list and re-render
        const members = await fetchMembers();
        if (members.active) activeMembers = members.active;
        if (members.subs) subMembers = members.subs;
        allMembers = [
            ...activeMembers.map(m => ({ ...m, type: 'active' })),
            ...subMembers.map(m => ({ ...m, type: 'sub' }))
        ];
        renderRoster();
    }
}

// ==================== ET CLOCK & COUNTDOWN ====================

function startETClock() {
    updateETClock();
    setInterval(updateETClock, 1000);
}

function updateETClock() {
    const clockEl = document.getElementById('etClock');
    if (!clockEl) return;

    const now = new Date();
    const etTime = now.toLocaleString('en-US', {
        timeZone: 'America/New_York',
        hour: 'numeric',
        minute: '2-digit',
        hour12: true
    });
    clockEl.textContent = etTime;
}

function getGameCountdown(dateStr, timeStr) {
    const gameDateTime = new Date(`${dateStr}T${timeStr}:00`);
    // Adjust for ET timezone
    const etOffset = getETOffset();
    gameDateTime.setHours(gameDateTime.getHours() - etOffset);

    const now = new Date();
    const diff = gameDateTime - now;

    if (diff <= 0) {
        return { text: 'Game time!', urgent: true };
    }

    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    if (days > 0) {
        return { text: `${days}d ${hours}h`, urgent: false };
    } else if (hours > 0) {
        return { text: `${hours}h ${minutes}m`, urgent: hours < 2 };
    } else {
        return { text: `${minutes}m`, urgent: true };
    }
}

function getETOffset() {
    // Get current ET offset from UTC (handles DST)
    const now = new Date();
    const utc = now.getTime() + (now.getTimezoneOffset() * 60000);
    const et = new Date(utc + (-5 * 3600000)); // EST is UTC-5
    // Check if DST
    const jan = new Date(now.getFullYear(), 0, 1);
    const jul = new Date(now.getFullYear(), 6, 1);
    const stdOffset = Math.max(jan.getTimezoneOffset(), jul.getTimezoneOffset());
    const isDST = now.getTimezoneOffset() < stdOffset;
    return isDST ? -4 : -5; // EDT is UTC-4, EST is UTC-5
}

// ==================== MY ACCOUNT ====================

function showMyAccount() {
    const modal = document.getElementById('accountModal');
    if (!modal || !state.user) return;

    // Set current status
    const member = allMembers.find(m => m.id === state.currentPlayer);
    const isVet = member?.type === 'sub';

    document.getElementById('statusActive').checked = !isVet;
    document.getElementById('statusRetired').checked = isVet;

    // Load saved email/phone (from user object if available)
    document.getElementById('accountEmail').value = state.user.email || '';
    document.getElementById('accountPhone').value = state.user.phone || '';

    modal.classList.add('active');

    // Add auto-save listener for status toggle
    document.querySelectorAll('input[name="playerStatus"]').forEach(radio => {
        radio.addEventListener('change', handleStatusChange);
    });
}

async function handleStatusChange(e) {
    const retire = e.target.value === 'retired';
    if (state.currentPlayer) {
        const result = await updateMemberAPI(state.currentPlayer, { isVet: retire });
        if (result) {
            await fetchData();
            renderAll();
            // Show brief feedback
            const label = e.target.parentElement.querySelector('.toggle-label');
            if (label) {
                const original = label.textContent;
                label.textContent = 'Saved!';
                setTimeout(() => label.textContent = original, 1500);
            }
        }
    }
}

async function saveAccountSettings() {
    const email = document.getElementById('accountEmail').value.trim();
    const phone = document.getElementById('accountPhone').value.trim();

    // Save to user profile
    try {
        const response = await fetch(`${AUTH_BASE}/account`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ email, phone })
        });

        if (response.ok) {
            state.user.email = email;
            state.user.phone = phone;
            document.getElementById('accountModal').classList.remove('active');
            alert('Settings saved!');
        } else {
            showError('Failed to save settings');
        }
    } catch (error) {
        console.error('Error saving account settings:', error);
        showError('Failed to save settings');
    }
}

// ==================== FORM HELPERS ====================

function toggleCustomGameMode() {
    const select = document.getElementById('gameMode');
    const customInput = document.getElementById('gameModeCustom');
    if (select && customInput) {
        if (select.value === 'Other') {
            customInput.style.display = 'block';
            customInput.focus();
        } else {
            customInput.style.display = 'none';
            customInput.value = '';
        }
    }
}

// ==================== HELP TAB ====================

function initHelpToggle() {
    const buttons = document.querySelectorAll('.help-btn');
    buttons.forEach(btn => {
        btn.addEventListener('click', () => {
            buttons.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            const view = btn.dataset.view;
            document.getElementById('playerGuide').style.display = view === 'player' ? 'block' : 'none';
            document.getElementById('managerGuide').style.display = view === 'manager' ? 'block' : 'none';
        });
    });
}

// ==================== EVENT HANDLERS ====================

function initEventHandlers() {
    // Tab switching
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;

            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

            tab.classList.add('active');
            document.getElementById(tabName)?.classList.add('active');
        });
    });

    // Create game form
    const createForm = document.getElementById('createGameForm');
    if (createForm) {
        createForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            // Handle custom game mode
            let gameMode = document.getElementById('gameMode')?.value || 'War';
            if (gameMode === 'Other') {
                const customMode = document.getElementById('gameModeCustom')?.value?.trim();
                if (customMode) {
                    gameMode = customMode;
                } else {
                    showError('Please enter a game mode');
                    return;
                }
            }

            const gameData = {
                date: document.getElementById('gameDate').value,
                time: document.getElementById('gameTime').value,
                opponent: document.getElementById('opponent').value,
                league: document.getElementById('gameLeague')?.value || '',
                division: document.getElementById('gameDivision')?.value || '',
                gameMode: gameMode,
                teamSize: parseInt(document.getElementById('teamSize')?.value) || 10,
                notes: document.getElementById('gameNotes').value
            };

            let result;
            if (state.editingGameId) {
                // Update existing game
                result = await updateGame(state.editingGameId, gameData);
                if (result) {
                    const index = state.games.findIndex(g => g.id === state.editingGameId);
                    if (index !== -1) {
                        state.games[index] = result;
                    }
                    state.editingGameId = null;
                    updateFormMode(false);
                }
            } else {
                // Create new game
                result = await createGame(gameData);
                if (result) {
                    state.games.push(result);
                }
            }

            if (result) {
                createForm.reset();
                // Reset team size to default and hide custom input
                if (document.getElementById('teamSize')) {
                    document.getElementById('teamSize').value = '10';
                }
                if (document.getElementById('gameModeCustom')) {
                    document.getElementById('gameModeCustom').style.display = 'none';
                    document.getElementById('gameModeCustom').value = '';
                }
                renderAll();
            }
        });
    }

    // Logout button
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }

    // Link player button
    const linkBtn = document.getElementById('linkPlayerBtn');
    if (linkBtn) {
        linkBtn.addEventListener('click', () => {
            const select = document.getElementById('linkPlayerSelect');
            if (select.value) {
                linkPlayer(select.value);
            } else {
                showError('Please select your player name');
            }
        });
    }

    // Role preference change
    const roleSelect = document.getElementById('rolePreference');
    if (roleSelect) {
        roleSelect.addEventListener('change', async () => {
            if (state.currentPlayer) {
                const result = await setPlayerPreferenceAPI(state.currentPlayer, roleSelect.value);
                if (result) {
                    state.playerPreferences[state.currentPlayer] = roleSelect.value;
                    renderRoster();
                }
            }
        });
    }

    // Modal close buttons
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.addEventListener('click', () => {
            btn.closest('.modal').classList.remove('active');
            // Clear temp roster arrays when closing roster modal
            if (btn.closest('#rosterModal')) {
                tempRoster = [];
                tempSubs = [];
            }
        });
    });

    // Close modals on outside click
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
                // Clear temp roster arrays when closing roster modal
                if (modal.id === 'rosterModal') {
                    tempRoster = [];
                    tempSubs = [];
                }
            }
        });
    });

    // Privacy Policy links
    document.querySelectorAll('.privacy-link, #privacyLink').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('privacyModal').classList.add('active');
        });
    });

    // Terms of Service links
    document.querySelectorAll('.tos-link, #tosLink').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('tosModal').classList.add('active');
        });
    });

    // Help toggle
    initHelpToggle();
}

// ==================== INITIALIZATION ====================

async function init() {
    // Start the ET clock
    startETClock();

    // Check authentication
    const isAuthenticated = await checkAuth();
    updateAuthUI();

    // Check for auth errors in URL
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('error')) {
        showError('Login failed. Please try again.');
        window.history.replaceState({}, '', window.location.pathname);
    }

    // If authenticated but no player linked, show link modal
    if (isAuthenticated && !state.currentPlayer) {
        setTimeout(() => showLinkModal(), 500);
    }

    // Fetch data and render
    await fetchData();
    await fetchLeagues();
    await fetchDivisions();

    // Fetch members from API (updates isVet status etc.)
    const members = await fetchMembers();
    if (members.active) activeMembers = members.active;
    if (members.subs) subMembers = members.subs;
    // Rebuild allMembers with updated data
    allMembers = [
        ...activeMembers.map(m => ({ ...m, type: 'active' })),
        ...subMembers.map(m => ({ ...m, type: 'sub' }))
    ];

    updateAuthUI(); // Update again after preferences are loaded
    renderAll();

    // Set up event handlers
    initEventHandlers();

    // Auto-refresh every 30 seconds (also updates countdown timers)
    setInterval(async () => {
        await fetchData();
        renderAll();
    }, 30000);

    // Update countdowns every minute
    setInterval(() => {
        renderGames();
    }, 60000);
}

// Start the app
document.addEventListener('DOMContentLoaded', init);
