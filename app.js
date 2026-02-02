// Game Over Pop1 War Team Calendar - Frontend App
// Uses backend API with Discord OAuth

// ==================== CONFIG ====================
const API_BASE = window.location.origin + '/api';
const AUTH_BASE = window.location.origin + '/auth';

// ==================== MEMBER DATA ====================

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

const allMembers = [
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
                playerId: data.playerId
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
        return data;
    } catch (error) {
        console.error('Failed to fetch data:', error);
        showError('Failed to load data from server');
        return null;
    } finally {
        state.loading = false;
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

async function updateRosterAPI(gameId, roster) {
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/roster`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ roster })
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

async function postToDiscordAPI(gameId) {
    try {
        const response = await fetch(`${API_BASE}/discord/post/${gameId}`, {
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

// ==================== UI FUNCTIONS ====================

function updateAuthUI() {
    const loginSection = document.getElementById('loginSection');
    const userSection = document.getElementById('userSection');
    const managerTab = document.querySelector('.tab[data-tab="manage"]');

    if (state.user) {
        loginSection.style.display = 'none';
        userSection.style.display = 'flex';

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
        managerTab.style.display = state.isManager ? 'inline-block' : 'none';

        // Show linked player or prompt to link
        if (state.currentPlayer) {
            const member = allMembers.find(m => m.id === state.currentPlayer);
            linkedPlayer.textContent = member ? `Playing as: ${member.name}` : '';
            playerPreference.style.display = 'flex';

            // Set current preference
            const roleSelect = document.getElementById('rolePreference');
            roleSelect.value = state.playerPreferences[state.currentPlayer] || 'starter';
        } else {
            linkedPlayer.innerHTML = '<a href="#" onclick="showLinkModal(); return false;">Link your player profile</a>';
            playerPreference.style.display = 'none';
        }
    } else {
        loginSection.style.display = 'flex';
        userSection.style.display = 'none';
        managerTab.style.display = 'none';
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

    const availableCount = game.available?.length || 0;
    const rosterCount = game.roster?.length || 0;

    let availabilityButtons = '';
    if (state.user && state.currentPlayer) {
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

    // Roster list
    const rosterPlayers = (game.roster || []).map(id => {
        return `<span class="roster-chip">${getMemberName(id)}</span>`;
    }).join('');

    return `
        <div class="game-card ${isOnRoster ? 'on-roster' : ''}">
            <div class="game-header">
                <div class="game-date">${formatDate(game.date)}</div>
                <div class="game-time">${formatTime(game.time)}</div>
            </div>
            <div class="game-opponent">vs ${game.opponent}</div>
            ${game.notes ? `<div class="game-notes">${game.notes}</div>` : ''}

            <div class="game-status">
                <span class="status-badge">Roster: ${rosterCount}/10</span>
                <span class="status-badge available-badge">${availableCount} Available</span>
            </div>

            ${availabilityButtons}

            ${availablePlayers ? `
                <div class="available-section">
                    <h4>Available Players</h4>
                    <div class="player-chips">${availablePlayers}</div>
                </div>
            ` : ''}

            ${rosterPlayers ? `
                <div class="roster-section">
                    <h4>Final Roster</h4>
                    <div class="roster-chips">${rosterPlayers}</div>
                </div>
            ` : ''}

            ${state.isManager ? `
                <div class="manager-actions">
                    <button class="btn btn-secondary" onclick="openRosterModal('${game.id}')">Select Roster</button>
                    <button class="btn btn-discord-small" onclick="postToDiscord('${game.id}')">Post to Discord</button>
                </div>
            ` : ''}
        </div>
    `;
}

function renderRoster() {
    const activeContainer = document.getElementById('activeRoster');
    const subContainer = document.getElementById('subRoster');

    if (activeContainer) {
        activeContainer.innerHTML = activeMembers.map(m => renderMemberCard(m)).join('');
    }
    if (subContainer) {
        subContainer.innerHTML = subMembers.map(m => renderMemberCard(m, true)).join('');
    }
}

function renderMemberCard(member, isSub = false) {
    const pref = state.playerPreferences[member.id];
    const prefLabel = pref === 'sub' ? 'Prefers Sub' : 'Starter';

    return `
        <div class="member-card ${isSub ? 'sub-member' : ''}">
            <div class="member-name">${member.name}</div>
            <div class="member-tags">
                ${member.region ? `<span class="member-tag region">${member.region}</span>` : ''}
                ${member.note ? `<span class="member-tag note">${member.note}</span>` : ''}
                <span class="member-tag pref ${pref || 'starter'}">${prefLabel}</span>
            </div>
            <div class="member-year">Since ${member.year}</div>
        </div>
    `;
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
                ${game.notes ? `<br><em>${game.notes}</em>` : ''}
            </div>
            <div class="game-actions">
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

async function postToDiscord(gameId) {
    try {
        await postToDiscordAPI(gameId);
        alert('Posted to Discord successfully!');
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

    const renderPlayerCheckbox = (member, isAvailable, isSub = false) => {
        const isSelected = currentRoster.includes(member.id);
        const pref = state.playerPreferences[member.id];
        return `
            <label class="roster-checkbox ${isSelected ? 'selected' : ''} ${isSub ? 'sub' : ''}">
                <input type="checkbox" value="${member.id}" ${isSelected ? 'checked' : ''}>
                <span>${member.name}</span>
                ${member.region ? `<span class="tag">[${member.region}]</span>` : ''}
                ${pref === 'sub' ? '<span class="tag">(sub)</span>' : ''}
                ${isAvailable ? '<span class="tag available">Available</span>' : ''}
            </label>
        `;
    };

    const availableActive = activeMembers.filter(m => game.available?.includes(m.id));
    const availableSubs = subMembers.filter(m => game.available?.includes(m.id));
    const otherActive = activeMembers.filter(m => !game.available?.includes(m.id));
    const otherSubs = subMembers.filter(m => !game.available?.includes(m.id));

    content.innerHTML = `
        <div class="roster-counter">Selected: <span id="rosterCount">${currentRoster.length}</span>/10</div>

        <div class="roster-section">
            <h4>Available Starters</h4>
            <div class="roster-list">
                ${availableActive.map(m => renderPlayerCheckbox(m, true)).join('')}
            </div>
        </div>

        <div class="roster-section">
            <h4>Available Subs</h4>
            <div class="roster-list">
                ${availableSubs.map(m => renderPlayerCheckbox(m, true, true)).join('')}
            </div>
        </div>

        <div class="roster-section">
            <h4>Other Members (not confirmed)</h4>
            <div class="roster-list">
                ${otherActive.map(m => renderPlayerCheckbox(m, false)).join('')}
                ${otherSubs.map(m => renderPlayerCheckbox(m, false, true)).join('')}
            </div>
        </div>

        <button class="btn btn-primary" onclick="saveRoster('${gameId}')">Save Roster</button>
    `;

    // Add change listener for counter
    content.querySelectorAll('input[type="checkbox"]').forEach(cb => {
        cb.addEventListener('change', updateRosterCount);
    });

    modal.classList.add('active');
}

function updateRosterCount() {
    const count = document.querySelectorAll('#rosterModalContent input[type="checkbox"]:checked').length;
    const counter = document.getElementById('rosterCount');
    if (counter) counter.textContent = count;
}

async function saveRoster(gameId) {
    const checkboxes = document.querySelectorAll('#rosterModalContent input[type="checkbox"]:checked');
    const roster = Array.from(checkboxes).map(cb => cb.value);

    const result = await updateRosterAPI(gameId, roster);
    if (result) {
        const game = state.games.find(g => g.id === gameId);
        if (game) {
            game.roster = result.roster;
        }
        document.getElementById('rosterModal').classList.remove('active');
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
            const gameData = {
                date: document.getElementById('gameDate').value,
                time: document.getElementById('gameTime').value,
                opponent: document.getElementById('opponent').value,
                notes: document.getElementById('gameNotes').value
            };

            const game = await createGame(gameData);
            if (game) {
                state.games.push(game);
                createForm.reset();
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
        });
    });

    // Close modals on outside click
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
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
    renderAll();

    // Set up event handlers
    initEventHandlers();

    // Auto-refresh every 30 seconds
    setInterval(async () => {
        await fetchData();
        renderAll();
    }, 30000);
}

// Start the app
document.addEventListener('DOMContentLoaded', init);
