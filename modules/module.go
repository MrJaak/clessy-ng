package modules

import (
	"git.fromouter.space/crunchy-rocks/emoji"
	"git.fromouter.space/hamcha/tg"
	"github.com/cockroachdb/pebble"
)

type ModuleOptions struct {
	API    *tg.Telegram
	Name   string
	KV     *pebble.DB
	Emojis emoji.Table
}

type Module interface {
	Initialize(options ModuleOptions) error
	OnUpdate(tg.APIUpdate)
}
