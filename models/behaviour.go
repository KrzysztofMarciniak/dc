package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type DiscordBotBehaviourType struct {
	ID   uint   `gorm:"primaryKey"`
	Key  string `gorm:"size:50;uniqueIndex"`
	Name string `gorm:"size:100"`
}

type DiscordBotBehaviourTrigger struct {
	ID uint `gorm:"primaryKey"`

	BehaviourID uint `gorm:"index;not null;constraint:OnDelete:CASCADE;"`

	EventType string `gorm:"size:50;index"` // on_message, on_reaction_add, etc.
	Type      string `gorm:"size:30;index"` // command, contains, regex, always
	Value     string `gorm:"type:text"`
}

type DiscordBotBehaviour struct {
	ID      uint   `gorm:"primaryKey"`
	BotUUID string `gorm:"index;not null"`

	TypeID uint
	Type   DiscordBotBehaviourType `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	Config []byte `gorm:"type:json"`

	Priority int  `gorm:"default:0;index"`
	Enabled  bool `gorm:"default:true;index"`

	Triggers []DiscordBotBehaviourTrigger `gorm:"foreignKey:BehaviourID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type DiscordBotEvent struct {
	Name      string `json:"name"`
	BotUUID   string `json:"bot_uuid"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Raw       []byte `json:"raw"`
	CreatedAt time.Time
}

type BehaviourResult struct {
	BehaviourID uint
	Matched     bool
	Executed    bool
	Continue    bool
	Error       string
	Actions     []BehaviourAction
	DurationMs  int64
}

type BehaviourAction struct {
	Type    string
	Target  string
	Payload string
}

type BotRuntime struct {
	DB *gorm.DB
}

type BehaviourHandler interface {
	Handle(evt DiscordBotEvent, config string) BehaviourResult
}

var BehaviourRegistry = map[string]func() BehaviourHandler{}
var registryMu sync.RWMutex
var regexCache sync.Map

// -----------------------------
// REGISTRY
// -----------------------------

func RegisterBehaviourHandler(key string, factory func() BehaviourHandler) {
	registryMu.Lock()
	defer registryMu.Unlock()
	BehaviourRegistry[key] = factory
}

// -----------------------------
// EXECUTION
// -----------------------------

func executeBehaviour(b DiscordBotBehaviour, evt DiscordBotEvent) BehaviourResult {
	registryMu.RLock()
	factory, ok := BehaviourRegistry[b.Type.Key]
	registryMu.RUnlock()

	if !ok || factory == nil {
		return BehaviourResult{
			BehaviourID: b.ID,
			Matched:     true,
			Executed:    false,
			Continue:    true,
			Error:       "no handler registered",
		}
	}

	handler := factory()
	res := handler.Handle(evt, string(b.Config))
	res.BehaviourID = b.ID
	res.Matched = true
	return res
}

func (r *BotRuntime) HandleEvent(evt DiscordBotEvent) ([]BehaviourResult, error) {
	var bot DiscordBotManagement

	err := r.DB.
		Preload("Behaviours.Type").
		Preload("Behaviours.Triggers", "event_type = ?", evt.Name).
		Where("uuid = ? AND retired = ?", evt.BotUUID, false).
		First(&bot).Error
	if err != nil {
		return nil, err
	}

	behaviours := make([]DiscordBotBehaviour, 0, len(bot.Behaviours))
	for _, b := range bot.Behaviours {
		if b.Enabled {
			behaviours = append(behaviours, b)
		}
	}

	sort.Slice(behaviours, func(i, j int) bool {
		return behaviours[i].Priority < behaviours[j].Priority
	})

	results := make([]BehaviourResult, 0, len(behaviours))

	for _, b := range behaviours {
		if !matchTriggers(b, evt) {
			continue
		}

		res := safeExecute(b, evt)
		//for _, a := range res.Actions {
		//executor.Execute(evt, a)
		//		}
		results = append(results, res)

		if !res.Continue {
			break
		}
	}

	return results, nil
}

func safeExecute(b DiscordBotBehaviour, evt DiscordBotEvent) (res BehaviourResult) {
	defer func() {
		if r := recover(); r != nil {
			res = BehaviourResult{
				BehaviourID: b.ID,
				Matched:     true,
				Executed:    false,
				Continue:    true,
				Error:       "panic recovered",
			}
		}
	}()

	return executeBehaviour(b, evt)
}

// -----------------------------
// TRIGGER MATCHING
// -----------------------------

func matchTriggers(b DiscordBotBehaviour, evt DiscordBotEvent) bool {
	for _, t := range b.Triggers {
		if t.EventType != evt.Name {
			continue
		}

		switch t.Type {
		case "always":
			return true

		case "contains":
			if strings.Contains(evt.Content, t.Value) {
				return true
			}

		case "command":
			if strings.HasPrefix(evt.Content, t.Value) {
				return true
			}

		case "regex":
			v, ok := regexCache.Load(t.Value)
			if !ok {
				re, err := regexp.Compile(t.Value)
				if err != nil {
					continue
				}
				regexCache.Store(t.Value, re)
				v = re
			}

			if v.(*regexp.Regexp).MatchString(evt.Content) {
				return true
			}
		}
	}

	return false
}

// -----------------------------
// VALIDATION
// -----------------------------

func isValidTriggerType(t string) bool {
	switch t {
	case "always", "contains", "command", "regex":
		return true
	}
	return false
}

func isValidEventType(e string) bool {
	switch e {
	case "on_message", "on_reaction_add", "on_member_join", "cron":
		return true
	}
	return false
}

func isValidJSON(s string) bool {
	var js any
	return json.Unmarshal([]byte(s), &js) == nil
}

// -----------------------------
// BEHAVIOUR CRUD
// -----------------------------

func CreateBehaviour(db *gorm.DB, botUUID string, typeID uint, config string, priority int) (DiscordBotBehaviour, error) {
	if !isValidJSON(config) {
		return DiscordBotBehaviour{}, errors.New("invalid json config")
	}

	b := DiscordBotBehaviour{
		BotUUID:  botUUID,
		TypeID:   typeID,
		Config:   []byte(config),
		Priority: priority,
		Enabled:  true,
	}

	err := db.Create(&b).Error
	return b, err
}

func GetBehaviour(db *gorm.DB, id uint) (DiscordBotBehaviour, error) {
	var b DiscordBotBehaviour

	err := db.
		Preload("Triggers").
		Preload("Type").
		First(&b, id).Error

	return b, err
}

func ListBehaviours(db *gorm.DB, botUUID string) ([]DiscordBotBehaviour, error) {
	var list []DiscordBotBehaviour

	err := db.
		Preload("Triggers").
		Where("bot_uuid = ?", botUUID).
		Order("priority asc").
		Find(&list).Error

	return list, err
}

func ListBehavioursFull(db *gorm.DB, botUUID string) ([]DiscordBotBehaviour, error) {
	var list []DiscordBotBehaviour

	err := db.
		Preload("Type").
		Preload("Triggers").
		Where("bot_uuid = ?", botUUID).
		Order("priority asc").
		Find(&list).Error

	return list, err
}

func UpdateBehaviour(db *gorm.DB, id uint, config string, priority int, enabled bool) error {
	if !isValidJSON(config) {
		return errors.New("invalid json config")
	}

	return db.Model(&DiscordBotBehaviour{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"config":   []byte(config),
			"priority": priority,
			"enabled":  enabled,
		}).Error
}

func SetBehaviourEnabled(db *gorm.DB, id uint, enabled bool) error {
	return db.Model(&DiscordBotBehaviour{}).
		Where("id = ?", id).
		Update("enabled", enabled).Error
}

func DeleteBehaviour(db *gorm.DB, id uint) error {
	return db.Delete(&DiscordBotBehaviour{}, id).Error
}

func ReorderBehaviours(db *gorm.DB, botUUID string, orderedIDs []uint) error {
	for i, id := range orderedIDs {
		if err := db.Model(&DiscordBotBehaviour{}).
			Where("id = ? AND bot_uuid = ?", id, botUUID).
			Update("priority", i).Error; err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------
// TRIGGER CRUD
// -----------------------------

func CreateTrigger(db *gorm.DB, behaviourID uint, eventType, ttype, value string) (DiscordBotBehaviourTrigger, error) {
	if !isValidTriggerType(ttype) || !isValidEventType(eventType) {
		return DiscordBotBehaviourTrigger{}, errors.New("invalid trigger")
	}

	t := DiscordBotBehaviourTrigger{
		BehaviourID: behaviourID,
		EventType:   eventType,
		Type:        ttype,
		Value:       value,
	}

	err := db.Create(&t).Error
	return t, err
}

func ListTriggers(db *gorm.DB, behaviourID uint) ([]DiscordBotBehaviourTrigger, error) {
	var list []DiscordBotBehaviourTrigger

	err := db.
		Where("behaviour_id = ?", behaviourID).
		Find(&list).Error

	return list, err
}

func UpdateTrigger(db *gorm.DB, id uint, eventType, ttype, value string) error {
	if !isValidTriggerType(ttype) || !isValidEventType(eventType) {
		return errors.New("invalid trigger")
	}

	return db.Model(&DiscordBotBehaviourTrigger{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"event_type": eventType,
			"type":       ttype,
			"value":      value,
		}).Error
}

func DeleteTrigger(db *gorm.DB, id uint) error {
	return db.Delete(&DiscordBotBehaviourTrigger{}, id).Error
}

func DeleteTriggersByBehaviour(db *gorm.DB, behaviourID uint) error {
	return db.Where("behaviour_id = ?", behaviourID).
		Delete(&DiscordBotBehaviourTrigger{}).Error
}
