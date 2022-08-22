package proverbio

import (
	_ "embed"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/hamcha/tg"
)

//go:embed proverbi.txt
var proverbidata string

var proverbipairs struct {
	start []string
	end   []string
}

type Module struct {
	client *tg.Telegram
	name   string
}

func (m *Module) Initialize(options modules.ModuleOptions) error {
	m.client = options.API
	m.name = options.Name

	lines := strings.Split(proverbidata, "\n")
	for i, line := range lines {
		pair := strings.SplitN(line, "/", 2)
		if len(pair) < 2 {
			log.Printf("[proverbio] Found line without separator (#%d), skipping\n", i)
			continue
		}
		proverbipairs.start = append(proverbipairs.start, strings.TrimSpace(pair[0]))
		proverbipairs.end = append(proverbipairs.end, strings.TrimSpace(pair[1]))
	}

	log.Printf("[proverbio] Loaded %d pairs (%d combinations!)\n", len(proverbipairs.start), len(proverbipairs.start)*len(proverbipairs.end))
	return nil
}

func (m *Module) OnUpdate(update tg.APIUpdate) {
	// Not a message? Ignore
	if update.Message == nil {
		return
	}

	if utils.IsCommand(*update.Message, m.name, "proverbio") {
		start := rand.Intn(len(proverbipairs.start))
		end := rand.Intn(len(proverbipairs.end))
		m.client.SendTextMessage(tg.ClientTextMessageData{
			ChatID: update.Message.Chat.ChatID,
			Text:   fmt.Sprintf("<b>Dice il saggio:</b>\n%s %s", proverbipairs.start[start], proverbipairs.end[end]),
		})
		return
	}
}
