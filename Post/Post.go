package Post

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type PostCreateState struct {
	RedactionFlag bool
	SenderID      int64
	State         int
	Data          string
	StreamName    string
	Game          string
	Comments      string
	Contact       string
	PhotoFileID   string
	ValuePersons  string
	Duration      string

	Entity []tgbotapi.MessageEntity
}
