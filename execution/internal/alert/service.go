package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Level string

const (
	LevelInfo    Level = "info"
	LevelWarn    Level = "warn"
	LevelError   Level = "error"
	LevelBreached Level = "breached"
)

type Alert struct {
	Level     Level     `json:"level"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Service struct {
	logger         zerolog.Logger
	mu             sync.RWMutex
	telegramToken  string
	telegramChatID string
	discordWebhook string
	buffer         []Alert
	maxBuffer      int
}

func NewService(telegramToken, telegramChatID, discordWebhook string, logger zerolog.Logger) *Service {
	return &Service{
		logger:         logger.With().Str("component", "alerts").Logger(),
		telegramToken:  telegramToken,
		telegramChatID: telegramChatID,
		discordWebhook: discordWebhook,
		buffer:         make([]Alert, 0, 100),
		maxBuffer:      1000,
	}
}

func (s *Service) Send(level Level, title, message string) {
	alert := Alert{
		Level:     level,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
	}

	s.mu.Lock()
	s.buffer = append(s.buffer, alert)
	if len(s.buffer) > s.maxBuffer {
		s.buffer = s.buffer[len(s.buffer)-s.maxBuffer:]
	}
	s.mu.Unlock()

	prefix := "●"
	color := ""
	switch level {
	case LevelBreached:
		prefix = "■"
		color = "\033[31m"
	case LevelError:
		prefix = "✕"
		color = "\033[31m"
	case LevelWarn:
		prefix = "▲"
		color = "\033[33m"
	case LevelInfo:
		prefix = "●"
		color = "\033[32m"
	}

	s.logger.Info().Msgf("%s%s\033[0m [%s] %s: %s", color, prefix, level, title, message)

	go s.dispatch(alert)
}

func (s *Service) dispatch(alert Alert) {
	if s.telegramToken != "" && s.telegramChatID != "" {
		s.sendTelegram(alert)
	}
	if s.discordWebhook != "" {
		s.sendDiscord(alert)
	}
}

func (s *Service) sendTelegram(alert Alert) {
	icon := ""
	switch alert.Level {
	case LevelBreached, LevelError:
		icon = "🔴 "
	case LevelWarn:
		icon = "🟡 "
	case LevelInfo:
		icon = "🟢 "
	}

	text := fmt.Sprintf("%s*%s*\n%s", icon, alert.Title, alert.Message)

	body := map[string]string{
		"chat_id":    s.telegramChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	payload, _ := json.Marshal(body)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.telegramToken)
	http.Post(url, "application/json", bytes.NewReader(payload))
}

func (s *Service) sendDiscord(alert Alert) {
	color := 0x00ff00
	switch alert.Level {
	case LevelBreached, LevelError:
		color = 0xff0000
	case LevelWarn:
		color = 0xffa500
	}

	embed := map[string]any{
		"embeds": []map[string]any{{
			"title":       alert.Title,
			"description": alert.Message,
			"color":       color,
			"timestamp":   alert.Timestamp.Format(time.RFC3339),
		}},
	}
	payload, _ := json.Marshal(embed)
	http.Post(s.discordWebhook, "application/json", bytes.NewReader(payload))
}

func (s *Service) Recent() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Alert, len(s.buffer))
	copy(result, s.buffer)
	return result
}
