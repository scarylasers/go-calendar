package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// ==================== TYPES ====================

type Member struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Year   int    `json:"year"`
	Region string `json:"region,omitempty"`
	Note   string `json:"note,omitempty"`
	IsSub  bool   `json:"isSub,omitempty"`
}

type Game struct {
	ID          string   `json:"id"`
	Date        string   `json:"date"`
	Time        string   `json:"time"`
	Opponent    string   `json:"opponent"`
	Notes       string   `json:"notes"`
	Available   []string `json:"available"`
	Unavailable []string `json:"unavailable"`
	Roster      []string `json:"roster"`
	Reminded    bool     `json:"reminded"`
}

type User struct {
	DiscordID   string `json:"discordId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	PlayerID    string `json:"playerId"`
	IsManager   bool   `json:"isManager"`
}

type AllData struct {
	Games             []Game            `json:"games"`
	PlayerPreferences map[string]string `json:"playerPreferences"`
	DiscordWebhook    string            `json:"discordWebhook"`
}

type Session struct {
	DiscordID   string `json:"d"`
	Username    string `json:"u"`
	IsManager   bool   `json:"m"`
	PlayerID    string `json:"p"`
	ExpiresAt   int64  `json:"e"`
}

// ==================== GLOBALS ====================

var db *sql.DB

var (
	discordClientID     = os.Getenv("DISCORD_CLIENT_ID")
	discordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	discordBotToken     = os.Getenv("DISCORD_BOT_TOKEN")
	discordGuildID      = os.Getenv("DISCORD_GUILD_ID")
	discordManagerRole  = os.Getenv("DISCORD_MANAGER_ROLE")
	sessionSecret       = os.Getenv("SESSION_SECRET")
	baseURL             = os.Getenv("BASE_URL") // e.g., https://go-pop1-calendar.onrender.com
)

var ActiveMembers = []Member{
	{ID: "alock", Name: "GO-Alock4", Year: 2021},
	{ID: "cronides", Name: "GO-Cronides", Year: 2021, Note: "Former Head of GO"},
	{ID: "ghostxrp", Name: "GO_GhostXRP", Year: 2021},
	{ID: "bramrianne", Name: "GO_ BramRianne", Year: 2022, Region: "EU"},
	{ID: "thesean", Name: "GO_The.Sean", Year: 2023, Region: "EU", Note: "Former Head of GO Europe"},
	{ID: "babs", Name: "GO_BABs", Year: 2023},
	{ID: "honeyluvv", Name: "GO_HoneyLuvv", Year: 2023},
	{ID: "zoloto", Name: "GO_Zoloto", Year: 2023, Region: "EU"},
	{ID: "humanoid", Name: "GO_Humanoid", Year: 2023, Region: "EU", Note: "Head of GO Europe"},
	{ID: "jesshawk", Name: "GO_JessHawk3", Year: 2023, Note: "Head of GO"},
	{ID: "deathraider", Name: "GO_Deathraider255", Year: 2024, Region: "EU"},
	{ID: "pinkpwnage", Name: "GOxPinkPWNAGE5", Year: 2024},
	{ID: "sami", Name: "Sami_008", Year: 2024},
	{ID: "silverlining", Name: "GO_SilverLining23", Year: 2024, Region: "EU"},
	{ID: "colinthe5", Name: "Go_Colinthe5", Year: 2024, Note: "Head of GO League"},
	{ID: "thedon", Name: "GO_The_Don", Year: 2024},
	{ID: "headhelper", Name: "GO HeadHelper", Year: 2024},
	{ID: "cosmo", Name: "GO_Cosmo", Year: 2025},
	{ID: "fxyz", Name: "f(x,y,z)", Year: 2025},
	{ID: "amberloaf", Name: "AmberLoaf", Year: 2025},
	{ID: "scarylasers", Name: "GO_ScaryLasers", Year: 2025},
	{ID: "glotones", Name: "GLOTONES", Year: 2025},
	{ID: "smich", Name: "GO_SMICH1989", Year: 2025, Region: "EU"},
	{ID: "flexx", Name: "GO_FlexX", Year: 2025},
	{ID: "kygelli", Name: "GO_KyGelli", Year: 2025},
	{ID: "chr1sp", Name: "GO_Chr1sP", Year: 2025},
}

var SubMembers = []Member{
	{ID: "shark", Name: "Shark", Year: 2021, IsSub: true},
	{ID: "docbutler", Name: "GO_DocButler", Year: 2021, IsSub: true},
	{ID: "drsmartazz", Name: "GO_DrSmartAzz", Year: 2021, IsSub: true},
	{ID: "pbandc", Name: "PBandC-GO", Year: 2021, IsSub: true},
	{ID: "maverick", Name: "GO_Maverick", Year: 2021, IsSub: true},
	{ID: "honeygun", Name: "HoneyGUN", Year: 2022, IsSub: true},
	{ID: "loki", Name: "GO_Loki714", Year: 2022, IsSub: true},
	{ID: "lizlow", Name: "LizLow91", Year: 2023, IsSub: true},
	{ID: "kc", Name: "GO_KC", Year: 2023, IsSub: true},
	{ID: "lester", Name: "GO_Lester", Year: 2024, IsSub: true},
	{ID: "kingslayer", Name: "GO_Kingslayer", Year: 2024, Region: "EU", IsSub: true},
	{ID: "bacon", Name: "AllTheBaconAndEGGz", Year: 2025, IsSub: true},
	{ID: "stooobe", Name: "GO_STOOOBE", Year: 2025, Note: "Former Head of GO", IsSub: true},
	{ID: "johnharple", Name: "GO_JohnHarple", Year: 2025, IsSub: true},
}

// ==================== DATABASE ====================

func initDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set, running without database")
		return nil
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS games (
			id TEXT PRIMARY KEY,
			date TEXT NOT NULL,
			time TEXT NOT NULL,
			opponent TEXT NOT NULL,
			notes TEXT DEFAULT '',
			available TEXT DEFAULT '[]',
			unavailable TEXT DEFAULT '[]',
			roster TEXT DEFAULT '[]',
			reminded BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create games table: %v", err)
	}

	// Add reminded column if it doesn't exist
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS reminded BOOLEAN DEFAULT FALSE`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS player_preferences (
			player_id TEXT PRIMARY KEY,
			preference TEXT DEFAULT 'starter'
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create player_preferences table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create settings table: %v", err)
	}

	// Users table for Discord OAuth
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			discord_id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			display_name TEXT,
			avatar TEXT,
			player_id TEXT,
			is_manager BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// ==================== SESSION HELPERS ====================

func createSessionToken(session Session) (string, error) {
	if sessionSecret == "" {
		sessionSecret = "dev-secret-change-in-production"
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(sessionSecret))
	h.Write(data)
	sig := h.Sum(nil)

	// Encode as base64: data.signature
	encoded := base64.URLEncoding.EncodeToString(data) + "." + base64.URLEncoding.EncodeToString(sig)
	return encoded, nil
}

func parseSessionToken(token string) (*Session, error) {
	if sessionSecret == "" {
		sessionSecret = "dev-secret-change-in-production"
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	sig, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// Verify HMAC
	h := hmac.New(sha256.New, []byte(sessionSecret))
	h.Write(data)
	expectedSig := h.Sum(nil)

	if !hmac.Equal(sig, expectedSig) {
		return nil, fmt.Errorf("invalid signature")
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().Unix() > session.ExpiresAt {
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func getSessionFromRequest(r *http.Request) *Session {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}

	session, err := parseSessionToken(cookie.Value)
	if err != nil {
		return nil
	}

	return session
}

func setSessionCookie(w http.ResponseWriter, session Session) {
	token, err := createSessionToken(session)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})
}

// ==================== DISCORD API HELPERS ====================

func getDiscordOAuthURL() string {
	if baseURL == "" {
		baseURL = "https://go-pop1-calendar.onrender.com"
	}
	redirectURI := baseURL + "/auth/discord/callback"

	params := url.Values{}
	params.Set("client_id", discordClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "identify guilds.members.read")

	return "https://discord.com/api/oauth2/authorize?" + params.Encode()
}

func exchangeCodeForToken(code string) (string, error) {
	if baseURL == "" {
		baseURL = "https://go-pop1-calendar.onrender.com"
	}
	redirectURI := baseURL + "/auth/discord/callback"

	data := url.Values{}
	data.Set("client_id", discordClientID)
	data.Set("client_secret", discordClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	resp, err := http.PostForm("https://discord.com/api/oauth2/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("discord error: %s", result.Error)
	}

	return result.AccessToken, nil
}

func getDiscordUser(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func checkUserHasManagerRole(accessToken, userID string) bool {
	if discordGuildID == "" || discordManagerRole == "" {
		return false
	}

	// Get guild member info
	url := fmt.Sprintf("https://discord.com/api/users/@me/guilds/%s/member", discordGuildID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error fetching guild member: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	var member struct {
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return false
	}

	// We need to get guild roles to match role name to ID
	// Using bot token to get roles
	if discordBotToken == "" {
		return false
	}

	rolesURL := fmt.Sprintf("https://discord.com/api/guilds/%s/roles", discordGuildID)
	rolesReq, _ := http.NewRequest("GET", rolesURL, nil)
	rolesReq.Header.Set("Authorization", "Bot "+discordBotToken)

	rolesResp, err := http.DefaultClient.Do(rolesReq)
	if err != nil {
		return false
	}
	defer rolesResp.Body.Close()

	var roles []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(rolesResp.Body).Decode(&roles); err != nil {
		return false
	}

	// Find the manager role ID
	var managerRoleID string
	for _, role := range roles {
		if strings.EqualFold(role.Name, discordManagerRole) {
			managerRoleID = role.ID
			break
		}
	}

	if managerRoleID == "" {
		return false
	}

	// Check if user has the role
	for _, roleID := range member.Roles {
		if roleID == managerRoleID {
			return true
		}
	}

	return false
}

// ==================== USER DATABASE FUNCTIONS ====================

func getUserByDiscordID(discordID string) (*User, error) {
	if db == nil {
		return nil, nil
	}

	var u User
	var playerID, displayName, avatar sql.NullString
	err := db.QueryRow(`
		SELECT discord_id, username, display_name, avatar, player_id, is_manager
		FROM users WHERE discord_id = $1
	`, discordID).Scan(&u.DiscordID, &u.Username, &displayName, &avatar, &playerID, &u.IsManager)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.DisplayName = displayName.String
	u.Avatar = avatar.String
	u.PlayerID = playerID.String

	return &u, nil
}

func createOrUpdateUser(discordID, username, displayName, avatar string, isManager bool) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(`
		INSERT INTO users (discord_id, username, display_name, avatar, is_manager)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (discord_id) DO UPDATE SET
			username = $2,
			display_name = $3,
			avatar = $4,
			is_manager = $5
	`, discordID, username, displayName, avatar, isManager)

	return err
}

func linkUserToPlayer(discordID, playerID string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(`UPDATE users SET player_id = $1 WHERE discord_id = $2`, playerID, discordID)
	return err
}

func getLinkedPlayerIDs() ([]string, error) {
	if db == nil {
		return []string{}, nil
	}

	rows, err := db.Query(`SELECT player_id FROM users WHERE player_id IS NOT NULL AND player_id != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

func getUserByPlayerID(playerID string) (*User, error) {
	if db == nil {
		return nil, nil
	}

	var u User
	var pID, displayName, avatar sql.NullString
	err := db.QueryRow(`
		SELECT discord_id, username, display_name, avatar, player_id, is_manager
		FROM users WHERE player_id = $1
	`, playerID).Scan(&u.DiscordID, &u.Username, &displayName, &avatar, &pID, &u.IsManager)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.DisplayName = displayName.String
	u.Avatar = avatar.String
	u.PlayerID = pID.String

	return &u, nil
}

// ==================== HELPERS ====================

func getMemberName(memberID string) string {
	allMembers := append(ActiveMembers, SubMembers...)
	for _, m := range allMembers {
		if m.ID == memberID {
			return m.Name
		}
	}
	return memberID
}

func generateGameID() string {
	timestamp := time.Now().UnixMilli()
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 9)
	for i := range suffix {
		suffix[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("game_%d_%s", timestamp, string(suffix))
}

func parseJSONArray(s string) []string {
	if s == "" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return []string{}
	}
	return arr
}

func toJSONString(arr []string) string {
	if arr == nil {
		arr = []string{}
	}
	b, _ := json.Marshal(arr)
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// ==================== DATA FUNCTIONS ====================

func getAllGames() ([]Game, error) {
	if db == nil {
		return []Game{}, nil
	}

	rows, err := db.Query("SELECT id, date, time, opponent, notes, available, unavailable, roster, COALESCE(reminded, false) FROM games ORDER BY date, time")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var available, unavailable, roster string
		if err := rows.Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.Notes, &available, &unavailable, &roster, &g.Reminded); err != nil {
			return nil, err
		}
		g.Available = parseJSONArray(available)
		g.Unavailable = parseJSONArray(unavailable)
		g.Roster = parseJSONArray(roster)
		games = append(games, g)
	}

	if games == nil {
		games = []Game{}
	}
	return games, nil
}

func getGameByID(gameID string) (*Game, error) {
	if db == nil {
		return nil, nil
	}

	var g Game
	var available, unavailable, roster string
	err := db.QueryRow("SELECT id, date, time, opponent, notes, available, unavailable, roster, COALESCE(reminded, false) FROM games WHERE id = $1", gameID).
		Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.Notes, &available, &unavailable, &roster, &g.Reminded)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	g.Available = parseJSONArray(available)
	g.Unavailable = parseJSONArray(unavailable)
	g.Roster = parseJSONArray(roster)
	return &g, nil
}

func createGame(date, gameTime, opponent, notes string) (*Game, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	gameID := generateGameID()
	_, err := db.Exec(
		"INSERT INTO games (id, date, time, opponent, notes, available, unavailable, roster, reminded) VALUES ($1, $2, $3, $4, $5, '[]', '[]', '[]', false)",
		gameID, date, gameTime, opponent, notes,
	)
	if err != nil {
		return nil, err
	}

	return getGameByID(gameID)
}

func updateGame(gameID string, updates map[string]interface{}) (*Game, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	for key, value := range updates {
		var val interface{}
		switch v := value.(type) {
		case []string:
			val = toJSONString(v)
		case string:
			val = v
		case bool:
			val = v
		default:
			continue
		}
		_, err := db.Exec(fmt.Sprintf("UPDATE games SET %s = $1 WHERE id = $2", key), val, gameID)
		if err != nil {
			return nil, err
		}
	}

	return getGameByID(gameID)
}

func deleteGame(gameID string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec("DELETE FROM games WHERE id = $1", gameID)
	return err
}

func getAllPreferences() (map[string]string, error) {
	prefs := make(map[string]string)
	if db == nil {
		return prefs, nil
	}

	rows, err := db.Query("SELECT player_id, preference FROM player_preferences")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var playerID, pref string
		if err := rows.Scan(&playerID, &pref); err != nil {
			return nil, err
		}
		prefs[playerID] = pref
	}

	return prefs, nil
}

func setPreference(playerID, preference string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(`
		INSERT INTO player_preferences (player_id, preference)
		VALUES ($1, $2)
		ON CONFLICT (player_id) DO UPDATE SET preference = $2
	`, playerID, preference)
	return err
}

func getSetting(key string) (string, error) {
	if db == nil {
		return "", nil
	}

	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = $1", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func setSetting(key, value string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(`
		INSERT INTO settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2
	`, key, value)
	return err
}

// ==================== AUTH HANDLERS ====================

func handleDiscordLogin(w http.ResponseWriter, r *http.Request) {
	if discordClientID == "" {
		writeError(w, http.StatusServiceUnavailable, "Discord OAuth not configured")
		return
	}
	http.Redirect(w, r, getDiscordOAuthURL(), http.StatusTemporaryRedirect)
}

func handleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/?error=no_code", http.StatusTemporaryRedirect)
		return
	}

	// Exchange code for token
	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		http.Redirect(w, r, "/?error=token_failed", http.StatusTemporaryRedirect)
		return
	}

	// Get user info
	discordUser, err := getDiscordUser(accessToken)
	if err != nil {
		log.Printf("Get user error: %v", err)
		http.Redirect(w, r, "/?error=user_failed", http.StatusTemporaryRedirect)
		return
	}

	discordID := discordUser["id"].(string)
	username := discordUser["username"].(string)
	displayName := ""
	if dn, ok := discordUser["global_name"]; ok && dn != nil {
		displayName = dn.(string)
	}
	avatar := ""
	if av, ok := discordUser["avatar"]; ok && av != nil {
		avatar = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordID, av.(string))
	}

	// Check manager role
	isManager := checkUserHasManagerRole(accessToken, discordID)

	// Save/update user in database
	if err := createOrUpdateUser(discordID, username, displayName, avatar, isManager); err != nil {
		log.Printf("Save user error: %v", err)
	}

	// Get existing user to check if they have a player linked
	existingUser, _ := getUserByDiscordID(discordID)
	playerID := ""
	if existingUser != nil {
		playerID = existingUser.PlayerID
	}

	// Create session
	session := Session{
		DiscordID: discordID,
		Username:  username,
		IsManager: isManager,
		PlayerID:  playerID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	setSessionCookie(w, session)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleAuthMe(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Get full user from DB
	user, _ := getUserByDiscordID(session.DiscordID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"discordId":     session.DiscordID,
		"username":      session.Username,
		"isManager":     session.IsManager,
		"playerId":      session.PlayerID,
		"avatar":        func() string { if user != nil { return user.Avatar } return "" }(),
		"displayName":   func() string { if user != nil { return user.DisplayName } return "" }(),
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleLinkPlayer(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var body struct {
		PlayerID string `json:"playerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if body.PlayerID == "" {
		writeError(w, http.StatusBadRequest, "Player ID required")
		return
	}

	// Check if player is already linked
	existingUser, _ := getUserByPlayerID(body.PlayerID)
	if existingUser != nil && existingUser.DiscordID != session.DiscordID {
		writeError(w, http.StatusConflict, "This player is already linked to another account")
		return
	}

	// Link player
	if err := linkUserToPlayer(session.DiscordID, body.PlayerID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update session cookie with new player ID
	session.PlayerID = body.PlayerID
	setSessionCookie(w, *session)

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ==================== MEMBER HANDLERS ====================

func handleGetMembers(w http.ResponseWriter, r *http.Request) {
	linkedIDs, _ := getLinkedPlayerIDs()
	linkedMap := make(map[string]bool)
	for _, id := range linkedIDs {
		linkedMap[id] = true
	}

	type MemberWithStatus struct {
		Member
		Linked bool `json:"linked"`
	}

	var active []MemberWithStatus
	var subs []MemberWithStatus

	for _, m := range ActiveMembers {
		active = append(active, MemberWithStatus{Member: m, Linked: linkedMap[m.ID]})
	}
	for _, m := range SubMembers {
		subs = append(subs, MemberWithStatus{Member: m, Linked: linkedMap[m.ID]})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"active": active,
		"subs":   subs,
	})
}

// ==================== GAME HANDLERS ====================

func handleGetAllData(w http.ResponseWriter, r *http.Request) {
	games, err := getAllGames()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	prefs, err := getAllPreferences()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AllData{
		Games:             games,
		PlayerPreferences: prefs,
		DiscordWebhook:    "",
	})
}

func handleGetGames(w http.ResponseWriter, r *http.Request) {
	games, err := getAllGames()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, games)
}

func handleCreateGame(w http.ResponseWriter, r *http.Request) {
	// Check manager permission
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	var body struct {
		Date     string `json:"date"`
		Time     string `json:"time"`
		Opponent string `json:"opponent"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if body.Date == "" || body.Time == "" || body.Opponent == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	game, err := createGame(body.Date, body.Time, body.Opponent, body.Notes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, game)
}

func handleDeleteGame(w http.ResponseWriter, r *http.Request) {
	// Check manager permission
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	gameID := vars["id"]

	if err := deleteGame(gameID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleUpdateRoster(w http.ResponseWriter, r *http.Request) {
	// Check manager permission
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	gameID := vars["id"]

	game, err := getGameByID(gameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if game == nil {
		writeError(w, http.StatusNotFound, "Game not found")
		return
	}

	var body struct {
		Roster []string `json:"roster"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	updated, err := updateGame(gameID, map[string]interface{}{"roster": body.Roster})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func handleSetAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	game, err := getGameByID(gameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if game == nil {
		writeError(w, http.StatusNotFound, "Game not found")
		return
	}

	var body struct {
		PlayerID    string `json:"playerId"`
		IsAvailable bool   `json:"isAvailable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if body.PlayerID == "" {
		writeError(w, http.StatusBadRequest, "Player ID required")
		return
	}

	// Optional: verify session matches player ID
	session := getSessionFromRequest(r)
	if session != nil && session.PlayerID != "" && session.PlayerID != body.PlayerID {
		writeError(w, http.StatusForbidden, "Can only set your own availability")
		return
	}

	available := make([]string, 0)
	unavailable := make([]string, 0)
	for _, p := range game.Available {
		if p != body.PlayerID {
			available = append(available, p)
		}
	}
	for _, p := range game.Unavailable {
		if p != body.PlayerID {
			unavailable = append(unavailable, p)
		}
	}

	if body.IsAvailable {
		available = append(available, body.PlayerID)
	} else {
		unavailable = append(unavailable, body.PlayerID)
	}

	updated, err := updateGame(gameID, map[string]interface{}{
		"available":   available,
		"unavailable": unavailable,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func handleGetPreferences(w http.ResponseWriter, r *http.Request) {
	prefs, err := getAllPreferences()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

func handleSetPreference(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["playerId"]

	var body struct {
		Preference string `json:"preference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if body.Preference != "starter" && body.Preference != "sub" {
		writeError(w, http.StatusBadRequest, "Invalid preference. Must be 'starter' or 'sub'")
		return
	}

	// Optional: verify session matches player ID
	session := getSessionFromRequest(r)
	if session != nil && session.PlayerID != "" && session.PlayerID != playerID {
		writeError(w, http.StatusForbidden, "Can only set your own preference")
		return
	}

	if err := setPreference(playerID, body.Preference); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"player_id":  playerID,
		"preference": body.Preference,
	})
}

func handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	webhook, _ := getSetting("discord_webhook")
	response := map[string]interface{}{
		"configured": webhook != "",
	}
	if webhook != "" && len(webhook) > 10 {
		response["preview"] = "****" + webhook[len(webhook)-10:]
	}
	writeJSON(w, http.StatusOK, response)
}

func handleSetWebhook(w http.ResponseWriter, r *http.Request) {
	// Check manager permission
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	var body struct {
		Webhook string `json:"webhook"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := setSetting("discord_webhook", body.Webhook); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handlePostToDiscord(w http.ResponseWriter, r *http.Request) {
	// Check manager permission
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	gameID := vars["id"]

	webhook, _ := getSetting("discord_webhook")
	if webhook == "" {
		writeError(w, http.StatusBadRequest, "Discord webhook not configured")
		return
	}

	game, err := getGameByID(gameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if game == nil {
		writeError(w, http.StatusNotFound, "Game not found")
		return
	}

	formattedDate := game.Date
	if t, err := time.Parse("2006-01-02", game.Date); err == nil {
		formattedDate = t.Format("Monday, Jan 02, 2006")
	}

	formattedTime := game.Time
	if parts := strings.Split(game.Time, ":"); len(parts) >= 2 {
		if hour, err := strconv.Atoi(parts[0]); err == nil {
			ampm := "AM"
			if hour >= 12 {
				ampm = "PM"
			}
			hour12 := hour % 12
			if hour12 == 0 {
				hour12 = 12
			}
			formattedTime = fmt.Sprintf("%d:%s %s ET", hour12, parts[1], ampm)
		}
	}

	var rosterNames []string
	for _, pid := range game.Roster {
		rosterNames = append(rosterNames, getMemberName(pid))
	}

	rosterValue := "TBD"
	if len(rosterNames) > 0 {
		rosterValue = strings.Join(rosterNames, "\n")
	}

	embed := map[string]interface{}{
		"title": fmt.Sprintf("🎮 Game Day: %s", formattedDate),
		"color": 0x00f0ff,
		"fields": []map[string]interface{}{
			{"name": "⏰ Time", "value": formattedTime, "inline": true},
			{"name": "⚔️ Opponent", "value": game.Opponent, "inline": true},
			{"name": fmt.Sprintf("👥 Roster (%d/10)", len(rosterNames)), "value": rosterValue, "inline": false},
		},
		"footer":    map[string]string{"text": "Game Over Pop1 War Team"},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if game.Notes != "" {
		fields := embed["fields"].([]map[string]interface{})
		fields = append(fields, map[string]interface{}{
			"name":   "📝 Notes",
			"value":  game.Notes,
			"inline": false,
		})
		embed["fields"] = fields
	}

	payload := map[string]interface{}{
		"username": "Game Over Bot",
		"embeds":   []map[string]interface{}{embed},
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(webhook, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to post to Discord: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		writeError(w, http.StatusInternalServerError, "Discord API error: "+string(body))
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ==================== REMINDER HELPERS (for cron job) ====================

func handleGetPendingReminders(w http.ResponseWriter, r *http.Request) {
	// This endpoint is for the cron job - should be called internally
	if db == nil {
		writeJSON(w, http.StatusOK, []Game{})
		return
	}

	// Get tomorrow's date
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	rows, err := db.Query(`
		SELECT id, date, time, opponent, notes, available, unavailable, roster, COALESCE(reminded, false)
		FROM games
		WHERE date = $1 AND (reminded = false OR reminded IS NULL) AND roster != '[]'
	`, tomorrow)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var available, unavailable, roster string
		if err := rows.Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.Notes, &available, &unavailable, &roster, &g.Reminded); err != nil {
			continue
		}
		g.Available = parseJSONArray(available)
		g.Unavailable = parseJSONArray(unavailable)
		g.Roster = parseJSONArray(roster)
		games = append(games, g)
	}

	if games == nil {
		games = []Game{}
	}

	writeJSON(w, http.StatusOK, games)
}

func handleMarkReminded(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	if _, err := updateGame(gameID, map[string]interface{}{"reminded": true}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleGetUserDiscordID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["playerId"]

	user, err := getUserByPlayerID(playerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "No Discord account linked to this player")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"discordId": user.DiscordID})
}

// ==================== MAIN ====================

func main() {
	rand.Seed(time.Now().UnixNano())

	// Re-read env vars in case they weren't set at init
	discordClientID = os.Getenv("DISCORD_CLIENT_ID")
	discordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	discordGuildID = os.Getenv("DISCORD_GUILD_ID")
	discordManagerRole = os.Getenv("DISCORD_MANAGER_ROLE")
	sessionSecret = os.Getenv("SESSION_SECRET")
	baseURL = os.Getenv("BASE_URL")

	if err := initDB(); err != nil {
		log.Printf("Warning: %v", err)
	}

	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/auth/discord", handleDiscordLogin).Methods("GET")
	r.HandleFunc("/auth/discord/callback", handleDiscordCallback).Methods("GET")
	r.HandleFunc("/auth/me", handleAuthMe).Methods("GET")
	r.HandleFunc("/auth/logout", handleLogout).Methods("POST")
	r.HandleFunc("/auth/link", handleLinkPlayer).Methods("POST")

	// API routes
	r.HandleFunc("/api/data", handleGetAllData).Methods("GET")
	r.HandleFunc("/api/members", handleGetMembers).Methods("GET")
	r.HandleFunc("/api/games", handleGetGames).Methods("GET")
	r.HandleFunc("/api/games", handleCreateGame).Methods("POST")
	r.HandleFunc("/api/games/{id}", handleDeleteGame).Methods("DELETE")
	r.HandleFunc("/api/games/{id}/roster", handleUpdateRoster).Methods("PUT")
	r.HandleFunc("/api/games/{id}/availability", handleSetAvailability).Methods("POST")
	r.HandleFunc("/api/preferences", handleGetPreferences).Methods("GET")
	r.HandleFunc("/api/preferences/{playerId}", handleSetPreference).Methods("PUT")
	r.HandleFunc("/api/webhook", handleGetWebhook).Methods("GET")
	r.HandleFunc("/api/webhook", handleSetWebhook).Methods("PUT")
	r.HandleFunc("/api/discord/post/{id}", handlePostToDiscord).Methods("POST")

	// Internal routes for cron job
	r.HandleFunc("/api/internal/pending-reminders", handleGetPendingReminders).Methods("GET")
	r.HandleFunc("/api/internal/mark-reminded/{id}", handleMarkReminded).Methods("POST")
	r.HandleFunc("/api/internal/user-discord/{playerId}", handleGetUserDiscordID).Methods("GET")

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
