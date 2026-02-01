package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
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
}

type AllData struct {
	Games             []Game            `json:"games"`
	PlayerPreferences map[string]string `json:"playerPreferences"`
	DiscordWebhook    string            `json:"discordWebhook"`
}

// ==================== GLOBALS ====================

var db *sql.DB

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
	{ID: "shark", Name: "Shark", Year: 2021},
	{ID: "docbutler", Name: "GO_DocButler", Year: 2021},
	{ID: "drsmartazz", Name: "GO_DrSmartAzz", Year: 2021},
	{ID: "pbandc", Name: "PBandC-GO", Year: 2021},
	{ID: "maverick", Name: "GO_Maverick", Year: 2021},
	{ID: "honeygun", Name: "HoneyGUN", Year: 2022},
	{ID: "loki", Name: "GO_Loki714", Year: 2022},
	{ID: "lizlow", Name: "LizLow91", Year: 2023},
	{ID: "kc", Name: "GO_KC", Year: 2023},
	{ID: "lester", Name: "GO_Lester", Year: 2024},
	{ID: "kingslayer", Name: "GO_Kingslayer", Year: 2024, Region: "EU"},
	{ID: "bacon", Name: "AllTheBaconAndEGGz", Year: 2025},
	{ID: "stooobe", Name: "GO_STOOOBE", Year: 2025, Note: "Former Head of GO"},
	{ID: "johnharple", Name: "GO_JohnHarple", Year: 2025},
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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create games table: %v", err)
	}

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

	log.Println("Database initialized successfully")
	return nil
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

	rows, err := db.Query("SELECT id, date, time, opponent, notes, available, unavailable, roster FROM games ORDER BY date, time")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var available, unavailable, roster string
		if err := rows.Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.Notes, &available, &unavailable, &roster); err != nil {
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
	err := db.QueryRow("SELECT id, date, time, opponent, notes, available, unavailable, roster FROM games WHERE id = $1", gameID).
		Scan(&g.ID, &g.Date, &g.Time, &g.Opponent, &g.Notes, &available, &unavailable, &roster)
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
		"INSERT INTO games (id, date, time, opponent, notes, available, unavailable, roster) VALUES ($1, $2, $3, $4, $5, '[]', '[]', '[]')",
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
		var val string
		switch v := value.(type) {
		case []string:
			val = toJSONString(v)
		case string:
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

// ==================== HANDLERS ====================

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
	vars := mux.Vars(r)
	gameID := vars["id"]

	if err := deleteGame(gameID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func handleUpdateRoster(w http.ResponseWriter, r *http.Request) {
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
	resp, err := http.Post(webhook, "application/json", strings.NewReader(string(payloadBytes)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to post to Discord: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		writeError(w, http.StatusInternalServerError, "Discord API error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ==================== MAIN ====================

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := initDB(); err != nil {
		log.Printf("Warning: %v", err)
	}

	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/data", handleGetAllData).Methods("GET")
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

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
