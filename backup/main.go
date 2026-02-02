package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Data structures for export
type Game struct {
	ID          string   `json:"id"`
	Date        string   `json:"date"`
	Time        string   `json:"time"`
	Opponent    string   `json:"opponent"`
	League      string   `json:"league"`
	Division    string   `json:"division"`
	GameMode    string   `json:"gameMode"`
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
	PlayerID    string `json:"playerId"`
	IsManager   bool   `json:"isManager"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type Member struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Year      int    `json:"year"`
	Region    string `json:"region"`
	Note      string `json:"note"`
	IsSub     bool   `json:"isSub"`
	SortOrder int    `json:"sortOrder"`
}

type Preference struct {
	PlayerID   string `json:"playerId"`
	Preference string `json:"preference"`
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Backup struct {
	Timestamp   string       `json:"timestamp"`
	Games       []Game       `json:"games"`
	Users       []User       `json:"users"`
	Members     []Member     `json:"members"`
	Preferences []Preference `json:"preferences"`
	Settings    []Setting    `json:"settings"`
}

var db *sql.DB

func main() {
	log.Println("Starting database backup...")

	dbURL := os.Getenv("DATABASE_URL")
	webhookURL := os.Getenv("BACKUP_WEBHOOK_URL")

	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	if webhookURL == "" {
		log.Fatal("BACKUP_WEBHOOK_URL not set")
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create backup
	backup := Backup{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Export games
	backup.Games, err = exportGames()
	if err != nil {
		log.Printf("Warning: Failed to export games: %v", err)
	}

	// Export users
	backup.Users, err = exportUsers()
	if err != nil {
		log.Printf("Warning: Failed to export users: %v", err)
	}

	// Export members
	backup.Members, err = exportMembers()
	if err != nil {
		log.Printf("Warning: Failed to export members: %v", err)
	}

	// Export preferences
	backup.Preferences, err = exportPreferences()
	if err != nil {
		log.Printf("Warning: Failed to export preferences: %v", err)
	}

	// Export settings
	backup.Settings, err = exportSettings()
	if err != nil {
		log.Printf("Warning: Failed to export settings: %v", err)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal backup: %v", err)
	}

	// Send to Discord
	err = sendToDiscord(webhookURL, backup, jsonData)
	if err != nil {
		log.Fatalf("Failed to send backup to Discord: %v", err)
	}

	log.Println("Backup completed successfully!")
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

func exportGames() ([]Game, error) {
	rows, err := db.Query(`
		SELECT id, date, time, opponent,
			COALESCE(league, ''), COALESCE(division, ''),
			COALESCE(game_mode, 'War'), COALESCE(team_size, 10),
			COALESCE(notes, ''), COALESCE(available, '[]'),
			COALESCE(unavailable, '[]'), COALESCE(roster, '[]'),
			COALESCE(subs, '[]'), COALESCE(withdrawals, '[]'),
			COALESCE(reminded, false)
		FROM games ORDER BY date, time
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var available, unavailable, roster, subs, withdrawals string
		err := rows.Scan(&g.ID, &g.Date, &g.Time, &g.Opponent,
			&g.League, &g.Division, &g.GameMode, &g.TeamSize,
			&g.Notes, &available, &unavailable, &roster,
			&subs, &withdrawals, &g.Reminded)
		if err != nil {
			continue
		}
		g.Available = parseJSONArray(available)
		g.Unavailable = parseJSONArray(unavailable)
		g.Roster = parseJSONArray(roster)
		g.Subs = parseJSONArray(subs)
		g.Withdrawals = parseJSONArray(withdrawals)
		games = append(games, g)
	}
	return games, nil
}

func exportUsers() ([]User, error) {
	rows, err := db.Query(`
		SELECT discord_id, username,
			COALESCE(display_name, ''), COALESCE(player_id, ''),
			COALESCE(is_manager, false),
			COALESCE(email, ''), COALESCE(phone, '')
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.DiscordID, &u.Username, &u.DisplayName,
			&u.PlayerID, &u.IsManager, &u.Email, &u.Phone)
		if err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func exportMembers() ([]Member, error) {
	rows, err := db.Query(`
		SELECT id, name, COALESCE(year, 2025),
			COALESCE(region, ''), COALESCE(note, ''),
			COALESCE(is_sub, false), COALESCE(sort_order, 0)
		FROM members ORDER BY is_sub, sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		err := rows.Scan(&m.ID, &m.Name, &m.Year, &m.Region,
			&m.Note, &m.IsSub, &m.SortOrder)
		if err != nil {
			continue
		}
		members = append(members, m)
	}
	return members, nil
}

func exportPreferences() ([]Preference, error) {
	rows, err := db.Query(`SELECT player_id, preference FROM player_preferences`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []Preference
	for rows.Next() {
		var p Preference
		err := rows.Scan(&p.PlayerID, &p.Preference)
		if err != nil {
			continue
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func exportSettings() ([]Setting, error) {
	rows, err := db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var s Setting
		err := rows.Scan(&s.Key, &s.Value)
		if err != nil {
			continue
		}
		// Mask sensitive values
		if s.Key == "discord_webhook" && len(s.Value) > 20 {
			s.Value = s.Value[:20] + "..."
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func sendToDiscord(webhookURL string, backup Backup, jsonData []byte) error {
	// Create summary message
	summary := fmt.Sprintf("**Database Backup - %s**\n\n"+
		"**Games:** %d\n"+
		"**Users:** %d\n"+
		"**Members:** %d\n"+
		"**Preferences:** %d\n"+
		"**Settings:** %d\n\n"+
		"Full backup attached as JSON file.",
		time.Now().UTC().Format("Jan 02, 2006 15:04 UTC"),
		len(backup.Games),
		len(backup.Users),
		len(backup.Members),
		len(backup.Preferences),
		len(backup.Settings))

	// Create multipart form for file upload
	var body bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"

	// Add content field
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"payload_json\"\r\n\r\n")
	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"content": summary,
		"embeds": []map[string]interface{}{
			{
				"title": "GO Calendar Backup",
				"color": 0x00f0ff,
				"footer": map[string]string{
					"text": "Automated backup",
				},
			},
		},
	})
	body.Write(payloadJSON)
	body.WriteString("\r\n")

	// Add file
	filename := fmt.Sprintf("backup_%s.json", time.Now().UTC().Format("2006-01-02"))
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\n", filename))
	body.WriteString("Content-Type: application/json\r\n\r\n")
	body.Write(jsonData)
	body.WriteString("\r\n")
	body.WriteString("--" + boundary + "--\r\n")

	req, err := http.NewRequest("POST", webhookURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Discord API error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
