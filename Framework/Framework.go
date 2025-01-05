package Framework

import (
	"MessBot/Conf"
	"MessBot/Db"
	"MessBot/Message"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"time"
)

func LoopFramework(conf Conf.Conf) {
	for {
		err := Framework(conf)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
		}
	}
}

func Framework(conf Conf.Conf) error {
	db, err := Db.NewDB()
	bot, err := tgbotapi.NewBotAPI(conf.TgBotToken)
	if err != nil {
		return err
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updateConfig.AllowedUpdates = []string{"message", "callback_query"}
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		flag := false
		if update.CallbackQuery == nil && len(update.Message.Entities) != 0 {
			if update.Message.Entities[0].Type == "bot_command" {
				switch update.Message.Text {
				case "/start":
					mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Для начала работы с ботом необходимо ввести выданный вам пароль.\n\nДля подачи заявки на публикацию в канале, распишите ваш пост в обычном сообщении, так же вы можете прикрепить картинку. Мы сообщим вам как пост пройдёт модерацию")
					_, err = bot.Send(mes)
					if err != nil {
						return err
					}
				case "/info":
					mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Для начала работы с ботом необходимо ввести выданный вам пароль.\n\nДля подачи заявления на рассмотрение вашего поста, распишите ваш пост в обычном сообщении, так же вы можете прикрепить картинку. Мы сообщим вам как пост пройдут модерацию")
					_, err = bot.Send(mes)
					if err != nil {
						return err
					}
				default:
					mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда.")
					_, err = bot.Send(mes)
					if err != nil {
						return err
					}
				}
				continue
			}
		}
		if update.Message != nil {
			if update.Message.MigrateToChatID != 0 {
				if conf.ModersChat == update.Message.MigrateFromChatID {
					conf.ModersChat = update.Message.MigrateToChatID
					continue
				} else if conf.StreamersChat == update.Message.MigrateFromChatID {
					conf.StreamersChat = update.Message.MigrateToChatID
					continue
				}
			}
			if update.Message.From.ID == update.Message.Chat.ID {
				if update.Message.Text == conf.Key {
					if ok := CheckID(conf, update.Message.From.ID); !ok {
						conf.WhiteList = append(conf.WhiteList, update.Message.From.ID)
						err = NewConf(conf)
						if err != nil {
							return err
						}
						mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Пароль принят.")
						_, err = bot.Send(mes)
						if err != nil {
							return err
						}
						continue
					} else {
						mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Вам уже выданны права")
						_, err = bot.Send(mes)
						if err != nil {
							return err
						}
						//TODO: Для подачи заявления на рассмотрение вашего поста, распишите
						//ваш пост в обычном сообщении, так же вы можете прикрепить картинку.
						//Мы сообщим вам как пост пройдут модерацию
					}
				}
			}

			fmt.Printf("%+v\n", update.Message)
			for _, v := range conf.WhiteList {
				if update.Message.Chat.ID != update.Message.From.ID {
					continue
				}
				if v == update.Message.From.ID {
					flag = true
				}
			}
			if flag != true {
				mes := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав")
				_, err := bot.Send(mes)
				if err != nil {
					return err
				}
				continue
			}
			mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваш пост отправлен на рассмотрение в модерацию")
			_, err := bot.Send(mes)
			if err != nil {
				return err
			}
			if update.Message.Photo != nil {
				TempMess := Message.NewTempMess(update.Message.From.ID, update.Message.Photo[0].FileID, update.Message.Caption, update.Message.CaptionEntities)
				photo := Message.Photo(TempMess, conf.ModersChat)
				messid, err := Message.SendPhotoToModer(bot, photo)
				if err != nil {
					return err
				}
				TempMess.MessID = update.Message.MessageID
				if err = db.Add(messid, TempMess); err != nil {
					return err
				}
			}
			if update.Message.Photo == nil && update.Message.Text != "" {
				TempMess := Message.NewTempMess(update.Message.From.ID, "", update.Message.Text, update.Message.Entities)
				mess := Message.Mess(TempMess, conf.ModersChat)

				messid, err := Message.SendMessageToModer(bot, mess)
				if err != nil {
					return err
				}

				TempMess.MessID = update.Message.MessageID
				if err = db.Add(messid, TempMess); err != nil {
					return err
				}
			}
		}
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				return err
			}
			messId := update.CallbackQuery.Message.MessageID
			switch update.CallbackQuery.Data {
			case "Принять":
				TempMess, err := db.Get(messId)
				if err != nil {
					return err
				}
				if err := Message.AcceptCallBackToSender(bot, TempMess.SenderID); err != nil {
					return err
				}
				//TODO: Ненужный избыток. комменты делаются через настройки тг.
				//groupLink := fmt.Sprintf("https://t.me/c/%d/%d?thread=%d",
				//	-TempMess.SenderID, // Делаем ID положительным
				//	TempMess.MessID+1000000,
				//	TempMess.MessID,
				//)
				if len(TempMess.File) != 0 {
					photo := Message.Photo(TempMess, conf.StreamersChat)
					//photo.ReplyMarkup = Message.CommentButton(groupLink)
					err := Message.SendPhotoToStreamers(bot, photo)
					if err != nil {
						return err
					}
				} else {
					mess := Message.Mess(TempMess, conf.StreamersChat)
					//mess.ReplyMarkup = Message.CommentButton(groupLink)
					err := Message.SendMessageToStreamers(bot, mess)
					if err != nil {
						return err
					}
				}
				if err = db.Delete(messId); err != nil {
					return err
				}
			case "Отказать":
				TempMess, err := db.Get(messId)
				if err != nil {
					return err
				}
				if err := Message.RefuseCallBackToSender(bot, TempMess.SenderID); err != nil {
				}
				db.Delete(TempMess.MessID)
			}
			del := tgbotapi.NewDeleteMessage(conf.ModersChat, update.CallbackQuery.Message.MessageID)
			_, err := bot.Request(del)
			if err != nil {
				log.Printf("Error delet message. May be he is deleted! : %v\n", err)
				continue
			}
		}
	}
	return nil
}

func NewConf(con Conf.Conf) error {
	file, err := os.Create("MessConfig.cfg")
	if err != nil {
		return err
	}
	conBute, err := json.Marshal(con)
	if err != nil {
		return err
	}
	file.Write(conBute)
	file.Close()
	return nil
}

func CheckID(conf Conf.Conf, id int64) bool {
	ok := false
	for _, v := range conf.WhiteList {
		if v == id {
			ok = true
			break
		}
	}
	return ok
}
