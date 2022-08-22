package utils

import (
	"github.com/cockroachdb/pebble"
	jsoniter "github.com/json-iterator/go"
)

func ReadJSON(db *pebble.DB, key string, dst interface{}) error {
	data, closer, err := db.Get([]byte(key))
	if err != nil {
		return err
	}
	defer closer.Close()
	return jsoniter.ConfigFastest.Unmarshal(data, dst)
}
