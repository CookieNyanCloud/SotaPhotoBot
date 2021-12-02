package main

import (
	"context"
	"errors"
	"github.com/cookienyancloud/photoSota/configs"
	"github.com/cookienyancloud/photoSota/driveService"
	"github.com/cookienyancloud/photoSota/tgBot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)


const (
	credFile = "api.json"
	find     = "найти"
	add      = "добавить"
)

func main() {

	var ctx = context.Background()

	conf, err := configs.InitConf()
	if err != nil {
		log.Fatalf("error init conf: %v\n", err)
	}

	users := make(map[int64]string)

	driveAcc, err := drive.NewService(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		log.Fatalf("Unable to parse credantials file: %v", err)
	}
	driveSrv := driveService.NewDriveService(driveAcc, conf.DrivePpl, conf.DriveZg)

	bot, updates, err := tgBot.StartBot(conf.TgToken)
	if err != nil {
		log.Fatalf("error connecting to bot: %v", err)
	}
	for update := range updates {

		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() && update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "выберите режим")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(find),
					tgbotapi.NewKeyboardButton(add),
				),
			)
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Text == find {
			users[update.Message.Chat.ID] = find
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "введите значение поиска")
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Text == add {
			users[update.Message.Chat.ID] = add
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "пришлите фото, тип, название, автора")
			_, _ = bot.Send(msg)
			continue
		}

		switch users[update.Message.Chat.ID] {

		case find:
			files, err := driveSrv.GetPhotos(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(find),
						tgbotapi.NewKeyboardButton(add),
					),
				)
				_, _ = bot.Send(msg)
				continue
			}
			for _, file := range files {
				msg := tgbotapi.NewDocument(update.Message.Chat.ID, &tgbotapi.FileReader{
					Name:   file.Name(),
					Reader: &file,
				})
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(find),
						tgbotapi.NewKeyboardButton(add),
					),
				)
				_, _ = bot.Send(msg)
			}

		case add:
			nameDir := strings.Split(update.Message.Caption, ",")
			if len(nameDir) != 3 || update.Message.Document.FileID == "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "не формат")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(find),
						tgbotapi.NewKeyboardButton(add),
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
			switch nameDir[0] {
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
				),
			)
			_, _ = bot.Send(msg)
			continue
		}

	}
}

func downloadFile(URL string) (os.File, error) {
	var file os.File
	response, err := http.Get(URL)
	if err != nil {
		return os.File{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return os.File{}, errors.New("received non 200 response code")
	}

	//Write the bytes to the fiel
	_, err = io.Copy(&file, response.Body)
	if err != nil {
		return os.File{}, err
	}

	return file, nil
}
