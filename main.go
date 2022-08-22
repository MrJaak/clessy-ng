package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	"git.fromouter.space/crunchy-rocks/clessy-ng/modules"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/macro"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/meme"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/metafora"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/proverbio"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/remind"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/snapchat"
	"git.fromouter.space/crunchy-rocks/clessy-ng/modules/unsplash"
	"git.fromouter.space/crunchy-rocks/clessy-ng/utils"

	"git.fromouter.space/crunchy-rocks/emoji"
	"git.fromouter.space/hamcha/tg"
	"github.com/cockroachdb/pebble"
)

var mods = map[string]modules.Module{
	"metafora":  &metafora.Module{},
	"proverbio": &proverbio.Module{},
	"macro":     &macro.Module{},
	"remind":    &remind.Module{},
	"unsplash":  &unsplash.Module{},
	"snapchat":  &snapchat.Module{},
	"meme":      &meme.Module{},
}

func checkErr(err error, message string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		log.Fatalf("FATAL: "+message+":\n\t%s", args...)
	}
}

func main() {
	disable := flag.String("disable", "", "Blacklist mods (separated by comma)")
	enable := flag.String("enable", "", "Whitelist mods (separated by comma)")
	macrofile := flag.String("import-macros", "", "If specified, path to JSON file containing macros to import to DB")
	remindfile := flag.String("import-reminders", "", "If specified, path to JSON file containing reminders to import to DB")
	flag.Parse()

	// Make Telegram API client
	api := tg.MakeAPIClient(utils.RequireEnv("CLESSY_TOKEN"))
	name, err := api.GetMe()
	checkErr(err, "could not retrieve bot info")

	// Load emojis
	emojis, err := emoji.ScanEmojiDirectory(utils.RequireEnv("CLESSY_EMOJI_PATH"))
	if err != nil {
		log.Printf("[x-emoji] Error while loading emojis: %s\n", err.Error())
		log.Println("[x-emoji] Emoji support will be disabled")
	} else {
		log.Printf("[x-emoji] Loaded %d emojis\n", len(emojis))
	}

	// Initialize your database
	db, err := pebble.Open(utils.RequireEnv("CLESSY_DB_DIR"), &pebble.Options{})
	checkErr(err, "could not open database")
	defer db.Close()

	// Perform imports
	if *macrofile != "" {
		byt, err := os.ReadFile(*macrofile)
		checkErr(err, "could not load macro file")

		db.Set([]byte("mod/macros/data"), byt, &pebble.WriteOptions{Sync: true})
		log.Println("Imported macros")
	}
	if *remindfile != "" {
		byt, err := os.ReadFile(*macrofile)
		checkErr(err, "could not load macro file")

		db.Set([]byte("mod/macros/data"), byt, &pebble.WriteOptions{Sync: true})
		log.Println("Imported reminders")
	}

	// Pick modules
	toActivate := make(map[string]modules.Module)
	if *disable != "" {
		for _, modname := range strings.Split(*disable, ",") {
			modname = strings.TrimSpace(modname)
			delete(mods, modname)
		}
		toActivate = mods
	} else if *enable != "" {
		for _, modname := range strings.Split(*enable, ",") {
			toActivate[modname] = mods[modname]
		}
	} else {
		toActivate = mods
	}

	// Initialize modules
	for modname, mod := range toActivate {
		log.Printf("Initializing %s", modname)
		err := mod.Initialize(modules.ModuleOptions{
			API:    api,
			Name:   name.Username,
			KV:     db,
			Emojis: emojis,
		})
		checkErr(err, "Starting module %s failed with error", modname)
	}

	// Set webhook and handle calls
	webhook := utils.RequireEnv("CLESSY_WEBHOOK")
	uri, err := url.Parse(webhook)
	checkErr(err, "Specified webhook is not a valid URL")

	api.SetWebhook(webhook)
	log.Fatal(api.HandleWebhook(utils.EnvFallback("CLESSY_BIND", ":8080"), uri.Path, func(update tg.APIUpdate) {
		for _, mod := range toActivate {
			mod.OnUpdate(update)
		}
	}))
}
