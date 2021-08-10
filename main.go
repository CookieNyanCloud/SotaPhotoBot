package main

import (
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	stiker      = "stiker.png"
	stikercent  = "cetn.png"
	stikerbok   = "chert.png"
	watermarked = "watermarked.jpeg"
	webhook     = "https://photosotabot.herokuapp.com/"
)

func main() {

	port := os.Getenv("PORT")

	go func() {
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("creating bot:", err)
	}
	log.Println("bot created")

	if _, err = bot.SetWebhook(tgbotapi.NewWebhook(webhook)); err != nil {
		log.Fatalf("setting webhook %v: %v", webhook, err)
	}
	log.Println("webhook set")

	updates := bot.ListenForWebhook("/")

	//bot.Debug = true
	//log.Printf("Authorized on account %s", bot.Self.UserName)
	//u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60
	//updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if (update.Message.Text) != "" {
			continue
		}

		leng := len(*update.Message.Photo)
		phUrl, err := bot.GetFileDirectURL((*update.Message.Photo)[leng-1].FileID)
		filename := "new.jpg"
		err = DownloadFile(phUrl, filename)
		if err != nil {
			err.Error()
		}

		imgb, _ := os.Open(filename)
		img, _ := jpeg.Decode(imgb)
		defer imgb.Close()

		var varstick string
		var offsetX, offsetY int
		pos := update.Message.Caption
		if pos == "1" {
			varstick = stikerbok
		} else if pos == "2" {
			varstick = stikercent
		} else {
			varstick = stikerbok
		}

		widthF, heightF := getImageDimension(filename)
		widthS, heightS := getImageDimension(varstick)

		switch varstick {
		case stikerbok:
			offsetX = 0
			offsetY = heightF - heightS
		case stikercent:
			offsetX = (widthF / 2) - (widthS / 2)
			offsetY = (heightF / 2) - (heightS / 2)
		default:
			offsetX = 0
			offsetY = heightF - heightS
		}

		wmb, _ := os.Open(varstick)
		watermark, _ := png.Decode(wmb)
		defer wmb.Close()

		offset := image.Pt(offsetX, offsetY)
		b := img.Bounds()
		m := image.NewRGBA(b)
		draw.Draw(m, b, img, image.ZP, draw.Src)
		draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)

		imgw, _ := os.Create(watermarked)
		//png.Encode(imgw,m)
		err = jpeg.Encode(imgw, m, &jpeg.Options{jpeg.DefaultQuality})
		if err != nil {
			err.Error()
			//return
		}
		defer imgw.Close()

		//msg:= tgbotapi.NewPhotoUpload(update.Message.Chat.ID,"watermark-new-stiker.png")
		msg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, "watermarked.jpeg")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)

		//prevname = filename
	}
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

func DownloadFile(URL, fileName string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
