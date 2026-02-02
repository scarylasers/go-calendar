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
	IsVet  bool   `json:"isVet,omitempty"`
}

type Game struct {
	ID          string   `json:"id"`
	Date        string   `json:"date"`
	Time        string   `json:"time"`
	Opponent    string   `json:"opponent"`
	League      string   `json:"league,omitempty"`
	Division    string   `json:"division,omitempty"`
	GameMode    string   `json:"gameMode,omitempty"`
	TeamSize    int      `json:"teamSize"`
	Notes       string   `json:"notes"`
	Available   []string `json:"available"`
	Unavailable []string `json:"unavailable"`
	Roster      []string `json:"roster"`
	Subs        []string `json:"subs"`
	Withdrawals []string `json:"withdrawals"`
	Reminded    bool     `json:"reminded"`
}

type User struct {
	DiscordID   string `json:"discordId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	PlayerID    string `json:"playerId"`
	IsManager   bool   `json:"isManager"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
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
	{ID: "shark", Name: "Shark", Year: 2021, IsVet: true},
	{ID: "docbutler", Name: "GO_DocButler", Year: 2021, IsVet: true},
	{ID: "drsmartazz", Name: "GO_DrSmartAzz", Year: 2021, IsVet: true},
	{ID: "pbandc", Name: "PBandC-GO", Year: 2021, IsVet: true},
	{ID: "maverick", Name: "GO_Maverick", Year: 2021, IsVet: true},
	{ID: "honeygun", Name: "HoneyGUN", Year: 2022, IsVet: true},
	{ID: "loki", Name: "GO_Loki714", Year: 2022, IsVet: true},
	{ID: "lizlow", Name: "LizLow91", Year: 2023, IsVet: true},
	{ID: "kc", Name: "GO_KC", Year: 2023, IsVet: true},
	{ID: "lester", Name: "GO_Lester", Year: 2024, IsVet: true},
	{ID: "kingslayer", Name: "GO_Kingslayer", Year: 2024, Region: "EU", IsVet: true},
	{ID: "bacon", Name: "AllTheBaconAndEGGz", Year: 2025, IsVet: true},
	{ID: "stooobe", Name: "GO_STOOOBE", Year: 2025, Note: "Former Head of GO", IsVet: true},
	{ID: "johnharple", Name: "GO_JohnHarple", Year: 2025, IsVet: true},
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
	// Add league and division columns
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS league TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS division TEXT DEFAULT ''`)
	// Add game mode and team size columns
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS game_mode TEXT DEFAULT 'War'`)
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS team_size INTEGER DEFAULT 10`)
	// Add withdrawals column for tracking players who pulled out
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS withdrawals TEXT DEFAULT '[]'`)
	// Add subs column for backup players
	db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS subs TEXT DEFAULT '[]'`)

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
	// Add email and phone columns for notifications
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT DEFAULT ''`)

	// Members table for roster management
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS members (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			year INTEGER DEFAULT 2025,
			region TEXT,
			note TEXT,
			is_vet BOOLEAN DEFAULT FALSE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create members table: %v", err)
	}

	// Seed members if table is empty
	var count int
	db.QueryRow("SELECT COUNT(*) FROM members").Scan(&count)
	if count == 0 {
		seedMembers()
	}

	log.Println("Database initialized successfully")
	return nil
}

func seedMembers() {
	log.Println("Seeding members table with initial data...")

	// Insert active members
	for i, m := range ActiveMembers {
		db.Exec(`INSERT INTO members (id, name, year, region, note, is_vet, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			m.ID, m.Name, m.Year, m.Region, m.Note, false, i)
	}

	// Insert sub members
	for i, m := range SubMembers {
		db.Exec(`INSERT INTO members (id, name, year, region, note, is_vet, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			m.ID, m.Name, m.Year, m.Region, m.Note, true, i)
	}

	log.Printf("Seeded %d active members and %d subs", len(ActiveMembers), len(SubMembers))
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
	var playerID, displayName, avatar, email, phone sql.NullString
	err := db.QueryRow(`
		SELECT discord_id, username, display_name, avatar, player_id, is_manager, COALESCE(email, ''), COALESCE(phone, '')
		FROM users WHERE discord_id = $1
	`, discordID).Scan(&u.DiscordID, &u.Username, &displayName, &avatar, &playerID, &u.IsManager, &email, &phone)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.DisplayName = displayName.String
	u.Avatar = avatar.String
	u.PlayerID = playerID.String
	u.Email = email.String
	u.Phone = phone.String

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

func getDiscordIDForPlayer(playerID string) string {
	var discordID string
	err := db.QueryRow("SELECT discord_id FROM users WHERE player_id = $1", playerID).Scan(&discordID)
	if err != nil {
		return ""
	}
	return discordID
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

func getUserAvatar(user *User) string {
	if user != nil {
		return user.Avatar
	}
	return ""
}

func getUserDisplayName(user *User) string {
	if user != nil {
		return user.DisplayName
	}
	return ""
}

// ==================== DATA FUNCTIONS ====================

func getAllGames() ([]Game, error) {
	if db == nil {
		return []Game{}, nil
	}

	rows, err := db.Query(`SELECT id, date, time, opponent, COALESCE(league, ''), COALESCE(division, ''),
		COALESCE(game_mode, 'War'), COALESCE(team_size, 10), notes, available, unavailable, roster,
		COALESCE(subs, '[]'), COALESCE(withdrawals, '[]'), COALESCE(reminded, false) FROM games ORDER BY date, time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var available, unavailable, roster, subs, withdrawals string
		if err := rows.Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.League, &g.Division, &g.GameMode, &g.TeamSize, &g.Notes, &available, &unavailable, &roster, &subs, &withdrawals, &g.Reminded); err != nil {
			return nil, err
		}
		g.Available = parseJSONArray(available)
		g.Unavailable = parseJSONArray(unavailable)
		g.Roster = parseJSONArray(roster)
		g.Subs = parseJSONArray(subs)
		g.Withdrawals = parseJSONArray(withdrawals)
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
	var available, unavailable, roster, subs, withdrawals string
	err := db.QueryRow(`SELECT id, date, time, opponent, COALESCE(league, ''), COALESCE(division, ''),
		COALESCE(game_mode, 'War'), COALESCE(team_size, 10), notes, available, unavailable, roster,
		COALESCE(subs, '[]'), COALESCE(withdrawals, '[]'), COALESCE(reminded, false) FROM games WHERE id = $1`, gameID).
		Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.League, &g.Division, &g.GameMode, &g.TeamSize, &g.Notes, &available, &unavailable, &roster, &subs, &withdrawals, &g.Reminded)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	g.Available = parseJSONArray(available)
	g.Unavailable = parseJSONArray(unavailable)
	g.Roster = parseJSONArray(roster)
	g.Subs = parseJSONArray(subs)
	g.Withdrawals = parseJSONArray(withdrawals)
	return &g, nil
}

func createGame(date, gameTime, opponent, league, division, gameMode string, teamSize int, notes string) (*Game, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	if gameMode == "" {
		gameMode = "War"
	}
	if teamSize <= 0 {
		teamSize = 10
	}

	gameID := generateGameID()
	_, err := db.Exec(
		`INSERT INTO games (id, date, time, opponent, league, division, game_mode, team_size, notes, available, unavailable, roster, subs, withdrawals, reminded)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, '[]', '[]', '[]', '[]', '[]', false)`,
		gameID, date, gameTime, opponent, league, division, gameMode, teamSize, notes,
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

	response := map[string]interface{}{
		"authenticated": true,
		"discordId":     session.DiscordID,
		"username":      session.Username,
		"isManager":     session.IsManager,
		"playerId":      session.PlayerID,
		"avatar":        getUserAvatar(user),
		"displayName":   getUserDisplayName(user),
	}

	// Include email and phone if available
	if user != nil {
		response["email"] = user.Email
		response["phone"] = user.Phone
	}

	writeJSON(w, http.StatusOK, response)
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

func handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var body struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if db == nil {
		writeError(w, http.StatusInternalServerError, "Database not connected")
		return
	}

	_, err := db.Exec(`UPDATE users SET email = $1, phone = $2 WHERE discord_id = $3`,
		body.Email, body.Phone, session.DiscordID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleGetLinkedUsers(w http.ResponseWriter, r *http.Request) {
	// Return a map of player_id -> {avatar, displayName}
	if db == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{})
		return
	}

	rows, err := db.Query(`
		SELECT player_id, avatar, display_name, username
		FROM users
		WHERE player_id IS NOT NULL AND player_id != ''
	`)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{})
		return
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var playerID, avatar, displayName, username string
		var avatarPtr, displayNamePtr *string

		if err := rows.Scan(&playerID, &avatarPtr, &displayNamePtr, &username); err != nil {
			continue
		}

		if avatarPtr != nil {
			avatar = *avatarPtr
		}
		if displayNamePtr != nil {
			displayName = *displayNamePtr
		}
		if displayName == "" {
			displayName = username
		}

		result[playerID] = map[string]string{
			"avatar":      avatar,
			"displayName": displayName,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// ==================== MEMBER HANDLERS ====================

func getMembersFromDB() ([]Member, []Member, error) {
	if db == nil {
		return ActiveMembers, SubMembers, nil
	}

	rows, err := db.Query("SELECT id, name, year, COALESCE(region, ''), COALESCE(note, ''), is_vet FROM members ORDER BY is_vet, sort_order, name")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var active []Member
	var subs []Member

	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Year, &m.Region, &m.Note, &m.IsVet); err != nil {
			continue
		}
		if m.IsVet {
			subs = append(subs, m)
		} else {
			active = append(active, m)
		}
	}

	return active, subs, nil
}

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

	active, subs, err := getMembersFromDB()
	if err != nil {
		// Fallback to hardcoded
		active = ActiveMembers
		subs = SubMembers
	}

	var activeWithStatus []MemberWithStatus
	var subsWithStatus []MemberWithStatus

	for _, m := range active {
		activeWithStatus = append(activeWithStatus, MemberWithStatus{Member: m, Linked: linkedMap[m.ID]})
	}
	for _, m := range subs {
		subsWithStatus = append(subsWithStatus, MemberWithStatus{Member: m, Linked: linkedMap[m.ID]})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"active": activeWithStatus,
		"subs":   subsWithStatus,
	})
}

func handleAddMember(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	var input struct {
		Name   string `json:"name"`
		IsVet  bool   `json:"isVet"`
		Region string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Generate ID from name
	id := strings.ToLower(strings.ReplaceAll(input.Name, " ", ""))
	id = strings.ReplaceAll(id, "_", "")
	id = strings.ReplaceAll(id, "-", "")

	// Get max sort order
	var maxOrder int
	db.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM members WHERE is_vet = $1", input.IsVet).Scan(&maxOrder)

	_, err := db.Exec(`INSERT INTO members (id, name, year, region, is_vet, sort_order) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, input.Name, 2025, input.Region, input.IsVet, maxOrder+1)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to add member")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"name":   input.Name,
		"isVet":  input.IsVet,
		"region": input.Region,
	})
}

func handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	memberID := vars["id"]

	var input struct {
		Name   *string `json:"name"`
		IsVet  *bool   `json:"isVet"`
		Region *string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argNum := 1

	if input.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *input.Name)
		argNum++
	}
	if input.IsVet != nil {
		updates = append(updates, fmt.Sprintf("is_vet = $%d", argNum))
		args = append(args, *input.IsVet)
		argNum++
	}
	if input.Region != nil {
		updates = append(updates, fmt.Sprintf("region = $%d", argNum))
		args = append(args, *input.Region)
		argNum++
	}

	if len(updates) == 0 {
		writeError(w, http.StatusBadRequest, "No updates provided")
		return
	}

	args = append(args, memberID)
	query := fmt.Sprintf("UPDATE members SET %s WHERE id = $%d", strings.Join(updates, ", "), argNum)

	_, err := db.Exec(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update member")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleDeleteMember(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	memberID := vars["id"]

	_, err := db.Exec("DELETE FROM members WHERE id = $1", memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete member")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleUpdateMemberOrder(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	var input struct {
		Type  string   `json:"type"` // "active" or "subs"
		Order []string `json:"order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update sort order for each member
	for i, id := range input.Order {
		db.Exec("UPDATE members SET sort_order = $1 WHERE id = $2", i, id)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
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
		League   string `json:"league"`
		Division string `json:"division"`
		GameMode string `json:"gameMode"`
		TeamSize int    `json:"teamSize"`
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

	game, err := createGame(body.Date, body.Time, body.Opponent, body.League, body.Division, body.GameMode, body.TeamSize, body.Notes)
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

func handleUpdateGame(w http.ResponseWriter, r *http.Request) {
	// Check manager permission
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	gameID := vars["id"]

	var body struct {
		Date     string `json:"date"`
		Time     string `json:"time"`
		Opponent string `json:"opponent"`
		League   string `json:"league"`
		Division string `json:"division"`
		GameMode string `json:"gameMode"`
		TeamSize int    `json:"teamSize"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if body.Date == "" || body.Time == "" || body.Opponent == "" {
		writeError(w, http.StatusBadRequest, "Date, time, and opponent are required")
		return
	}

	// Update the game in database
	_, err := db.Exec(`
		UPDATE games SET date = $1, time = $2, opponent = $3, league = $4,
		division = $5, game_mode = $6, team_size = $7, notes = $8
		WHERE id = $9
	`, body.Date, body.Time, body.Opponent, body.League, body.Division,
		body.GameMode, body.TeamSize, body.Notes, gameID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return updated game
	game, err := getGameByID(gameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, game)
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
		Subs   []string `json:"subs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Enforce roster size limit
	if len(body.Roster) > game.TeamSize {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Roster cannot exceed %d players", game.TeamSize))
		return
	}

	updates := map[string]interface{}{"roster": body.Roster}
	if body.Subs != nil {
		updates["subs"] = body.Subs
	}

	// Clear withdrawals if roster is now full (subs have been assigned)
	if len(body.Roster) >= game.TeamSize {
		updates["withdrawals"] = []string{}
	}

	updated, err := updateGame(gameID, updates)
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

func handleWithdrawFromRoster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	session := getSessionFromRequest(r)
	if session == nil || session.PlayerID == "" {
		writeError(w, http.StatusUnauthorized, "Must be logged in with linked player")
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

	// Check if player is on roster
	isOnRoster := false
	newRoster := make([]string, 0)
	for _, p := range game.Roster {
		if p == session.PlayerID {
			isOnRoster = true
		} else {
			newRoster = append(newRoster, p)
		}
	}

	if !isOnRoster {
		writeError(w, http.StatusBadRequest, "You are not on this roster")
		return
	}

	// Add to withdrawals list (if not already there)
	withdrawals := game.Withdrawals
	alreadyWithdrawn := false
	for _, w := range withdrawals {
		if w == session.PlayerID {
			alreadyWithdrawn = true
			break
		}
	}
	if !alreadyWithdrawn {
		withdrawals = append(withdrawals, session.PlayerID)
	}

	// Also mark as unavailable
	unavailable := game.Unavailable
	isAlreadyUnavailable := false
	for _, p := range unavailable {
		if p == session.PlayerID {
			isAlreadyUnavailable = true
			break
		}
	}
	if !isAlreadyUnavailable {
		unavailable = append(unavailable, session.PlayerID)
	}

	// Remove from available
	available := make([]string, 0)
	for _, p := range game.Available {
		if p != session.PlayerID {
			available = append(available, p)
		}
	}

	updated, err := updateGame(gameID, map[string]interface{}{
		"roster":      newRoster,
		"withdrawals": withdrawals,
		"available":   available,
		"unavailable": unavailable,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send notification to managers via Discord
	go notifyManagersOfWithdrawal(game, session.PlayerID)

	writeJSON(w, http.StatusOK, updated)
}

func notifyManagersOfWithdrawal(game *Game, playerID string) {
	if discordBotToken == "" || discordGuildID == "" {
		return
	}

	playerName := getMemberName(playerID)

	// Get all managers from users table
	if db == nil {
		return
	}

	rows, err := db.Query(`SELECT discord_id FROM users WHERE is_manager = true`)
	if err != nil {
		log.Printf("Error fetching managers for withdrawal notification: %v", err)
		return
	}
	defer rows.Close()

	message := fmt.Sprintf("⚠️ **Sub Needed**\n\n**%s** needs a sub for:\n📅 %s at %s\n⚔️ vs %s\n\nPlease find a replacement.",
		playerName, game.Date, game.Time, game.Opponent)

	for rows.Next() {
		var discordID string
		if err := rows.Scan(&discordID); err != nil {
			continue
		}
		sendDiscordDM(discordID, message)
	}
}

func sendDiscordDM(userID, message string) error {
	if discordBotToken == "" {
		return fmt.Errorf("bot token not configured")
	}

	// Create DM channel
	dmPayload := map[string]string{"recipient_id": userID}
	dmBytes, _ := json.Marshal(dmPayload)

	req, _ := http.NewRequest("POST", "https://discord.com/api/users/@me/channels", bytes.NewReader(dmBytes))
	req.Header.Set("Authorization", "Bot "+discordBotToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var channel struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return err
	}

	// Send message to DM channel
	msgPayload := map[string]string{"content": message}
	msgBytes, _ := json.Marshal(msgPayload)

	msgReq, _ := http.NewRequest("POST", "https://discord.com/api/channels/"+channel.ID+"/messages", bytes.NewReader(msgBytes))
	msgReq.Header.Set("Authorization", "Bot "+discordBotToken)
	msgReq.Header.Set("Content-Type", "application/json")

	msgResp, err := http.DefaultClient.Do(msgReq)
	if err != nil {
		return err
	}
	defer msgResp.Body.Close()

	return nil
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

// ==================== LEAGUES & DIVISIONS ====================

func getListSetting(key string) []string {
	value, _ := getSetting(key)
	if value == "" {
		return []string{}
	}
	var list []string
	json.Unmarshal([]byte(value), &list)
	return list
}

func setListSetting(key string, list []string) error {
	data, _ := json.Marshal(list)
	return setSetting(key, string(data))
}

func handleGetLeagues(w http.ResponseWriter, r *http.Request) {
	leagues := getListSetting("leagues")
	writeJSON(w, http.StatusOK, leagues)
}

func handleAddLeague(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	leagues := getListSetting("leagues")
	// Check if already exists
	for _, l := range leagues {
		if l == body.Name {
			writeError(w, http.StatusConflict, "League already exists")
			return
		}
	}
	leagues = append(leagues, body.Name)
	if err := setListSetting("leagues", leagues); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, leagues)
}

func handleDeleteLeague(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	leagues := getListSetting("leagues")
	newLeagues := []string{}
	for _, l := range leagues {
		if l != name {
			newLeagues = append(newLeagues, l)
		}
	}
	if err := setListSetting("leagues", newLeagues); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, newLeagues)
}

func handleGetDivisions(w http.ResponseWriter, r *http.Request) {
	divisions := getListSetting("divisions")
	writeJSON(w, http.StatusOK, divisions)
}

func handleAddDivision(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	divisions := getListSetting("divisions")
	for _, d := range divisions {
		if d == body.Name {
			writeError(w, http.StatusConflict, "Division already exists")
			return
		}
	}
	divisions = append(divisions, body.Name)
	if err := setListSetting("divisions", divisions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, divisions)
}

func handleDeleteDivision(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	divisions := getListSetting("divisions")
	newDivisions := []string{}
	for _, d := range divisions {
		if d != name {
			newDivisions = append(newDivisions, d)
		}
	}
	if err := setListSetting("divisions", newDivisions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, newDivisions)
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
	mentionPlayers := r.URL.Query().Get("mention") == "true"

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
	var mentionString string
	for _, pid := range game.Roster {
		rosterNames = append(rosterNames, getMemberName(pid))
		// If mentioning, look up Discord ID for this player
		if mentionPlayers {
			discordID := getDiscordIDForPlayer(pid)
			if discordID != "" {
				mentionString += fmt.Sprintf("<@%s> ", discordID)
			}
		}
	}

	rosterValue := "TBD"
	if len(rosterNames) > 0 {
		rosterValue = strings.Join(rosterNames, "\n")
	}

	// Build link for players to withdraw/request sub
	gameLink := baseURL + "/?game=" + gameID

	embed := map[string]interface{}{
		"title": fmt.Sprintf("🎮 Game Day: %s", formattedDate),
		"color": 0x00f0ff,
		"fields": []map[string]interface{}{
			{"name": "⏰ Time", "value": formattedTime, "inline": true},
			{"name": "⚔️ Opponent", "value": game.Opponent, "inline": true},
			{"name": fmt.Sprintf("👥 Roster (%d/10)", len(rosterNames)), "value": rosterValue, "inline": false},
			{"name": "🔗 Can't Make It?", "value": fmt.Sprintf("[Click here to request a sub](%s)", gameLink), "inline": false},
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

	// Add mentions as content (outside embed so they ping)
	if mentionPlayers && mentionString != "" {
		payload["content"] = "📢 Roster Alert! " + mentionString
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

func handleAnnounceGame(w http.ResponseWriter, r *http.Request) {
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

	// Build link for players to mark availability
	gameLink := baseURL + "/?game=" + gameID

	embed := map[string]interface{}{
		"title":       fmt.Sprintf("📢 Game Scheduled: %s", formattedDate),
		"description": "A new game has been scheduled! Please mark your availability.",
		"color":       0xf59e0b, // Orange/warning color
		"fields": []map[string]interface{}{
			{"name": "⏰ Time", "value": formattedTime, "inline": true},
			{"name": "⚔️ Opponent", "value": game.Opponent, "inline": true},
			{"name": "🎮 Game Mode", "value": game.Mode, "inline": true},
			{"name": "✅ Mark Availability", "value": fmt.Sprintf("[Click here to mark if you can play](%s)", gameLink), "inline": false},
		},
		"footer":    map[string]string{"text": "Game Over Pop1 War Team • Please respond ASAP!"},
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
		"content":  "📢 **New Game Scheduled!** Please check your availability!",
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

// ==================== TEST ENDPOINTS ====================

func handleTestDM(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if discordBotToken == "" {
		writeError(w, http.StatusServiceUnavailable, "Discord bot token not configured")
		return
	}

	message := "🧪 **Test Message from GO Calendar**\n\nThis is a test direct message. If you're seeing this, DM notifications are working correctly!"

	err := sendDiscordDM(session.DiscordID, message)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to send DM: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleTestWebhook(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	webhook, _ := getSetting("discord_webhook")
	if webhook == "" {
		writeError(w, http.StatusBadRequest, "Discord webhook not configured. Please set up a webhook first.")
		return
	}

	embed := map[string]interface{}{
		"title":       "🧪 Test Post from GO Calendar",
		"description": "This is a test message. If you're seeing this, your webhook is configured correctly!",
		"color":       0x00f0ff,
		"footer":      map[string]string{"text": "Game Over Pop1 War Team"},
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
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

func handleTestSubNotification(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromRequest(r)
	if session == nil || !session.IsManager {
		writeError(w, http.StatusForbidden, "Manager access required")
		return
	}

	if discordBotToken == "" {
		writeError(w, http.StatusServiceUnavailable, "Discord bot token not configured")
		return
	}

	// Get the requesting user's name
	playerName := "Test Player"
	if session.PlayerID != "" {
		playerName = getMemberName(session.PlayerID)
	}

	message := fmt.Sprintf("🧪 **Test Sub Notification**\n\nThis is a test of the 'Need a Sub' alert system.\n\n**%s** would need a sub for:\n📅 Tomorrow at 9:00 PM ET\n⚔️ vs Test Opponent\n\nIf you received this, manager notifications are working!",
		playerName)

	// Get all managers and send DM
	if db == nil {
		writeError(w, http.StatusInternalServerError, "Database not connected")
		return
	}

	rows, err := db.Query(`SELECT discord_id FROM users WHERE is_manager = true`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch managers: "+err.Error())
		return
	}
	defer rows.Close()

	sentCount := 0
	for rows.Next() {
		var discordID string
		if err := rows.Scan(&discordID); err != nil {
			continue
		}
		if err := sendDiscordDM(discordID, message); err == nil {
			sentCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"sentCount": sentCount,
	})
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
		SELECT id, date, time, opponent, COALESCE(league, ''), COALESCE(division, ''),
		COALESCE(game_mode, 'War'), COALESCE(team_size, 10), notes, available, unavailable, roster,
		COALESCE(subs, '[]'), COALESCE(withdrawals, '[]'), COALESCE(reminded, false)
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
		var available, unavailable, roster, subs, withdrawals string
		if err := rows.Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.League, &g.Division, &g.GameMode, &g.TeamSize, &g.Notes, &available, &unavailable, &roster, &subs, &withdrawals, &g.Reminded); err != nil {
			continue
		}
		g.Available = parseJSONArray(available)
		g.Unavailable = parseJSONArray(unavailable)
		g.Roster = parseJSONArray(roster)
		g.Subs = parseJSONArray(subs)
		g.Withdrawals = parseJSONArray(withdrawals)
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

// ==================== SECURITY MIDDLEWARE ====================

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy - don't leak URLs
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' https://cdn.discordapp.com data:; connect-src 'self'")
		// Permissions Policy - disable unnecessary browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
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
	r.HandleFunc("/auth/account", handleUpdateAccount).Methods("PUT")

	// API routes
	r.HandleFunc("/api/data", handleGetAllData).Methods("GET")
	r.HandleFunc("/api/members", handleGetMembers).Methods("GET")
	r.HandleFunc("/api/members", handleAddMember).Methods("POST")
	r.HandleFunc("/api/members/{id}", handleUpdateMember).Methods("PUT")
	r.HandleFunc("/api/members/{id}", handleDeleteMember).Methods("DELETE")
	r.HandleFunc("/api/members/order", handleUpdateMemberOrder).Methods("PUT")
	r.HandleFunc("/api/games", handleGetGames).Methods("GET")
	r.HandleFunc("/api/games", handleCreateGame).Methods("POST")
	r.HandleFunc("/api/games/{id}", handleUpdateGame).Methods("PUT")
	r.HandleFunc("/api/games/{id}", handleDeleteGame).Methods("DELETE")
	r.HandleFunc("/api/games/{id}/roster", handleUpdateRoster).Methods("PUT")
	r.HandleFunc("/api/games/{id}/availability", handleSetAvailability).Methods("POST")
	r.HandleFunc("/api/games/{id}/withdraw", handleWithdrawFromRoster).Methods("POST")
	r.HandleFunc("/api/preferences", handleGetPreferences).Methods("GET")
	r.HandleFunc("/api/preferences/{playerId}", handleSetPreference).Methods("PUT")
	r.HandleFunc("/api/webhook", handleGetWebhook).Methods("GET")
	r.HandleFunc("/api/webhook", handleSetWebhook).Methods("PUT")
	r.HandleFunc("/api/leagues", handleGetLeagues).Methods("GET")
	r.HandleFunc("/api/leagues", handleAddLeague).Methods("POST")
	r.HandleFunc("/api/leagues/{name}", handleDeleteLeague).Methods("DELETE")
	r.HandleFunc("/api/divisions", handleGetDivisions).Methods("GET")
	r.HandleFunc("/api/divisions", handleAddDivision).Methods("POST")
	r.HandleFunc("/api/divisions/{name}", handleDeleteDivision).Methods("DELETE")
	r.HandleFunc("/api/discord/post/{id}", handlePostToDiscord).Methods("POST")
	r.HandleFunc("/api/games/{id}/announce", handleAnnounceGame).Methods("POST")
	r.HandleFunc("/api/users/linked", handleGetLinkedUsers).Methods("GET")

	// Test routes
	r.HandleFunc("/api/test/dm", handleTestDM).Methods("POST")
	r.HandleFunc("/api/test/webhook", handleTestWebhook).Methods("POST")
	r.HandleFunc("/api/test/sub-notification", handleTestSubNotification).Methods("POST")

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
	log.Fatal(http.ListenAndServe(":"+port, securityHeaders(r)))
}
