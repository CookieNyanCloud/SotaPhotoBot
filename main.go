package main

import (
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	//"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	token  = "1909022612:AAFKXpeKSig9I5nPI7BjZUv8oC4eN56C_9o"
	stiker = "stiker.png"
	stikercent = "cetn.png"
	stikerbok = "chert.png"
	watermarked ="watermarked.jpeg"
)
func main() {

	//bot
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		err.Error()
		return
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if (update.Message.Text) != "" {
			continue
		}


		leng:=len(*update.Message.Photo)
		phUrl, err:= bot.GetFileDirectURL((*update.Message.Photo)[leng-1].FileID)
		filename:="new.jpg"
		err = DownloadFile(phUrl,filename)
		if err != nil {
			err.Error()
		}



		imgb, _ := os.Open(filename)
		img, _ := jpeg.Decode(imgb)
		defer imgb.Close()


		var varstick string
		var offsetX, offsetY int
		pos:=update.Message.Caption
		if pos == "1" {
			varstick = stikerbok
		} else if pos == "2" {
			varstick = stikercent
		} else {
			varstick = stiker
		}

		widthF, heightF := getImageDimension(filename)
		widthS, heightS := getImageDimension(varstick)

		switch varstick {
		case stikerbok:
			offsetX = 0
			offsetY = heightF-heightS
		case stikercent:
			offsetX = (widthF/2) - (widthS/2)
			offsetY = (heightF/2) - (heightS/2)
		default:
			offsetX = 0
			offsetY = heightF-heightS
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
		msg:= tgbotapi.NewPhotoUpload(update.Message.Chat.ID,"watermarked.jpeg")
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

//background:= "new.j/peg"
//watermark:= "stiker.png"
//addWaterMark(background, watermark)

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

func PlaceImage(outName, bgImg, markImg, markDimensions, locationDimensions string) {

	// Coordinate to super-impose on. e.g. 200x500
	locationX, locationY := ParseCoordinates(locationDimensions, "x")

	src := OpenImage(bgImg)

	// Resize the watermark to fit these dimensions, preserving aspect ratio.
	markFit := ResizeImage(markImg, markDimensions)

	// Place the watermark over the background in the location
	dst := imaging.Paste(src, markFit, image.Pt(locationX, locationY))

	err := imaging.Save(dst, outName)

	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	fmt.Printf("Placed image '%s' on '%s'.\n", markImg, bgImg)
}

func ParseCoordinates(input, delimiter string) (int, int) {

	arr := strings.Split(input, delimiter)

	// convert a string to an int
	x, err := strconv.Atoi(arr[0])

	if err != nil {
		log.Fatalf("failed to parse x coordinate: %v", err)
	}

	y, err := strconv.Atoi(arr[1])

	if err != nil {
		log.Fatalf("failed to parse y coordinate: %v", err)
	}

	return x, y
}

func OpenImage(name string) image.Image {
	src, err := imaging.Open(name)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	return src
}

func ResizeImage (image, dimensions string) image.Image {
	width, height := ParseCoordinates(dimensions, "x")
	src := OpenImage(image)
	return imaging.Fit(src, width, height, imaging.Lanczos)
}

func CalcWaterMarkPosition(bgDimensions, markDimensions image.Point, aspectRatio float64) (int, int) {

	bgX := bgDimensions.X
	bgY := bgDimensions.Y
	markX := markDimensions.X
	markY := markDimensions.Y

	padding := 20 * int(aspectRatio)

	return bgX - markX - padding, bgY - markY - padding
}

func addWaterMark(bgImg, watermark string) {

	outName := fmt.Sprintf("watermark-new-%s", watermark)

	src := OpenImage(bgImg)
	dem:= "200x200"

	markFit := ResizeImage(watermark, dem)

	bgDimensions := src.Bounds().Max
	markDimensions := markFit.Bounds().Max

	bgAspectRatio := math.Round(float64(bgDimensions.X) / float64(bgDimensions.Y))

	xPos, yPos := CalcWaterMarkPosition(bgDimensions, markDimensions, bgAspectRatio)

	PlaceImage(outName, bgImg, watermark, dem, fmt.Sprintf("%dx%d", xPos, yPos))

	fmt.Printf("Added watermark '%s' to image '%s' with dimensions %s.\n", watermark, bgImg, dem)
}


