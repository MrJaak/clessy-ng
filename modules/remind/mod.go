package remind

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/hamcha/tg"
	"git.sr.ht/~hamcha/containers"
	"github.com/cockroachdb/pebble"
	jsoniter "github.com/json-iterator/go"
)

type Reminder struct {
	TargetID  int64
	When      int64
	Text      string
	Reference *ReminderReference
}

type ReminderReference struct {
	Chat    int64
	Message int64
}

const reminderKey = "mod/reminder/pending"
const reminderMaxDuration = time.Hour * 24 * 30 * 3

var defaultLocation *time.Location = nil

var pending = containers.NewRWSyncMap[string, Reminder]()

type Module struct {
	client *tg.Telegram
	name   string
	kv     *pebble.DB
}

func (m *Module) Initialize(options modules.ModuleOptions) error {
	m.client = options.API
	m.name = options.Name
	m.kv = options.KV

	var err error
	defaultLocation, err = time.LoadLocation("Europe/Rome")
	if err != nil {
		log.Fatalf("[remind] Something is really wrong: %s\n", err.Error())
		return err
	}

	var reminders map[string]Reminder
	err = utils.ReadJSON(m.kv, reminderKey, &reminders)
	if err != nil {
		if !errors.Is(err, pebble.ErrNotFound) {
			log.Println("[remind] WARN: Could not load pending reminders (db error): " + err.Error())
			return err
		}
	}
	for id, reminder := range reminders {
		pending.SetKey(id, reminder)
		go m.schedule(id)
	}
	log.Printf("[remind] Loaded %d pending reminders\n", len(reminders))

	return nil
}

func (m *Module) OnUpdate(update tg.APIUpdate) {
	// Not a message? Ignore
	if update.Message == nil {
		return
	}
	message := *update.Message

	if utils.IsCommand(message, m.name, "ricordami") {
		// Supported formats:
		// Xs/m/h/d   => in X seconds/minutes/hours/days
		// HH:MM      => at HH:MM    (24 hour format)
		// HH:MM:SS   => at HH:MM:SS (24 hour format)
		// dd/mm/yyyy => same hour, specific date
		// dd/mm/yyyy-HH:MM    => specific hour, specific dat
		// dd/mm/yyyy-HH:MM:SS => specific hour, specific date

		parts := strings.SplitN(*message.Text, " ", 3)

		// Little hack to allow text-less reminders with replies
		if len(parts) == 2 && message.ReplyTo != nil {
			parts = append(parts, "")
		}

		if len(parts) < 3 {
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "<b>Sintassi</b>\n/ricordami <i>[quando]</i> Messaggio\n\n<b>Formati supportati per [quando]</b>:\n 10s 10m 10h 10d (secondi/minuti/ore/giorni)\n 13:20 15:55:01 (ora dello stesso giorno, formato 24h)\n 11/02/2099 11/02/2099-11:20:01 (giorno diverso, stessa ora [1] o specifica [2])",
				ReplyID: &message.MessageID,
			})
			return
		}

		format := parts[1]
		remindText := parts[2]

		loc := defaultLocation
		/*TODO REDO
		if uloc, ok := tzlocs[update.User.UserID]; ok {
			loc = uloc
		}
		*/
		timestamp, err := parseDuration(format, loc)
		if err != nil {
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    err.Error(),
				ReplyID: &message.MessageID,
			})
			return
		}

		id := strconv.FormatInt(message.Chat.ChatID, 36) + "-" + strconv.FormatInt(message.MessageID, 36)
		reminder := Reminder{
			TargetID: message.User.UserID,
			When:     timestamp.Unix(),
			Text:     remindText,
		}
		if message.ReplyTo != nil {
			reminder.Reference = &ReminderReference{
				Chat:    message.Chat.ChatID,
				Message: message.ReplyTo.MessageID,
			}
		}
		pending.SetKey(id, reminder)
		m.save()
		go m.schedule(id)

		whenday := "più tardi"
		_, todaym, todayd := time.Now().Date()
		_, targetm, targetd := timestamp.Date()
		if todaym != targetm || todayd != targetd {
			whenday = "il " + timestamp.In(loc).Format("2/1")
		}
		whentime := "alle " + timestamp.In(loc).Format("15:04:05")
		m.client.SendTextMessage(tg.ClientTextMessageData{
			ChatID:  message.Chat.ChatID,
			Text:    "Ok, vedrò di avvisarti " + whenday + " " + whentime,
			ReplyID: &message.MessageID,
		})
		return
	}

	if utils.IsCommand(message, m.name, "reminders") {
		// Should only work in private chats
		if message.Chat.Type != tg.ChatTypePrivate {

			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "Per favore chiedimi in privato dei reminder",
				ReplyID: &message.MessageID,
			})
			return
		}

		useritems := []Reminder{}
		for _, reminder := range pending.Copy() {
			if reminder.TargetID == message.User.UserID {
				useritems = append(useritems, reminder)
			}
		}

		if len(useritems) == 0 {
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    "Non ci sono reminder in coda per te",
				ReplyID: &message.MessageID,
			})
		} else {
			m.client.SendTextMessage(tg.ClientTextMessageData{
				ChatID:  message.Chat.ChatID,
				Text:    fmt.Sprintf("Ci sono <b>%d</b> reminder in coda per te", len(useritems)),
				ReplyID: &message.MessageID,
			})
		}
	}
}

func (m *Module) schedule(id string) {
	// Get reminder
	r := pending.GetKey(id)
	remaining := r.When - time.Now().Unix()
	if remaining > 0 {
		// Wait remaining time
		time.Sleep(time.Second * time.Duration(remaining))
	}
	// Remind!
	m.client.SendTextMessage(tg.ClientTextMessageData{
		ChatID:  r.TargetID,
		Text:    "<b>Heyla! Mi avevi chiesto di ricordarti questo:</b>\n" + r.Text,
		ReplyID: nil,
	})
	if r.Reference != nil {
		m.client.ForwardMessage(tg.ClientForwardMessageData{
			ChatID:     r.TargetID,
			FromChatID: r.Reference.Chat,
			MessageID:  r.Reference.Message,
		})
	}
	// Delete reminder from pending list and save list to disk
	pending.DeleteKey(id)
	m.save()
}

func (m *Module) save() {
	byt, err := jsoniter.ConfigFastest.Marshal(pending.Copy())
	if err != nil {
		log.Println("[remind] WARN: Could not encode reminders: " + err.Error())
	}
	err = m.kv.Set([]byte(reminderKey), byt, &pebble.WriteOptions{Sync: true})
	if err != nil {
		log.Println("[remind] WARN: Could not save reminders to db: " + err.Error())
		return
	}
}

func isSscanfValid(n int, err error) bool {
	return err == nil
}

func scanMixedDelay(str string) (bool, time.Time, error) {
	remaining := str
	now := time.Now()
	num := 0
	sep := ' '
	for len(remaining) > 1 {
		_, err := fmt.Sscanf(remaining, "%d%c", &num, &sep)
		if err != nil {
			return false, now, err
		}
		dur := time.Duration(num)
		switch unicode.ToLower(sep) {
		case 's':
			dur *= time.Second
		case 'm':
			dur *= time.Minute
		case 'h':
			dur *= time.Hour
		case 'd':
			dur *= time.Hour * 24
		default:
			return true, now, fmt.Errorf("La durata ha una unità che non conosco, usa una di queste: s (secondi) m (minuti) h (ore) d (giorni)")
		}
		now = now.Add(dur)
		nextIndex := strings.IndexRune(remaining, sep)
		remaining = remaining[nextIndex+1:]
	}
	fmt.Printf("tot: %s", now.Sub(time.Now()))
	return true, now, nil
}

func parseDuration(date string, loc *time.Location) (time.Time, error) {
	now := time.Now().In(loc)
	hour := now.Hour()
	min := now.Minute()
	sec := now.Second()
	day := now.Day()
	month := now.Month()
	year := now.Year()
	dayunspecified := false
	isDurationFmt, duration, err := scanMixedDelay(date)
	switch {
	case isSscanfValid(fmt.Sscanf(date, "%d/%d/%d-%d:%d:%d", &day, &month, &year, &hour, &min, &sec)):
	case isSscanfValid(fmt.Sscanf(date, "%d/%d/%d-%d:%d", &day, &month, &year, &hour, &min)):
		sec = 0
	case isSscanfValid(fmt.Sscanf(date, "%d/%d/%d", &day, &month, &year)):
		hour = now.Hour()
		min = now.Minute()
		sec = now.Second()
	case isSscanfValid(fmt.Sscanf(date, "%d:%d:%d", &hour, &min, &sec)):
		day = now.Day()
		month = now.Month()
		year = now.Year()
		dayunspecified = true
	case isSscanfValid(fmt.Sscanf(date, "%d:%d", &hour, &min)):
		day = now.Day()
		month = now.Month()
		year = now.Year()
		sec = 0
		dayunspecified = true
	case isDurationFmt:
		return duration, err
	default:
		return now, fmt.Errorf("Non capisco quando dovrei ricordartelo!")
	}
	targetDate := time.Date(year, month, day, hour, min, sec, 0, loc)
	if targetDate.Before(now) {
		// If day was not specified assume tomorrow
		if dayunspecified {
			targetDate = targetDate.Add(time.Hour * 24)
		} else {
			return now, fmt.Errorf("Non posso ricordarti cose nel passato!")
		}
	}
	if targetDate.After(now.Add(reminderMaxDuration)) {
		return now, fmt.Errorf("Non credo riuscirei a ricordarmi qualcosa per così tanto")
	}
	return targetDate, nil
}
