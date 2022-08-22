package meme

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/crunchy-rocks/draw2d"
	"git.fromouter.space/crunchy-rocks/draw2d/draw2dimg"
	"git.fromouter.space/crunchy-rocks/emoji"
	"git.fromouter.space/crunchy-rocks/freetype"
	"git.fromouter.space/hamcha/tg"
)

var memeFontData draw2d.FontData

type Module struct {
	client *tg.Telegram
	name   string
	emojis emoji.Table
}

func (m *Module) Initialize(options modules.ModuleOptions) error {
	m.client = options.API
	m.name = options.Name
	m.emojis = options.Emojis

	rand.Seed(time.Now().Unix())

	fontfile := utils.RequireEnv("CLESSY_MEME_FONT")
	bytes, err := os.ReadFile(fontfile)
	if err != nil {
		return err
	}

	font, err := freetype.ParseFont(bytes)
	if err != nil {
		return err
	}

	memeFontData = draw2d.FontData{
		Name:   "impact",
		Family: draw2d.FontFamilySans,
		Style:  0,
	}
	draw2d.RegisterFont(memeFontData, font)

	log.Println("[meme] Loaded!")

	return nil
}

func (m *Module) OnUpdate(update tg.APIUpdate) {
	// Not a message? Ignore
	if update.Message == nil {
		return
	}
	message := *update.Message

	// Make replies work
	if message.ReplyTo != nil && message.Text != nil && message.ReplyTo.Photo != nil {
		message.Photo = message.ReplyTo.Photo
		message.Caption = message.Text
	}

	if message.Photo != nil && message.Caption != nil {
		caption := *(message.Caption)
		if strings.HasPrefix(caption, "/meme ") && len(caption) > 6 {
			idx := strings.Index(caption, ";")
			if idx < 0 {
				m.client.SendTextMessage(tg.ClientTextMessageData{
					ChatID:  message.Chat.ChatID,
					Text:    "<b>Formato</b>: /meme TESTO IN ALTO<b>;</b>TESTO IN BASSO",
					ReplyID: &message.MessageID,
				})
				return
			}

			txtup := caption[6:idx]
			txtdw := caption[idx+1:]

			maxsz := 0
			photo := tg.APIPhotoSize{}
			for _, curphoto := range message.Photo {
				if curphoto.Width > maxsz {
					maxsz = curphoto.Width
					photo = curphoto
				}
			}

			byt, err := m.client.GetFile(tg.FileRequestData{
				FileID: photo.FileID,
			})
			if err != nil {
				log.Printf("[memegen] Received error: %s\n", err.Error())
				m.client.SendTextMessage(tg.ClientTextMessageData{
					ChatID:  message.Chat.ChatID,
					Text:    "<b>ERRORE!</b> @hamcha controlla la console!",
					ReplyID: &message.MessageID,
				})
				return
			}

			img, _, err := image.Decode(bytes.NewReader(byt))
			if err != nil {
				log.Printf("[memegen] Image decode error: %s\n", err.Error())
				m.client.SendTextMessage(tg.ClientTextMessageData{
					ChatID:  message.Chat.ChatID,
					Text:    "<b>ERRORE!</b> Non riesco a leggere l'immagine",
					ReplyID: &message.MessageID,
				})
				return
			}

			m.client.SendChatAction(tg.ClientChatActionData{
				ChatID: message.Chat.ChatID,
				Action: tg.ActionUploadingPhoto,
			})

			//TODO Clean up this mess

			// Create target image
			bounds := img.Bounds()
			iwidth := float64(bounds.Size().X)
			iheight := float64(bounds.Size().Y)

			timg := image.NewRGBA(bounds)
			gc := draw2dimg.NewGraphicContext(timg)
			gc.Emojis = m.emojis
			gc.SetStrokeColor(image.Black)
			gc.SetFillColor(image.White)
			gc.SetFontData(memeFontData)
			gc.DrawImage(img)

			write := func(text string, istop bool) {
				text = strings.ToUpper(strings.TrimSpace(text))
				gc.Restore()
				gc.Save()

				// Detect appropriate font size
				scale := iheight / iwidth * (iwidth / 10)
				gc.SetFontSize(scale)
				gc.SetLineWidth(scale / 15)

				// Get NEW bounds
				left, top, right, bottom := gc.GetStringBounds(text)

				width := right - left
				texts := []string{text}
				if width > iwidth {
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
					for width > iwidth && iter < 10 {
						log.Println("Warning, resizing!")
						gc.SetFontSize(scale * (0.8 - 0.1*float64(iter)))
						left, top, right, bottom = gc.GetStringBounds(texts[longid])
						width = right - left
						iter++
					}
				}

				height := bottom - top
				margin := float64(height / 50)
				lines := float64(len(texts) - 1)

				gc.Save()
				for id, txt := range texts {
					gc.Save()
					left, _, right, _ = gc.GetStringBounds(txt)
					width = right - left

					y := float64(0)
					if istop {
						y = (height+margin)*float64(id+1) + margin*5
					} else {
						y = iheight - (height * lines) + (height * float64(id)) - margin*5
					}

					gc.Translate((iwidth-width)/2, y)
					gc.StrokeString(txt)
					gc.FillString(txt)
					gc.Restore()
				}
			}
			write(txtup, true)
			write(txtdw, false)

			buf := new(bytes.Buffer)
			err = jpeg.Encode(buf, timg, &(jpeg.Options{Quality: 80}))
			if err != nil {
				log.Printf("[memegen] Image encode error: %s\n", err.Error())
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
				Filename: "meme.jpg",
				ReplyID:  &message.MessageID,
			})
		}
	}
}
