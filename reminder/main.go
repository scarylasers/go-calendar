package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Game represents a game that needs reminders
type Game struct {
	ID       int       `json:"id"`
	Date     time.Time `json:"date"`
	Opponent string    `json:"opponent"`
	Roster   []string  `json:"roster"`
}

// User represents a linked Discord user
type User struct {
	DiscordID   string `json:"discord_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	PlayerID    string `json:"player_id"`
}

func main() {
	log.Println("Starting reminder cron job...")

	// Get environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Get games that need reminders (within 24 hours, not yet reminded)
	games, err := getPendingGames(db)
	if err != nil {
		log.Fatalf("Failed to get pending games: %v", err)
	}

	if len(games) == 0 {
		log.Println("No games need reminders at this time")
		return
	}

	log.Printf("Found %d games needing reminders", len(games))

	// Process each game
	for _, game := range games {
		log.Printf("Processing game %d vs %s on %s", game.ID, game.Opponent, game.Date.Format("Jan 2, 2006"))

		if len(game.Roster) == 0 {
			log.Printf("Game %d has no roster set, skipping", game.ID)
			continue
		}

		// Get Discord IDs for rostered players
		discordIDs, err := getDiscordIDsForPlayers(db, game.Roster)
		if err != nil {
			log.Printf("Error getting Discord IDs for game %d: %v", game.ID, err)
			continue
		}

		// Send DM to each player
		sentCount := 0
		for _, discordID := range discordIDs {
			err := sendDiscordDM(botToken, discordID, game)
			if err != nil {
				log.Printf("Failed to send DM to %s: %v", discordID, err)
			} else {
				sentCount++
			}
		}

		log.Printf("Sent %d/%d reminders for game %d", sentCount, len(discordIDs), game.ID)

		// Mark game as reminded
		if err := markGameReminded(db, game.ID); err != nil {
			log.Printf("Failed to mark game %d as reminded: %v", game.ID, err)
		}
	}

	log.Println("Reminder cron job completed")
}

// getPendingGames returns games that are within 24 hours and haven't been reminded
func getPendingGames(db *sql.DB) ([]Game, error) {
	// Calculate the time window (24 hours from now, +/- 30 minutes buffer)
	now := time.Now().UTC()
	startWindow := now.Add(23*time.Hour + 30*time.Minute)
	endWindow := now.Add(24*time.Hour + 30*time.Minute)

	query := `
		SELECT id, date, opponent, roster
		FROM games
		WHERE date >= $1 AND date <= $2
		AND reminded = false
		AND roster IS NOT NULL
		AND array_length(roster, 1) > 0
	`

	rows, err := db.Query(query, startWindow, endWindow)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var game Game
		var rosterJSON []byte

		if err := rows.Scan(&game.ID, &game.Date, &game.Opponent, &rosterJSON); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Parse roster array
		if len(rosterJSON) > 0 {
			// PostgreSQL arrays come as {val1,val2,...}
			rosterStr := string(rosterJSON)
			rosterStr = strings.Trim(rosterStr, "{}")
			if rosterStr != "" {
				game.Roster = strings.Split(rosterStr, ",")
			}
		}

		games = append(games, game)
	}

	return games, nil
}

// getDiscordIDsForPlayers looks up Discord IDs for a list of player IDs
func getDiscordIDsForPlayers(db *sql.DB, playerIDs []string) ([]string, error) {
	if len(playerIDs) == 0 {
		return nil, nil
	}

	// Build query with placeholders
	placeholders := make([]string, len(playerIDs))
	args := make([]interface{}, len(playerIDs))
	for i, id := range playerIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT discord_id FROM users
		WHERE player_id IN (%s)
		AND discord_id IS NOT NULL
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var discordIDs []string
	for rows.Next() {
		var discordID string
		if err := rows.Scan(&discordID); err != nil {
			continue
		}
		discordIDs = append(discordIDs, discordID)
	}

	return discordIDs, nil
}

// sendDiscordDM sends a direct message to a Discord user
func sendDiscordDM(botToken, userID string, game Game) error {
	// First, create a DM channel with the user
	channelID, err := createDMChannel(botToken, userID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel: %w", err)
	}

	// Format the reminder message
	gameTime := game.Date.Format("Monday, January 2 at 3:04 PM")
	message := fmt.Sprintf(
		"**Game Reminder!**\n\n"+
			"You're on the roster for tomorrow's game!\n\n"+
			"**Opponent:** %s\n"+
			"**When:** %s ET\n\n"+
			"Good luck out there!",
		game.Opponent,
		gameTime,
	)

	// Send the message
	return sendMessage(botToken, channelID, message)
}

// createDMChannel creates a DM channel with a user and returns the channel ID
func createDMChannel(botToken, userID string) (string, error) {
	url := "https://discord.com/api/v10/users/@me/channels"

	payload := map[string]string{
		"recipient_id": userID,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create DM channel: status %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

// sendMessage sends a message to a Discord channel
func sendMessage(botToken, channelID, content string) error {
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)

	payload := map[string]string{
		"content": content,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message: status %d", resp.StatusCode)
	}

	return nil
}

// markGameReminded marks a game as having been reminded
func markGameReminded(db *sql.DB, gameID int) error {
	_, err := db.Exec("UPDATE games SET reminded = true WHERE id = $1", gameID)
	return err
}
