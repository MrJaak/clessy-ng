package snapchat

import (
	"bytes"
	"image"
	"image/color"
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

var snapFontData draw2d.FontData

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

	fontfile := utils.RequireEnv("CLESSY_SNAPCHAT_FONT")
	bytes, err := os.ReadFile(fontfile)
	if err != nil {
		return err
	}

	font, err := freetype.ParseFont(bytes)
	if err != nil {
		return err
	}

	snapFontData = draw2d.FontData{
		Name:   "sourcesans",
		Family: draw2d.FontFamilySans,
		Style:  draw2d.FontStyleBold,
	}
	draw2d.RegisterFont(snapFontData, font)

	log.Println("[snapchat] Loaded!")

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
		if strings.HasPrefix(caption, "/snap ") && len(caption) > 6 {
			txt := strings.TrimSpace(caption[6:])

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
				log.Printf("[snapchat] Received error: %s\n", err.Error())
				m.client.SendTextMessage(tg.ClientTextMessageData{
					ChatID:  message.Chat.ChatID,
					Text:    "<b>ERRORE!</b> @hamcha controlla la console!",
					ReplyID: &message.MessageID,
				})
				return
			}

			img, _, err := image.Decode(bytes.NewReader(byt))
			if err != nil {
				log.Printf("[snapchat] Image decode error: %s\n", err.Error())
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

			// Create target image
			bounds := img.Bounds()
			iwidth := float64(bounds.Size().Y) / 1.6
			iheight := float64(bounds.Size().Y)

			repos := iwidth < float64(bounds.Size().X)
			if !repos {
				iwidth = float64(bounds.Size().X)
			}

			timg := image.NewRGBA(image.Rect(0, 0, int(iwidth), int(iheight)))
			gc := draw2dimg.NewGraphicContext(timg)
			gc.Emojis = m.emojis
			gc.SetFontData(snapFontData)

			gc.Save()
			if repos {
				gc.Translate(-(float64(bounds.Size().X)-iwidth)/2, 0)
			}
			gc.DrawImage(img)
			gc.Restore()

			scale := iwidth / 25
			gc.SetFontSize(scale)

			lineMargin := scale / 3
			boxMargin := lineMargin
			topMargin := lineMargin / 6
			write := func(text string, startHeight float64) {
				texts := utils.WordWrap(gc, strings.TrimSpace(text), iwidth*0.9)
				totalHeight := startHeight
				firstLine := 0.
				for _, txt := range texts {
					_, top, _, bottom := gc.GetStringBounds(txt)
					height := (bottom - top)
					if firstLine == 0 {
						firstLine = height
					}
					totalHeight += lineMargin + height
				}

				// Draw background
				starty := startHeight - boxMargin - topMargin - firstLine
				endy := totalHeight + boxMargin - firstLine
				gc.Save()
				gc.SetFillColor(color.RGBA{0, 0, 0, 160})
				gc.BeginPath()
				gc.MoveTo(0, starty)
				gc.LineTo(iwidth, starty)
				gc.LineTo(iwidth, endy)
				gc.LineTo(0, endy)
				gc.Close()
				gc.Fill()
				gc.Restore()

				// Write lines
				gc.SetFillColor(image.White)
				height := startHeight
				for _, txt := range texts {
					left, top, right, bottom := gc.GetStringBounds(txt)
					width := right - left
					gc.FillStringAt(txt, (iwidth-width)/2, height)
					height += lineMargin + (bottom - top)
				}
			}
			write(txt, (rand.Float64()*0.4+0.3)*iheight)

			buf := new(bytes.Buffer)
			err = jpeg.Encode(buf, timg, &(jpeg.Options{Quality: 80}))
			if err != nil {
				log.Printf("[snapchat] Image encode error: %s\n", err.Error())
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
