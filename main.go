package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
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
	stikercent = "photo/cetn.png"
	stikerbok  = "photo/chert.png"
	newjpg      = "photo/new.jpg"
	watermarked = "photo/watermarked.jpeg"
	//stiker      = "photo/stiker.png"
	//newstick    = "photo/newstick.png"

)

//const (
//	//webhook     = "https://photosotabot.herokuapp.com/"
//	webhook = "https://34.116.203.162/"
//	//cert = "cert.pem"
//	//key = "key.pem"
//	//addr = "0.0.0.0:"
//	//port = "8443"
//)

const (
	cb  = "1"
	cg  = "2"
	cw  = "3"
	sb  = "4"
	sg  = "5"
	sw  = "6"
	sch = "7"
)

func main() {

	_ = godotenv.Load()
	token:= os.Getenv("TOKEN_B")
	println(token)
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	type usersState struct {
		name    string
		command string
	}
	users := make([]usersState, 0, 25)

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
		userkol := 0

		if update.Message.IsCommand() {
			println("command")
			for i := range users {
				if users[i].name == update.Message.From.UserName {
					exist = i
					println("EXIST", exist)
				}
				userkol++
			}
			println(userkol)

			if exist != -1 {
				users[exist].command = update.Message.Command()
			} else {
				println("NOTEXIST", exist)
				nowuser := usersState{
					name:    update.Message.From.UserName,
					command: update.Message.Command(),
				}
				users = append(users, nowuser)
			}

			continue
		}

		for i := range users {
			if users[i].name == update.Message.From.UserName {
				exist = i
			}

		}

		if (update.Message.Text) != "" && users[exist].command != sch {
			continue
		}

		var varstick string
		var offsetX, offsetY int

		com := users[exist].command
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
		case sch:
			out, _ := search(update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, out)
			msg.ReplyToMessageID = update.Message.MessageID
			_, _ = bot.Send(msg)


			continue
		default:
			varstick = stikerbok

		}

		leng := len(*update.Message.Photo)
		phUrl, err := bot.GetFileDirectURL((*update.Message.Photo)[leng-1].FileID)
		filename := newjpg
		err = DownloadFile(phUrl, filename)
		if err != nil {
			err.Error()
		}
		imgb, _ := os.Open(filename)
		img, _ := jpeg.Decode(imgb)
		defer imgb.Close()

		println(users[exist].command)


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
		err = jpeg.Encode(imgw, m, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			err.Error()
			//return
		}
		defer imgw.Close()

		msg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, watermarked)
		msg.ReplyToMessageID = update.Message.MessageID
		_, _ = bot.Send(msg)
		users[exist].command = ""

	}
}

func search(input string) (string, error) {
	//resp, err := http.Get("https://www.googleapis.com/drive/v3/files")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//sb := string(body)
	return input, nil

}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	IImage, _, err := image.DecodeConfig(file)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return IImage.Width, IImage.Height
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

func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(token)
}

//
//func main() {
//	//port := os.Getenv("PORT")
//	//port := port
//	//go func() {
//	//	log.Fatal(http.ListenAndServe(":"+port, nil))
//	//}()
//
//	bot, err := tgbotapi.NewBotAPI(token)
//	if err != nil {
//		log.Fatal("creating bot:", err)
//	}
//	log.Println("bot created")
//
//	if _, err = bot.SetWebhook(tgbotapi.NewWebhook(webhook+port)); err != nil {
//		log.Fatalf("setting webhook %v: %v", webhook, err)
//	}
//	log.Println("webhook set")
//	info, err := bot.GetWebhookInfo()
//	if err != nil {
//		log.Fatal(err)
//	}
//	if info.LastErrorDate != 0 {
//		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
//	}
//
//	updates := bot.ListenForWebhook("/"+ bot.Token)
//	//go http.ListenAndServeTLS(addr, cert, key, nil)
//	go http.ListenAndServe(addr+port,nil)
//
//
//	type usersState struct {
//		name string
//		command string
//	}
//	users:= make([]usersState, 0, 25)
//
//	for update := range updates {
//		//println(update.Message.IsCommand())
//		//println(users[0].id)
//		//if update.Message == nil { // ignore any non-Message Updates
//		//	continue
//		//}
//		//if (update.Message.Text) != "" {
//		//	continue
//		//}
//
//		exist := -1
//		userkol  := 0
//
//		if update.Message.IsCommand() {
//			println("command")
//			for i := range users {
//				if users[i].name == update.Message.From.UserName {
//					exist = i
//					println("EXIST",exist)
//				}
//				userkol++
//			}
//			println(userkol)
//
//			if exist != -1 {
//				users[exist].command = update.Message.Command()
//			} else {
//				println("NOTEXIST",exist)
//				nowuser := usersState{
//					name:      update.Message.From.UserName,
//					command: update.Message.Command(),
//				}
//				users = append(users, nowuser)
//			}
//
//			continue
//		}
//
//		if (update.Message.Text)!=""{
//			continue
//		}
//
//		for i := range users {
//			if users[i].name == update.Message.From.UserName {
//				exist = i
//			}
//
//		}
//
//		var varstick string
//		var offsetX, offsetY int
//
//		leng := len(*update.Message.Photo)
//		phUrl, err := bot.GetFileDirectURL((*update.Message.Photo)[leng-1].FileID)
//		filename := "new.jpg"
//		err = DownloadFile(phUrl, filename)
//		if err != nil {
//			err.Error()
//		}
//		imgb, _ := os.Open(filename)
//		img, _ := jpeg.Decode(imgb)
//		defer imgb.Close()
//
//		println(users[exist].command)
//		com:= users[exist].command
//		switch com {
//		case cb:
//			varstick = stikerbok
//		case cg:
//			varstick = stikercent
//		case cw:
//			varstick = stikercent
//		case sb:
//			varstick = stikerbok
//		case sg:
//			varstick = stikerbok
//		case sw:
//			varstick = stikerbok
//		default:
//			varstick = stikerbok
//
//		}
//
//		widthF, heightF := getImageDimension(filename)
//
//		//wff:=float64(widthF)
//		//hff:=float64(heightF)
//		//newWidthf:=math.Sqrt(((wff)*(hff))/4)
//		//newWidth:=int(math.Round(newWidthf))
//		//src, err := imaging.Open(varstick)
//		//src = imaging.Resize(src, newWidth, 0, imaging.Lanczos)
//		//err = imaging.Save(src, newstick)
//		//if err != nil {
//		//	log.Fatalf("failed to save image: %v", err)
//		//}
//
//		//widthS, heightS := getImageDimension(newstick)
//		widthS, heightS := getImageDimension(varstick)
//
//
//
//
//		switch varstick {
//		case stikerbok:
//			offsetX = 0
//			offsetY = heightF - heightS
//		case stikercent:
//			offsetX = (widthF / 2) - (widthS / 2)
//			offsetY = (heightF / 2) - (heightS / 2)
//		default:
//			offsetX = 0
//			offsetY = heightF - heightS
//		}
//		wmb, _ := os.Open(varstick)
//		//wmb, _ := os.Open(newstick)
//		watermark, _ := png.Decode(wmb)
//		defer wmb.Close()
//
//		offset := image.Pt(offsetX, offsetY)
//		b := img.Bounds()
//		m := image.NewRGBA(b)
//		draw.Draw(m, b, img, image.ZP, draw.Src)
//		draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)
//
//		imgw, _ := os.Create(watermarked)
//		err = jpeg.Encode(imgw, m, &jpeg.Options{jpeg.DefaultQuality})
//		if err != nil {
//			err.Error()
//			//return
//		}
//		defer imgw.Close()
//
//		msg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, "watermarked.jpeg")
//		msg.ReplyToMessageID = update.Message.MessageID
//		bot.Send(msg)
//		users[exist].command = ""
//
//
//
//
//	}
//}
//
//func getImageDimension(imagePath string) (int, int) {
//	file, err := os.Open(imagePath)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "%v\n", err)
//	}
//
//	image, _, err := image.DecodeConfig(file)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
//	}
//	return image.Width, image.Height
//}
//
//func DownloadFile(URL, fileName string) error {
//	response, err := http.Get(URL)
//	if err != nil {
//		return err
//	}
//	defer response.Body.Close()
//
//	if response.StatusCode != 200 {
//		return errors.New("Received non 200 response code")
//	}
//	//Create a empty file
//	file, err := os.Create(fileName)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	//Write the bytes to the fiel
//	_, err = io.Copy(file, response.Body)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func getClient(config *oauth2.Config) *http.Client {
//	// The file token.json stores the user's access and refresh tokens, and is
//	// created automatically when the authorization flow completes for the first
//	// time.
//	tokFile := "token.json"
//	tok, err := tokenFromFile(tokFile)
//	if err != nil {
//		tok = getTokenFromWeb(config)
//		saveToken(tokFile, tok)
//	}
//	return config.Client(context.Background(), tok)
//}
//
//func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
//	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
//	fmt.Printf("Go to the following link in your browser then type the "+
//		"authorization code: \n%v\n", authURL)
//
//	var authCode string
//	if _, err := fmt.Scan(&authCode); err != nil {
//		log.Fatalf("Unable to read authorization code %v", err)
//	}
//
//	tok, err := config.Exchange(context.TODO(), authCode)
//	if err != nil {
//		log.Fatalf("Unable to retrieve token from web %v", err)
//	}
//	return tok
//}
//
//func tokenFromFile(file string) (*oauth2.Token, error) {
//	f, err := os.Open(file)
//	if err != nil {
//		return nil, err
//	}
//	defer f.Close()
//	tok := &oauth2.Token{}
//	err = json.NewDecoder(f).Decode(tok)
//	return tok, err
//}
//
//func saveToken(path string, token *oauth2.Token) {
//	fmt.Printf("Saving credential file to: %s\n", path)
//	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
//	if err != nil {
//		log.Fatalf("Unable to cache oauth token: %v", err)
//	}
//	defer f.Close()
//	json.NewEncoder(f).Encode(token)
//}
