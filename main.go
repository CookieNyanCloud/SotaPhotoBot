package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cookienyancloud/photoSota/arch"
	"github.com/cookienyancloud/photoSota/sotatgbot"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	stikercent  = "photo/cetn.png"
	stikerbok   = "photo/chert.png"
	newjpg      = "photo/new.jpg"
	watermarked = "photo/watermarked.jpeg"
	//stiker      = "photo/stiker.png"
	//newstick    = "photo/newstick.png"
	urlPhotoGet = "http://localhost:8090/getphoto"
	urlPhotoSet = "http://localhost:8090/sendphoto"
)

const (
	cb  = "1"
	cg  = "2"
	cw  = "3"
	sb  = "4"
	sg  = "5"
	sw  = "6"
	sch = "7"
	pht = "8"
)

const (
	tokenA = "TOKEN_A"
	tokenB = "TOKEN_B"
)

type UsersState struct {
	Name    string
	Command string
}

func main() {

	bot, updates := sotatgbot.StartSotaBot(tokenB)

	users := make([]UsersState, 0, 25)

	for update := range updates {

		exist := -1
		userkol := 0
		if update.Message.IsCommand() {
			println("command")
			for i := range users {
				if users[i].Name == update.Message.From.UserName {
					exist = i
					println("EXIST", exist)
				}
				userkol++
			}
			println(userkol)

			if exist != -1 {
				users[exist].Command = update.Message.Command()
			} else {
				println("NOTEXIST", exist)
				nowuser := UsersState{
					Name:    update.Message.From.UserName,
					Command: update.Message.Command(),
				}
				users = append(users, nowuser)
			}

			continue
		}

		for i := range users {
			if users[i].Name == update.Message.From.UserName {
				exist = i
			}

		}

		if (update.Message.Text) != "" && users[exist].Command != sch {
			continue
		}

		var varstick string
		var offsetX, offsetY int

		com := users[exist].Command
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
			postBody, _ := json.Marshal(map[string]string{
				"name": update.Message.Text,
			})
			responseBody := bytes.NewBuffer(postBody)
			resp, err := http.Post(urlPhotoGet, "application/json", responseBody)
			if err != nil {
				fmt.Println(err.Error())
			}
			//body, err := ioutil.ReadAll(resp.Body)
			//if err != nil {
			//	fmt.Println(err.Error())
			//}
			out, err := os.Create("arch.zip")
			if err != nil {
				fmt.Println(err.Error())
			}
			_, err = io.Copy(out, resp.Body)
			if err = out.Close(); err != nil {
				fmt.Println(err.Error())
			}
			if err = resp.Body.Close(); err != nil {
				fmt.Println(err.Error())

			}
			files, err := arch.Unzip("arch.zip", "fromZip")
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(files)

			for _, v := range files {
				//msg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, v)
				//msg.ReplyToMessageID = update.Message.MessageID
				//_, _ = bot.Send(msg)
				msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, v)
				msg.ReplyToMessageID = update.Message.MessageID
				_, _ = bot.Send(msg)
			}
			users[exist].Command = ""

			err = arch.MyDelete("arch.zip")
			if err != nil {
				fmt.Println(err.Error())
			}
			err = arch.AllDelete(files)
			if err != nil {
				fmt.Println(err.Error())
			}

			continue
		case pht:

			data := strings.Split(update.Message.Caption, ", ")
			phUrl, err := bot.GetFileDirectURL(update.Message.Document.FileID)
			err = DownloadFile(phUrl, data[0]+".jpg")
			if err != nil {
				fmt.Println("err downloading: ", err.Error())
			}
			file, err := os.Open(data[0] + ".jpg")
			if err != nil {
				fmt.Println("err opening file: ", err.Error())
			}
			fileContents, err := ioutil.ReadAll(file)
			if err != nil {
				fmt.Println("err reading file: ", err.Error())
			}
			fi, err := file.Stat()
			if err != nil {
				fmt.Println("err getting stat: ", err.Error())
			}
			if err = file.Close(); err != nil {
				fmt.Println("err closing file: ", err.Error())
			}

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", fi.Name())
			if err != nil {
				fmt.Println("err creatFromFile file: ", err.Error())
			}
			if _, err = part.Write(fileContents); err != nil {
				fmt.Println("err creatFromFile: ", err.Error())
			}

			err = writer.WriteField("dirType", data[2])
			if err != nil {
				fmt.Println("err dirType: ", err.Error())
			}
			err = writer.WriteField("author",data[1])
			if err != nil {
				fmt.Println("err author: ", err.Error())
			}

			err = writer.Close()
			if err != nil {
				fmt.Println("err closing writer: ", err.Error())
			}

			client := &http.Client{
				Timeout: time.Second * 600,
			}
			//body := &bytes.Buffer{}
			//writer := multipart.NewWriter(body)
			//fw, err := writer.CreateFormFile("file", data[0]+".jpg")
			//if err != nil {
			//	fmt.Println("err CreateFromFile: ",err.Error())
			//}
			//file, err:= os.Open(data[0]+".jpg")
			//_, err = io.Copy(fw, file)
			//if err != nil {
			//	fmt.Println("err opening: ",err.Error())
			//}

			//req, err := http.NewRequest("POST", urlPhotoSet+"?author="+data[1], bytes.NewReader(body.Bytes()))
			req, err := http.NewRequest("POST", urlPhotoSet, body)
			if err != nil {
				fmt.Println("err creating request: ", err.Error())
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			rsp, err := client.Do(req)
			if err != nil {
				fmt.Println("err making request: ", err.Error())
			}
			if err = writer.Close(); err != nil {
				fmt.Println("err closing writer: ", err.Error())
			}
			//if err =file.Close(); err!= nil {
			//	fmt.Println("err closing file: ",err.Error())
			//}
			if rsp.StatusCode != http.StatusOK {
				log.Printf("Request failed with response code: %d", rsp.StatusCode)
			}
			if err = req.Body.Close(); err != nil {
				fmt.Println("err closing body: ", err.Error())
			}
			if err = arch.MyDelete(data[0] + ".jpg"); err != nil {
				fmt.Println("err deleting file: ", err.Error())
			}
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

		println(users[exist].Command)

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
		users[exist].Command = ""

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

func Upload(values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	resp, err := http.Post("/sendphoto", "multipart/form-data", &b)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
	}
	return
}
