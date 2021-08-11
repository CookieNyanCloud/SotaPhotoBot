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
	newstick   = "newstick.png"

	watermarked = "watermarked.jpeg"
	webhook     = "https://photosotabot.herokuapp.com/"
)

const (
	cb = "1"
	cg = "2"
	cw = "3"
	sb = "4"
	sg = "5"
	sw = "6"
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


	type usersState struct {
		name string
		command string
	}
	users:= make([]usersState, 0, 25)

	for update := range updates {
		//println(update.Message.IsCommand())
		//println(users[0].id)
		//if update.Message == nil { // ignore any non-Message Updates
		//	continue
		//}
		//if (update.Message.Text) != "" {
		//	continue
		//}

		exist := -1
		userkol  := 0

		if (update.Message.IsCommand()) {
			println("comand")
			for i := range users {
				if users[i].name == update.Message.From.UserName {
					exist = i
					println("EXIST",exist)
				}
				userkol++
			}
			println(userkol)

			if exist != -1 {
				users[exist].command = update.Message.Command()
			} else {
				println("NOTEXIST",exist)
				nowuser := usersState{
					name:      update.Message.From.UserName,
					command: update.Message.Command(),
				}
				users = append(users, nowuser)
			}

			continue
		}

		if (update.Message.Text)!=""{
			continue
		}

		for i := range users {
			if users[i].name == update.Message.From.UserName {
				exist = i
			}

		}

		var varstick string
		var offsetX, offsetY int

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

		println(users[exist].command)
		com:= users[exist].command
		switch com {
		case cb:
			varstick = stikerbok
		case cg:
			varstick = stikercent
		case cw:
			varstick = stikercent
		case sb:
			varstick = stikerbok
		case sg:
			varstick = stikerbok
		case sw:
			varstick = stikerbok
		default:
			varstick = stikerbok

		}

		widthF, heightF := getImageDimension(filename)

		//wff:=float64(widthF)
		//hff:=float64(heightF)
		//newWidthf:=math.Sqrt(((wff)*(hff))/4)
		//newWidth:=int(math.Round(newWidthf))
		//src, err := imaging.Open(varstick)
		//src = imaging.Resize(src, newWidth, 0, imaging.Lanczos)
		//err = imaging.Save(src, newstick)
		//if err != nil {
		//	log.Fatalf("failed to save image: %v", err)
		//}

		//widthS, heightS := getImageDimension(newstick)
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
		//wmb, _ := os.Open(newstick)
		watermark, _ := png.Decode(wmb)
		defer wmb.Close()

		offset := image.Pt(offsetX, offsetY)
		b := img.Bounds()
		m := image.NewRGBA(b)
		draw.Draw(m, b, img, image.ZP, draw.Src)
		draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)

		imgw, _ := os.Create(watermarked)
		err = jpeg.Encode(imgw, m, &jpeg.Options{jpeg.DefaultQuality})
		if err != nil {
			err.Error()
			//return
		}
		defer imgw.Close()

		msg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, "watermarked.jpeg")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		users[exist].command = ""




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

