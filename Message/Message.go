package Message

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type MessageInfo struct {
	MessIdInSenderChat int
	MessIdInModerChat  int
	PhotoPost          tgbotapi.PhotoConfig
	MessagePost        tgbotapi.MessageConfig
	SenderID           int64
	PostHavePhoto      bool
}

func SendPhoto(bot *tgbotapi.BotAPI, photo tgbotapi.PhotoConfig, chatid int64) (int, error) {
	photo.BaseChat.ChatID = chatid
	//photo.BaseChat.ReplyMarkup = Buttons()
	messInf, err := bot.Send(photo)
	return messInf.MessageID, err
}

func SendMessage(bot *tgbotapi.BotAPI, mess tgbotapi.MessageConfig, chatid int64) (int, error) {
	mess.BaseChat.ChatID = chatid
	//mess.BaseChat.ReplyMarkup = Buttons()
	messInf, err := bot.Send(mess)
	return messInf.MessageID, err
}

func Photo(entities []tgbotapi.MessageEntity, text, fileID string) tgbotapi.PhotoConfig {
	photo := tgbotapi.PhotoConfig{
		Caption:         text,
		CaptionEntities: entities,
		BaseFile: tgbotapi.BaseFile{
			File: tgbotapi.FileID(fileID),
		},
	}
	return photo
}

func Mess(entities []tgbotapi.MessageEntity, text string) tgbotapi.MessageConfig {
	mess := tgbotapi.MessageConfig{
		Text:     text,
		Entities: entities,
	}
	return mess
}

func Buttons() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Принять", "Принять"),
		tgbotapi.NewInlineKeyboardButtonData("Отказать", "Отказать")))
}

func ButtonsForUsers() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Редактировать", "Редактировать"),
		tgbotapi.NewInlineKeyboardButtonData("Продолжить", "Продолжить")), tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отменить пост", "Отменить пост")))
}

func ButtonContinue() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Пропустить", "Пропустить")), tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отменить пост", "Отменить пост")))
}

func ButtonRefuse() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отмена", "Отмена"),
		tgbotapi.NewInlineKeyboardButtonData("Без объяснения причины", "Без объяснения причины")))
}

func AcceptCallBackToSender(bot *tgbotapi.BotAPI, Messinfo MessageInfo) error {
	mess := tgbotapi.NewMessage(Messinfo.SenderID, "Ваш пост был принят к публикации.")
	mess.ReplyToMessageID = Messinfo.MessIdInSenderChat

	_, err := bot.Send(mess)
	return err
}

func RefuseCallBackToSender(text string, bot *tgbotapi.BotAPI, Messinfo MessageInfo) error {
	mess := tgbotapi.NewMessage(Messinfo.SenderID, fmt.Sprintf("В публикации было отказанно.\n\n Комментарий модерации:%v", text))
	mess.ReplyToMessageID = Messinfo.MessIdInSenderChat

	_, err := bot.Send(mess)
	return err
}

func EditPost(bot *tgbotapi.BotAPI, ModerChat int64, Messinfo MessageInfo) error {
	mess := tgbotapi.NewEditMessageTextAndMarkup(ModerChat, Messinfo.MessIdInModerChat,
		Messinfo.MessagePost.Text+"\nНапишите причину отказа.", ButtonRefuse())
	mess.Entities = Messinfo.MessagePost.Entities

	_, err := bot.Send(mess)
	return err
}

func EditPostWithPhoto(bot *tgbotapi.BotAPI, ModerChat int64, Messinfo MessageInfo) error {
	button := ButtonRefuse()
	photo := tgbotapi.NewEditMessageCaption(ModerChat, Messinfo.MessIdInModerChat, Messinfo.PhotoPost.Caption+"\nНапишите причину отказа")
	photo.BaseEdit.ReplyMarkup = &button
	photo.CaptionEntities = Messinfo.PhotoPost.CaptionEntities

	_, err := bot.Send(photo)
	return err
}

func EditPostBack(bot *tgbotapi.BotAPI, ModerChat int64, Messinfo MessageInfo) error {
	mess := tgbotapi.NewEditMessageTextAndMarkup(ModerChat, Messinfo.MessIdInModerChat,
		Messinfo.MessagePost.Text, Buttons())
	mess.Entities = Messinfo.MessagePost.Entities

	_, err := bot.Send(mess)
	return err
}

func EditPostBackWithPhoto(bot *tgbotapi.BotAPI, ModerChat int64, Messinfo MessageInfo) error {
	button := Buttons()
	photo := tgbotapi.NewEditMessageCaption(ModerChat, Messinfo.MessIdInModerChat, Messinfo.PhotoPost.Caption)
	photo.BaseEdit.ReplyMarkup = &button
	photo.CaptionEntities = Messinfo.PhotoPost.CaptionEntities

	_, err := bot.Send(photo)
	return err
}

func DeleteMessage(bot *tgbotapi.BotAPI, chat int64, id int) error {
	del := tgbotapi.NewDeleteMessage(chat, id)
	_, err := bot.Request(del)
	if err != nil {
		log.Printf("Error delet message. May be he is deleted! : %v\n", err)
		return err
	}
	return err
}
