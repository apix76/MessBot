package Message

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TempMess struct {
	MessID          int
	File            string
	Caption         string
	SenderID        int64
	CaptionEntities []tgbotapi.MessageEntity
}

func SendPhotoToModer(bot *tgbotapi.BotAPI, photo tgbotapi.PhotoConfig) (int, error) {
	photo.BaseChat.ReplyMarkup = Buttons()
	messInf, err := bot.Send(photo)
	return messInf.MessageID, err
}

func SendMessageToModer(bot *tgbotapi.BotAPI, mess tgbotapi.MessageConfig) (int, error) {
	mess.BaseChat.ReplyMarkup = Buttons()
	messInf, err := bot.Send(mess)
	return messInf.MessageID, err
}

func SendPhotoToStreamers(bot *tgbotapi.BotAPI, photo tgbotapi.PhotoConfig) error {
	_, err := bot.Send(photo)
	return err
}

func SendMessageToStreamers(bot *tgbotapi.BotAPI, mess tgbotapi.MessageConfig) error {
	_, err := bot.Send(mess)
	return err
}

func Photo(TempMess TempMess, ChatId int64) tgbotapi.PhotoConfig {
	photo := tgbotapi.PhotoConfig{
		Caption:         TempMess.Caption,
		CaptionEntities: TempMess.CaptionEntities,
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID: ChatId,
			},
			File: tgbotapi.FileID(TempMess.File),
		},
	}
	return photo
}

func Mess(TempMess TempMess, ChatID int64) tgbotapi.MessageConfig {
	mess := tgbotapi.MessageConfig{
		Text:     TempMess.Caption,
		Entities: TempMess.CaptionEntities,
		BaseChat: tgbotapi.BaseChat{
			ChatID: ChatID,
		},
	}
	return mess
}

func Buttons() interface{} {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Принять", "Принять"),
		tgbotapi.NewInlineKeyboardButtonData("Отказать", "Отказать")))
}

func NewTempMess(SenderID int64, FileID, Caption string, CaptionEntities []tgbotapi.MessageEntity) TempMess {
	return TempMess{
		SenderID:        SenderID,
		File:            FileID,
		Caption:         Caption,
		CaptionEntities: CaptionEntities,
	}
}

func AcceptCallBackToSender(bot *tgbotapi.BotAPI, userId int64) error {
	mess := tgbotapi.NewMessage(userId, "Ваш пост был принят к публикации.")
	_, err := bot.Send(mess)
	return err
}

func RefuseCallBackToSender(bot *tgbotapi.BotAPI, userId int64) error {
	mess := tgbotapi.NewMessage(userId, "В публикации было отказанно.")
	_, err := bot.Send(mess)
	return err
}
