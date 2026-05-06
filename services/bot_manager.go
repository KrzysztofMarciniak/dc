package services

import (
	"_discord_bot/db"
	"_discord_bot/models"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	// activeSessions tracks botUUID -> *discordgo.Session
	activeSessions = make(map[string]*discordgo.Session)
	sessionMu      sync.RWMutex
)

func StartBotBackground(botUUID string) error {
	bot, err := models.GetFullDiscordBot(db.DB, botUUID)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	username, err := models.ValidateBotCredentials(bot.Token)
	if err != nil {
		return fmt.Errorf("auth check failed: %v", err)
	}

	// Synchronously kill the old session before starting a new one
	StopBot(botUUID)

	dg, err := discordgo.New("Bot " + bot.Token)
	if err != nil {
		return fmt.Errorf("init error: %v", err)
	}

	// These require the Developer Portal toggles
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentMessageContent |
		discordgo.IntentsDirectMessages

	dg.AddHandler(handleDiscordMessage(botUUID))

	// BLOCKING CALL: This waits for the handshake.
	// If it fails with 4014 (Disallowed Intents), we catch it here.
	if err := dg.Open(); err != nil {
		if strings.Contains(err.Error(), "4014") {
			return fmt.Errorf("critical: Message Content Intent is disabled in Discord Dev Portal")
		}
		return fmt.Errorf("gateway connection failed: %v", err)
	}

	// Now that it's open and verified, save it to the map
	sessionMu.Lock()
	activeSessions[botUUID] = dg
	sessionMu.Unlock()

	log.Printf("[Worker] %s is now online", username)
	return nil
}

// StopBot kills the session and removes it from the registry
func StopBot(botUUID string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	sess, exists := activeSessions[botUUID]
	if !exists {
		return
	}

	// 1. Attempt to close the WebSocket connection
	if err := sess.Close(); err != nil {
		log.Printf("[Bot %s] Error closing session: %v", botUUID, err)
		// We still proceed to delete from map because the session is
		// essentially dead or corrupted at this point.
	}

	// 2. Remove from the memory registry
	delete(activeSessions, botUUID)
	log.Printf("[Bot %s] Worker Stopped and cleared from memory", botUUID)
}

// handleDiscordMessage returns the actual handler logic
func handleDiscordMessage(botUUID string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Don't reply to self or other bots
		if m.Author.ID == s.State.User.ID || m.Author.Bot {
			return
		}

		evt := models.DiscordBotEvent{
			Name:      "on_message",
			BotUUID:   botUUID,
			ChannelID: m.ChannelID,
			UserID:    m.Author.ID,
			Content:   m.Content,
		}

		runtime := models.BotRuntime{DB: db.DB}
		actions, err := runtime.HandleEvent(evt)
		if err != nil {
			log.Printf("[Bot %s] Logic Error: %v", botUUID, err)
			return
		}

		for _, action := range actions {
			if action.Type == "SEND_MESSAGE" {
				s.ChannelMessageSend(action.Target, action.Payload)
			}
		}
	}
}

func IsBotRunning(botUUID string) bool {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	_, running := activeSessions[botUUID]
	return running
}
