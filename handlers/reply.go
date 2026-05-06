package handlers

import (
	"_discord_bot/models"
	"encoding/json"
)

type ReplyConfig struct {
	Message string `json:"message"`
}

type SimpleReplyHandler struct{}

func (h *SimpleReplyHandler) Handle(evt models.DiscordBotEvent, config string) models.BehaviourResult {
	var cfg ReplyConfig
	_ = json.Unmarshal([]byte(config), &cfg) // Error handling omitted for speed

	return models.BehaviourResult{
		Matched:  true,
		Executed: true,
		Continue: false,
		Actions: []models.BehaviourAction{
			{
				Type:    "SEND_MESSAGE", // Must match your ExecuteAction switch
				Target:  evt.ChannelID,
				Payload: cfg.Message,
			},
		},
	}
}
