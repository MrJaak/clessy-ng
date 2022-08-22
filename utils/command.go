package utils

import (
	"strings"

	"git.fromouter.space/hamcha/tg"
)

func IsCommand(update tg.APIMessage, botname string, cmdname string) bool {
	if update.Text == nil {
		return false
	}

	text := strings.TrimSpace(*(update.Text))

	shortcmd := "/" + cmdname
	fullcmd := shortcmd + "@" + botname

	// Check short form
	if text == shortcmd || strings.HasPrefix(text, shortcmd+" ") {
		return true
	}

	// Check long form
	if text == fullcmd || strings.HasPrefix(text, fullcmd+" ") {
		return true
	}

	return false
}
