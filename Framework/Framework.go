package Framework

import (
	"MessBot/Conf"
	"MessBot/Db"
	"MessBot/Message"
	"MessBot/Post"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
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
	PostMaps := make(map[int64]Post.PostCreateState)

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

		if update.CallbackQuery == nil && len(update.Message.Entities) != 0 {
			if update.Message.Entities[0].Type == "bot_command" {
				switch update.Message.Text {
				case "/newpost":
					if _, ok := PostMaps[update.Message.Chat.ID]; ok {
						mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не можите начать создавать новый пост пока не закончите/отмените предыдущи")
						if _, err := bot.Send(mes); err != nil {
							return err
						}
					} else {
						post := Post.PostCreateState{State: 0, RedactionFlag: true, SenderID: update.Message.Chat.ID, Data: "", StreamName: "", Game: "", Comments: "", Contact: "", ValuePersons: "", Duration: ""}
						PostMaps[update.Message.Chat.ID] = post
						if err := StagesDescription(bot, post.State, update.Message.Chat.ID); err != nil {
							return err
						}
					}
				case "/start":
					mes := tgbotapi.NewMessage(update.Message.Chat.ID, "\n\nДля подачи заявки на публикацию в канале, распишите ваш пост в обычном сообщении, так же вы можете прикрепить картинку. Мы сообщим вам как пост пройдёт модерацию")
					_, err = bot.Send(mes)
					if err != nil {
						return err
					}
				case "/info": //TODO Работа бота изменена, переписать информацию
					mes := tgbotapi.NewMessage(update.Message.Chat.ID, "\n\nДля подачи заявления на рассмотрение вашего поста, , так же вы можете прикрепить картинку. Мы сообщим вам как пост пройдут модерацию")
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
		//Миграция/изменение id чата
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

			//Отказ в публикации

			if update.Message.Chat.ID == conf.ModersChat {
				moderId := update.Message.From.ID
				messId, exist, err := db.GetRefuseModer(moderId)

				if err != nil && exist {
					return err
				}
				if exist {
					MessInfo, err := db.GetPost(messId)
					if err = db.DeleteRefuseModer(moderId); err != nil {
						return err
					}

					if err = db.DeletePost(messId); err != nil {
						return err
					}

					if err = Message.RefuseCallBackToSender(update.Message.Text, bot, MessInfo); err != nil {
						return err
					}

					del := tgbotapi.NewDeleteMessage(conf.ModersChat, MessInfo.MessIdInModerChat)
					_, err = bot.Request(del)
					if err != nil {
						log.Printf("Error delet message. May be he is deleted! : %v\n", err)
						continue
					}
				}
			}
			//обновление поста
			if update.Message.Chat.ID == update.Message.From.ID && update.Message.Chat.ID != conf.ModersChat && update.Message.Chat.ID != conf.StreamersChat {
				post := PostMaps[update.Message.Chat.ID]
				if post.RedactionFlag {
					if post, err = UpdateStagesPost(bot, post, update); err != nil {
						return err
					}

					post.RedactionFlag = false
					PostMaps[update.Message.Chat.ID] = post
				}
			}

		}
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				return err
			}

			switch update.CallbackQuery.Data {
			case "Отменить пост":
				mes := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пост отменён")
				if _, err := bot.Send(mes); err != nil {
					return err
				}
				delete(PostMaps, update.CallbackQuery.Message.Chat.ID)
			case "Принять":
				MessageInfo, err := db.GetPost(update.CallbackQuery.Message.MessageID)
				if err != nil {
					return err
				}

				if err = Message.AcceptCallBackToSender(bot, MessageInfo); err != nil {
					return err
				}

				if MessageInfo.PostHavePhoto == true {
					MessageInfo.PhotoPost.ReplyMarkup = nil
					_, err = Message.SendPhoto(bot, MessageInfo.PhotoPost, conf.StreamersChat)
					if err != nil {
						return err
					}
				} else {
					MessageInfo.MessagePost.ReplyMarkup = nil
					_, err = Message.SendMessage(bot, MessageInfo.MessagePost, conf.StreamersChat)
					if err != nil {
						return err
					}
				}

				if err = db.DeletePost(MessageInfo.MessIdInModerChat); err != nil {
					return err
				}
				if err = Message.DeleteMessage(bot, conf.ModersChat, MessageInfo.MessIdInModerChat); err != nil {
					return err
				}

			case "Отказать":
				MessageInfo, err := db.GetPost(update.CallbackQuery.Message.MessageID)
				if err != nil {
					return err
				}

				if err = db.AddRefuseModer(update.CallbackQuery.From.ID, MessageInfo.MessIdInModerChat); err != nil {
					return err
				}

				if MessageInfo.PostHavePhoto {
					if err = Message.EditPostWithPhoto(bot, conf.ModersChat, MessageInfo); err != nil {
						return err
					}
				} else {
					if err = Message.EditPost(bot, conf.ModersChat, MessageInfo); err != nil {
						return err
					}
				}

			case "Отмена":
				MessInfo, err := db.GetPost(update.CallbackQuery.Message.MessageID)
				if err != nil {
					return err
				}

				if err = db.DeleteRefuseModer(update.CallbackQuery.From.ID); err != nil {
					return err
				}

				if MessInfo.PostHavePhoto {
					if err = Message.EditPostBackWithPhoto(bot, conf.ModersChat, MessInfo); err != nil {
						return err
					}
				} else {
					if err = Message.EditPostBack(bot, conf.ModersChat, MessInfo); err != nil {
						return err
					}
				}

			case "Без объяснения причины":
				MessInfo, err := db.GetPost(update.CallbackQuery.Message.MessageID)
				if err = db.DeleteRefuseModer(update.CallbackQuery.From.ID); err != nil {
					return err
				}

				if err = db.DeletePost(MessInfo.MessIdInModerChat); err != nil {
					return err
				}

				if err = Message.RefuseCallBackToSender("Без объяснения причины", bot, MessInfo); err != nil {
					return err
				}

				del := tgbotapi.NewDeleteMessage(conf.ModersChat, MessInfo.MessIdInModerChat)
				_, err = bot.Request(del)
				if err != nil {
					log.Printf("Error delet message. May be he is deleted! : %v\n", err)
					continue
				}

			case "Пропустить":

				post := PostMaps[update.CallbackQuery.Message.Chat.ID]

				if post.State < 8 {
					post.State += 1
					StagesDescription(bot, post.State, update.CallbackQuery.Message.Chat.ID)

					post.RedactionFlag = true
					PostMaps[update.CallbackQuery.Message.Chat.ID] = post
				}
				if post.State == 8 {
					MessInf, err := ConstructAndSend(bot, post, conf.ModersChat)
					if err != nil {
						return err
					}

					db.AddPost(MessInf.MessIdInModerChat, MessInf)
					delete(PostMaps, update.CallbackQuery.Message.Chat.ID)
				}

			case "Редактировать":
				post := PostMaps[update.CallbackQuery.Message.Chat.ID]

				if err = StagesDescription(bot, post.State, update.CallbackQuery.Message.Chat.ID); err != nil {
					return err
				}

				post.RedactionFlag = true
				PostMaps[update.CallbackQuery.Message.Chat.ID] = post

			case "Продолжить":
				post := PostMaps[update.CallbackQuery.Message.Chat.ID]

				if post.State < 8 {
					post.State += 1
					StagesDescription(bot, post.State, update.CallbackQuery.Message.Chat.ID)

					post.RedactionFlag = true
					PostMaps[update.CallbackQuery.Message.Chat.ID] = post
				}
				if post.State == 8 {
					MessInf, err := ConstructAndSend(bot, post, conf.ModersChat)
					if err != nil {
						return err
					}

					db.AddPost(MessInf.MessIdInModerChat, MessInf)
					delete(PostMaps, update.CallbackQuery.Message.Chat.ID)
				}
			}
		}
	}
	return nil
}

func CreatePost(post Post.PostCreateState, update tgbotapi.Update) (Post.PostCreateState, bool) {
	switch post.State {
	case 0:
		if update.Message.Text == "" {
			return post, false
		}

		post.StreamName = update.Message.Text
		post.Entity = update.Message.Entities
	case 1:
		if update.Message.Text == "" {
			return post, false
		}

		post.Game = "\n\nИгра: "
		post.Entity = AddEntities(CreateText(post), post.Entity, update.Message.Entities)
		post.Game += update.Message.Text

	case 2:
		if update.Message.Text == "" {
			return post, false
		}

		post.Data = "\n\nДата/время: "
		post.Entity = AddEntities(CreateText(post), post.Entity, update.Message.Entities)
		post.Data += update.Message.Text

	case 3:
		if update.Message.Text == "" {
			return post, false
		}

		post.Duration = "\n\nПродолжительность: "
		post.Entity = AddEntities(CreateText(post), post.Entity, update.Message.Entities)
		post.Duration += update.Message.Text

	case 4:
		if update.Message.Text == "" {
			return post, false
		}

		post.ValuePersons = "\n\nКоличество участников: "
		post.Entity = AddEntities(CreateText(post), post.Entity, update.Message.Entities)
		post.ValuePersons += update.Message.Text

	case 5:
		if update.Message.Text == "" {
			return post, false
		}

		post.Comments = "\n\n"
		post.Entity = AddEntities(CreateText(post), post.Entity, update.Message.Entities)
		post.Comments += update.Message.Text

	case 6:
		if update.Message.Text == "" {
			return post, false
		}

		post.Contact = "\n\n"
		post.Entity = AddEntities(CreateText(post), post.Entity, update.Message.Entities)
		post.Contact += update.Message.Text

	case 7:
		if update.Message.Photo == nil {
			return post, false
		}

		post.PhotoFileID = update.Message.Photo[0].FileID
	}
	return post, true
}

func ApprovalOfChanges(post Post.PostCreateState, bot *tgbotapi.BotAPI) error {
	var err error

	if post.PhotoFileID != "" {
		mes := Message.Photo(post.Entity, CreateText(post), post.PhotoFileID)
		mes.BaseChat.ReplyMarkup = Message.ButtonsForUsers()
		_, err = Message.SendPhoto(bot, mes, post.SenderID)
	} else {
		mes := Message.Mess(post.Entity, CreateText(post))
		mes.BaseChat.ReplyMarkup = Message.ButtonsForUsers()
		_, err = Message.SendMessage(bot, mes, post.SenderID)
	}
	return err
}

func GeneratePostWithPhoto(post Post.PostCreateState) tgbotapi.PhotoConfig {
	return Message.Photo(post.Entity, CreateText(post), post.PhotoFileID)
}

func GeneratePost(post Post.PostCreateState) tgbotapi.MessageConfig {
	return Message.Mess(post.Entity, CreateText(post))
}

func CreateText(post Post.PostCreateState) string {
	return fmt.Sprintf("%v%v%v%v%v%v%v", post.StreamName, post.Game, post.Data, post.Duration, post.ValuePersons, post.Comments, post.Contact)
}

func StagesDescription(bot *tgbotapi.BotAPI, stage int, id int64) error {
	var err error
	switch stage {
	case 0:
		mes := tgbotapi.NewMessage(id, "Отлично!\n\nВведите название стрима. (Опционально)")
		mes.BaseChat.ReplyMarkup = Message.ButtonContinue()
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 1:
		mes := tgbotapi.NewMessage(id, "Введите название игры.")
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 2:
		mes := tgbotapi.NewMessage(id, "Введите дату и время проведения события.")
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 3:
		mes := tgbotapi.NewMessage(id, "Сколько по длительности будет проходить событие. (Опционально)")
		mes.BaseChat.ReplyMarkup = Message.ButtonContinue()
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 4:
		mes := tgbotapi.NewMessage(id, "Сколько участников хотите приглосить. ")
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 5:
		mes := tgbotapi.NewMessage(id, "Добавьте комментарий к посту. (Опционально)")
		mes.BaseChat.ReplyMarkup = Message.ButtonContinue()
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 6:
		mes := tgbotapi.NewMessage(id, "Введите контакты для обратной связи. (Опционально)")
		mes.BaseChat.ReplyMarkup = Message.ButtonContinue()
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	case 7:
		mes := tgbotapi.NewMessage(id, "Прикрепите фото для вашей публикации. (Опционально)")
		mes.BaseChat.ReplyMarkup = Message.ButtonContinue()
		_, err = bot.Send(mes)
		if err != nil {
			return err
		}
	}
	return err
}

func UpdateStagesPost(bot *tgbotapi.BotAPI, post Post.PostCreateState, update tgbotapi.Update) (Post.PostCreateState, error) {
	var err error
	post, ok := CreatePost(post, update)

	if !ok {
		if post.State < 7 {
			mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Это не текст.")
			_, err = bot.Send(mes)
			if err != nil {
				return post, err
			}

		} else {
			mes := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не прикрепили фото.")
			_, err = bot.Send(mes)
			if err != nil {
				return post, err
			}

		}
		return post, err
	}

	err = ApprovalOfChanges(post, bot)
	if err != nil {
		return post, err
	}

	return post, err
}

func ConstructAndSend(bot *tgbotapi.BotAPI, post Post.PostCreateState, moderchat int64) (Message.MessageInfo, error) {
	MessageInfo := Message.MessageInfo{SenderID: post.SenderID}
	mes := tgbotapi.NewMessage(post.SenderID, "Ваш пост отправлен на рассмотрение в модерацию")
	_, err := bot.Send(mes)
	if err != nil {
		return MessageInfo, err
	}

	if post.PhotoFileID != "" {
		MessageInfo.PostHavePhoto = true
		MessageInfo.PhotoPost = GeneratePostWithPhoto(post)

		MessageInfo.MessIdInSenderChat, err = Message.SendPhoto(bot, MessageInfo.PhotoPost, MessageInfo.SenderID)
		if err != nil {
			return MessageInfo, err
		}

		MessageInfo.PhotoPost.ReplyMarkup = Message.Buttons()
		MessageInfo.MessIdInModerChat, err = Message.SendPhoto(bot, MessageInfo.PhotoPost, moderchat)
		if err != nil {
			return MessageInfo, err
		}

	} else {
		MessageInfo.PostHavePhoto = false
		MessageInfo.MessagePost = GeneratePost(post)

		MessageInfo.MessIdInSenderChat, err = Message.SendMessage(bot, MessageInfo.MessagePost, MessageInfo.SenderID)
		if err != nil {
			return MessageInfo, err
		}

		MessageInfo.MessagePost.ReplyMarkup = Message.Buttons()
		MessageInfo.MessIdInModerChat, err = Message.SendMessage(bot, MessageInfo.MessagePost, moderchat)
		if err != nil {
			return MessageInfo, err
		}
	}

	return MessageInfo, err
}

func AddEntities(text string, PostEntities, UpdateEntities []tgbotapi.MessageEntity) []tgbotapi.MessageEntity {
	for _, v := range UpdateEntities {
		v.Offset += len(text) + 1
		PostEntities = append(PostEntities, v)
	}
	return PostEntities
}
