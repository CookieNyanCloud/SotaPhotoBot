package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cookienyancloud/photoSota/configs"
	"github.com/cookienyancloud/photoSota/driveService"
	"github.com/cookienyancloud/photoSota/tgBot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	credFile = "driveapisearch.json"
	find     = "найти"
	add      = "добавить"
	addUser  = "добавить пользователя"
)

func main() {

	var ctx = context.Background()
	users, err := configs.GetUsers()
	if err != nil {
		log.Fatalf("error init users: %v\n", err)
	}
	conf, err := configs.InitConf()
	if err != nil {
		log.Fatalf("error init conf: %v\n", err)
	}

	driveAcc, err := drive.NewService(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		log.Fatalf("Unable to parse credantials file: %v", err)
	}
	driveSrv := driveService.NewDriveService(driveAcc)
	bot, updates, err := tgBot.StartBot(conf.TgToken)
	if err != nil {
		log.Fatalf("error connecting to bot: %v", err)
	}
	for update := range updates {

		if update.Message == nil {
			continue
		}
		_, ok := users[update.Message.Chat.UserName]
		if !ok {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "нет доступа")
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.IsCommand() && update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "выберите режим")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(find),
					tgbotapi.NewKeyboardButton(add),
					tgbotapi.NewKeyboardButton(addUser),
				),
			)
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Text == find {
			users[update.Message.Chat.UserName] = find
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "введите значение поиска")
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Text == add {
			users[update.Message.Chat.UserName] = add
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "пришлите фото, тип, название, автора")
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Text == addUser {
			users[update.Message.Chat.UserName] = addUser
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "пришлите ник")
			_, _ = bot.Send(msg)
			continue
		}

		switch users[update.Message.Chat.UserName] {

		case find:
			files, err := driveSrv.GetPhotos(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(find),
						tgbotapi.NewKeyboardButton(add),
						tgbotapi.NewKeyboardButton(addUser),

					),
				)
				_, _ = bot.Send(msg)
				continue
			}
			for _, resp := range files {
				msg := tgbotapi.NewDocument(update.Message.Chat.ID, &tgbotapi.FileReader{
					Name:   "filefromdrive.jpg",
					Reader: resp.Body,
				})
				_, _ = bot.Send(msg)
				err := resp.Body.Close()
				if err != nil {
					log.Printf("err closing body: %v", err)
				}
			}

		case add:
			nameDir := strings.Split(update.Message.Caption, ",")
			if len(nameDir) != 3 || update.Message.Document == nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "не формат")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(find),
						tgbotapi.NewKeyboardButton(add),
						tgbotapi.NewKeyboardButton(addUser),

					),
				)
				_, _ = bot.Send(msg)
				continue
			}
			s, err := bot.GetFileDirectURL(update.Message.Document.FileID)
			if err != nil {
				log.Printf("err gettind URL")
				return
			}
			file, err := downloadFile(s)
			if err != nil {
				log.Printf("err getting file from URL")
				return
			}
			var folder string
			switch strings.ToLower(nameDir[0]) {
			case "з":
				folder = conf.DriveZg
			case "л":
				folder = conf.DrivePpl
			default:
				folder = conf.DrivePpl
			}
			name := nameDir[1] + nameDir[2]
			err = driveSrv.SendPhotos(name, folder, file)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(find),
						tgbotapi.NewKeyboardButton(add),
						tgbotapi.NewKeyboardButton(addUser),

					),
				)
				_, _ = bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "готово")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(find),
					tgbotapi.NewKeyboardButton(add),
					tgbotapi.NewKeyboardButton(addUser),

				),
			)
			_, _ = bot.Send(msg)
		case addUser:
			name := strings.ReplaceAll(update.Message.Text, "@", "")
			err := configs.AddUser(users, update.Message.From.UserName, name)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintln("in case pushUser:", err))
				_, _ = bot.Send(msg)
				fmt.Println("in case pushUser:", err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "добавлен")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(find),
					tgbotapi.NewKeyboardButton(add),
					tgbotapi.NewKeyboardButton(addUser),

				),
			)
			_, _ = bot.Send(msg)

		}
	}
}

func downloadFile(URL string) (*http.Response, error) {
	//var file os.File
	response, err := http.Get(URL)
	if err != nil {
		//return os.File{}, err
		return nil, err
	}
	//defer response.Body.Close()

	if response.StatusCode != 200 {
		//return os.File{}, errors.New("received non 200 response code")
		return nil, errors.New("received non 200 response code")
	}
	//
	////Write the bytes to the fiel
	//_, err = io.Copy(&file, response.Body)
	//if err != nil {
	//	return os.File{}, err
	//}

	return response, nil
}
