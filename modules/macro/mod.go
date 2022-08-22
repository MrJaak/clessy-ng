package macro

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/hamcha/tg"
	"github.com/cockroachdb/pebble"
	jsoniter "github.com/json-iterator/go"
)

type Macro struct {
	Value  string
	Author tg.APIUser
	Time   time.Time
}

var macros map[string]Macro

const macroKey = "mod/macros/data"

type Module struct {
	client *tg.Telegram
	name   string
	kv     *pebble.DB
}

func (m *Module) Initialize(options modules.ModuleOptions) error {
	m.client = options.API
	m.name = options.Name
	m.kv = options.KV

	macros = make(map[string]Macro)
	err := utils.ReadJSON(m.kv, macroKey, &macros)
	if err != nil {
		if !errors.Is(err, pebble.ErrNotFound) {
			log.Println("[macro] WARN: Could not load macros (db error): " + err.Error())
			return err
		}
	}
	log.Printf("[macro] Loaded %d macros\n", len(macros))

	return nil
}

func (m *Module) OnUpdate(update tg.APIUpdate) {
	// Not a message? Ignore
	if update.Message == nil {
		return
	}

	if utils.IsCommand(*update.Message, m.name, "macro") {
		parts := strings.SplitN(*(update.Message.Text), " ", 3)
		switch len(parts) {
		case 2:
			name := strings.ToLower(parts[1])
			item, ok := macros[name]
			var out string
			if ok {
				out = fmt.Sprintf("<b>%s</b>\n%s\n<i>%s - %s</i>", name, item.Value, item.Author.Username, item.Time.Format("02-01-06 15:04"))
			} else {
				out = fmt.Sprintf("<b>%s</b>\n<i>macro inesistente</i>\n(configura con /macro %s <i>contenuto</i>)", name, name)
			}
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  update.Message.Chat.ChatID,
				Text:    out,
				ReplyID: &update.Message.MessageID,
			})
		case 3:
			name := strings.ToLower(parts[1])
			macros[name] = Macro{
				Value:  parts[2],
				Author: update.Message.User,
				Time:   time.Now(),
			}
			m.save()
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  update.Message.Chat.ChatID,
				Text:    fmt.Sprintf("<b>%s</b> → %s", name, parts[2]),
				ReplyID: &update.Message.MessageID,
			})
		default:
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  update.Message.Chat.ChatID,
				Text:    "<b>Sintassi</b>\n<b>Leggi</b>: /macro <i>nome-macro</i>\n<b>Scrivi</b>: /macro <i>nome-macro</i> <i>contenuto macro</i>",
				ReplyID: &update.Message.MessageID,
			})
		}
		return
	}
}

func (m *Module) save() {
	byt, err := jsoniter.ConfigFastest.Marshal(macros)
	if err != nil {
		log.Println("[macro] WARN: Could not encode macros: " + err.Error())
	}
	err = m.kv.Set([]byte(macroKey), byt, &pebble.WriteOptions{Sync: true})
	if err != nil {
		log.Println("[macro] WARN: Could not save macros to db: " + err.Error())
		return
	}
}
