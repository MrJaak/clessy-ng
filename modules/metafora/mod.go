package metafora

import (
	"math/rand"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/hamcha/tg"
)

var metaactions = []string{
	"Puppami", "Degustami", "Lucidami", "Manipolami", "Disidratami", "Irritami", "Martorizzami",
	"Lustrami", "Osannami", "Sorseggiami", "Assaporami", "Apostrofami", "Spremimi", "Dimenami",
	"Agitami", "Stimolami", "Suonami", "Strimpellami", "Stuzzicami", "Spintonami", "Sguinzagliami",
	"Modellami", "Sgrullami", "Cavalcami", "Perquotimi", "Misurami", "Sventolami", "Induriscimi",
	"Accordami", "Debuggami", "Accarezzami", "Revisionami", "Imbottigliami", "Badami", "Scuotimi",
	"Terremotami", "Incentivami", "Sollecitami", "Allenami", "Censiscimi", "Decollami", "Smagnetizzami",
	"Nobilitami", "Elevami", "Accrescimi", "Impostami", "Ereggimi", "Fischiettami", "Scaldami", "Gonfiami",
	"Lubrificami",
}

var metaobjects = []string{
	"il birillo", "il bastone", "l'ombrello", "il malloppo", "il manico", "il manganello",
	"il ferro", "la mazza", "l'archibugio", "il timone", "l'arpione", "il flauto", "la reliquia",
	"il fioretto", "lo scettro", "il campanile", "la proboscide", "il pino", "il maritozzo", "il perno",
	"il tubo da 100", "la verga", "l'idrante", "il pendolo", "la torre di Pisa", "la lancia",
	"il cilindro", "il lampione", "il joystick", "il Wiimote", "il PSMove", "l'albero maestro",
	"il trenino", "la sciabola", "il weedle", "il serpente", "il missile", "la limousine",
	"il selfie-stick", "il candelotto", "la falce", "la biscia", "la banana", "la pannocchia",
	"il papavero", "la carota", "la fava", "la salsiccia", "il cono", "l'hard drive", "la manopola",
	"la manovella", "il pennello", "l'asta", "il cacciavite", "lo spazzolino",
}

type Module struct {
	client *tg.Telegram
	name   string
}

func (m *Module) Initialize(options modules.ModuleOptions) error {
	m.client = options.API
	m.name = options.Name
	return nil
}

func (m *Module) OnUpdate(update tg.APIUpdate) {
	// Not a message? Ignore
	if update.Message == nil {
		return
	}

	if utils.IsCommand(*update.Message, m.name, "metafora") {
		m.client.SendTextMessage(tg.ClientTextMessageData{
			ChatID: update.Message.Chat.ChatID,
			Text:   metaforaAPI(),
		})
		return
	}
}

func metaforaAPI() string {
	n := rand.Intn(len(metaactions))
	m := rand.Intn(len(metaobjects))
	return metaactions[n] + " " + metaobjects[m]
}
