package services

import (
	"_discord_bot/db"
	"_discord_bot/models"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type botState struct {
	session  *discordgo.Session
	stopTime chan struct{}
}

var (
	activeSessions = make(map[string]*botState)
	sessionMu      sync.RWMutex
)

func Execute(s *discordgo.Session, botUUID string, action models.BehaviourAction) error {
	switch action.Type {
	case "SEND_MESSAGE":
		_, err := s.ChannelMessageSend(action.Target, action.Payload)
		if err != nil {
			log.Printf("[Bot %s] Execution Error: %v", botUUID, err)
			return err
		}
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
	return nil
}

func StartBotBackground(botUUID string) error {
	bot, err := models.GetFullDiscordBot(db.DB, botUUID)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	username, err := models.ValidateBotCredentials(bot.Token)
	if err != nil {
		return fmt.Errorf("auth check failed: %v", err)
	}

	StopBot(botUUID)

	dg, err := discordgo.New("Bot " + bot.Token)
	if err != nil {
		return fmt.Errorf("init error: %v", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentMessageContent |
		discordgo.IntentsDirectMessages

	dg.AddHandler(handleDiscordMessage(botUUID))

	if err := dg.Open(); err != nil {
		if strings.Contains(err.Error(), "4014") {
			return fmt.Errorf("critical: Message Content Intent disabled in Discord Portal")
		}
		return fmt.Errorf("gateway connection failed: %v", err)
	}

	stopChan := make(chan struct{})
	sessionMu.Lock()
	activeSessions[botUUID] = &botState{
		session:  dg,
		stopTime: stopChan,
	}
	sessionMu.Unlock()

	go runOnTimeWorker(botUUID, dg, stopChan)

	log.Printf("[Worker] %s is now online", username)
	return nil
}

func StopBot(botUUID string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	if state, exists := activeSessions[botUUID]; exists {
		close(state.stopTime)
		state.session.Close()
		delete(activeSessions, botUUID)
		log.Printf("[Bot %s] Worker Stopped", botUUID)
	}
}

func runOnTimeWorker(botUUID string, s *discordgo.Session, stop <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			var tasks []models.OnTimeBehaviour
			db.DB.Where("bot_uuid = ? AND enabled = ?", botUUID, true).Find(&tasks)

			now := time.Now().Unix()
			for _, t := range tasks {
				if (t.TimeType == "interval" && now-t.LastRun >= t.TimeValue) ||
					(t.TimeType == "specific" && now >= t.TimeValue && t.LastRun == 0) {

					updates := map[string]any{"last_run": now}
					if t.TimeType == "specific" {
						updates["enabled"] = false
					}

					if err := db.DB.Model(&t).Updates(updates).Error; err != nil {
						log.Printf("[Bot %s] Failed to update task status: %v", botUUID, err)
						continue
					}

					action := models.BehaviourAction{
						Type:    "SEND_MESSAGE",
						Target:  t.ChannelID,
						Payload: t.Message,
					}

					Execute(s, botUUID, action)
				}
			}
		}
	}
}
func handleDiscordMessage(botUUID string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	runtime := models.BotRuntime{DB: db.DB}
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
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

		actions, err := runtime.HandleEvent(evt)
		if err != nil {
			log.Printf("[Bot %s] Logic Error: %v", botUUID, err)
			return
		}

		for _, action := range actions {
			Execute(s, botUUID, action)
		}
	}
}

func IsBotRunning(botUUID string) bool {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	_, running := activeSessions[botUUID]
	return running
}
