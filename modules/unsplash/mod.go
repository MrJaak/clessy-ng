package unsplash

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/crunchy-rocks/draw2d"
	"git.fromouter.space/crunchy-rocks/draw2d/draw2dimg"
	"git.fromouter.space/crunchy-rocks/emoji"
	"git.fromouter.space/crunchy-rocks/freetype"
	"git.fromouter.space/hamcha/tg"
	"github.com/disintegration/imaging"
)

var quoteFontData draw2d.FontData
var pics []string

type Module struct {
	client *tg.Telegram
	name   string
	emojis emoji.Table
}

func (m *Module) Initialize(options modules.ModuleOptions) error {
	m.client = options.API
	m.name = options.Name
	m.emojis = options.Emojis

	fontfile := utils.RequireEnv("CLESSY_UNSPLASH_FONT")
	bytes, err := os.ReadFile(fontfile)
	if err != nil {
		return err
	}

	font, err := freetype.ParseFont(bytes)
	if err != nil {
		return err
	}

	quoteFontData = draw2d.FontData{
		Name:   "gillmt",
		Family: draw2d.FontFamilySans,
		Style:  draw2d.FontStyleBold,
	}
	draw2d.RegisterFont(quoteFontData, font)

	// Read all the pictures inside a folder and save them for later
	bgpath := utils.RequireEnv("CLESSY_UNSPLASH_BG_PATH")
	files, err := ioutil.ReadDir(bgpath)
	if err != nil {
		return err
	}

	for _, file := range files {
		pics = append(pics, filepath.Join(bgpath, file.Name()))
	}

	log.Printf("[unsplash] Loaded (%d available backgrounds)", len(pics))

	return nil
}

func (m *Module) OnUpdate(update tg.APIUpdate) {
	// Not a message? Ignore
	if update.Message == nil {
		return
	}
	message := *update.Message

	if utils.IsCommand(message, m.name, "unsplash") {
		text := ""
		user := message.User

		if message.ReplyTo != nil {
			switch {
			case message.ReplyTo.Text != nil:
				text = *(message.ReplyTo.Text)
			case message.ReplyTo.Caption != nil:
				text = *(message.ReplyTo.Caption)
			default:
				m.client.SendTextMessage(tg.ClientTextMessageData{
					ChatID:  message.Chat.ChatID,
					Text:    "Non c'e' niente di 'ispiratore' in questo..",
					ReplyID: &message.MessageID,
				})
				return
			}

			// For forwarded message take the original user
			if message.FwdUser != nil {
				user = *message.FwdUser
			} else {
				user = message.ReplyTo.User
			}
		} else {
			if strings.Index(*(message.Text), " ") > 0 {
				text = strings.TrimSpace(strings.SplitN(*(message.Text), " ", 2)[1])
			}
		}

		// Cleanup chars
		text = strings.Map(stripUnreadable, text)

		author := user.FirstName
		if user.LastName != "" {
			author = user.FirstName + " " + user.LastName
		}
		author += " (" + user.Username + ")"

		if strings.TrimSpace(text) == "" {
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "Non c'e' niente di 'ispiratore' in questo..",
				ReplyID: &message.MessageID,
			})
			return
		}

		file, err := os.Open(pics[rand.Intn(len(pics))])
		if err != nil {
			log.Printf("[unsplash] Could not open original image file: %s\n", err.Error())
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "<b>ERRORE!</b> @hamcha controlla la console!",
				ReplyID: &message.MessageID,
			})
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			log.Printf("[unsplash] Image decode error: %s\n", err.Error())
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "<b>ERRORE!</b> @hamcha controlla la console!",
				ReplyID: &message.MessageID,
			})
			return
		}

		m.client.SendChatAction(tg.ClientChatActionData{
			ChatID: message.Chat.ChatID,
			Action: tg.ActionUploadingPhoto,
		})

		// Darken image
		img = imaging.AdjustBrightness(imaging.AdjustGamma(imaging.AdjustSigmoid(img, 0.5, -6.0), 0.8), -20)

		// Create target image
		bounds := img.Bounds()
		iwidth := float64(bounds.Size().X)
		iheight := float64(bounds.Size().Y)

		timg := image.NewRGBA(bounds)
		gc := draw2dimg.NewGraphicContext(timg)
		gc.Emojis = m.emojis
		gc.SetFontData(quoteFontData)
		gc.DrawImage(img)
		gc.SetStrokeColor(image.Black)
		gc.SetFillColor(image.White)

		text = strings.ToUpper(strings.TrimSpace(text))
		gc.Restore()
		gc.Save()

		// Detect appropriate font size
		scale := iheight / iwidth * (iwidth / 10) * 0.8
		gc.SetFontSize(scale)
		gc.SetLineWidth(scale / 15)

		// Get NEW bounds
		left, top, right, bottom := gc.GetStringBounds(text)

		width := right - left
		texts := []string{text}
		if width*1.2 > iwidth {
			// Split text
			texts = utils.SplitCenter(text)

			// Get longest line
			longer := float64(0)
			longid := 0
			widths := make([]float64, len(texts))
			for id := range texts {
				tleft, _, tright, _ := gc.GetStringBounds(texts[id])
				widths[id] = tright - tleft
				if width > longer {
					longer = widths[id]
					longid = id
				}
			}

			// Still too big? Decrease font size again
			iter := 0
			for width*1.2 > iwidth && iter < 10 {
				scale *= (0.9 - 0.05*float64(iter))
				gc.SetFontSize(scale)
				left, top, right, bottom = gc.GetStringBounds(texts[longid])
				width = right - left
				iter++
			}
		}

		texts = append(texts, author)
		height := bottom - top + 20
		margin := float64(height / 50)
		txtheight := (height + margin) * float64(len(texts))

		gc.Save()
		for id, txt := range texts {
			gc.Save()
			left, _, right, _ = gc.GetStringBounds(txt)
			width = right - left

			x := (iwidth - width) / 2
			y := (iheight-txtheight)/2 + (height+margin*2)*float64(id+1)
			if id == len(texts)-1 {
				gc.SetFontSize(scale * 0.7)
				left, _, right, _ = gc.GetStringBounds(txt)
				width = right - left
				x = (iwidth - width) / 1.5
				y = (iheight-txtheight)/2 + (height+margin)*float64(id+1) + margin*6
			}

			gc.Translate(x, y)
			gc.StrokeString(txt)
			gc.FillString(txt)
			gc.Restore()
		}

		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, timg, &(jpeg.Options{Quality: 80}))
		if err != nil {
			log.Printf("[unsplash] Image encode error: %s\n", err.Error())
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "<b>ERRORE!</b> @hamcha controlla la console!",
				ReplyID: &message.MessageID,
			})
			return
		}
		m.client.SendPhoto(tg.ClientPhotoData{
			ChatID:   message.Chat.ChatID,
			Bytes:    buf.Bytes(),
			Filename: "quote.jpg",
			ReplyID:  &message.MessageID,
		})
	}
}

func stripUnreadable(r rune) rune {
	if r == '\n' || r == '\t' {
		return ' '
	}
	if r < 32 {
		return -1
	}
	return r
}
