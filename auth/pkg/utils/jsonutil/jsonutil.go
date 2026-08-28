package jsonutil

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3/log"
)

func Stringify(data any, indent ...bool) string {
	var b []byte
	var err error
	if len(indent) > 0 && indent[0] {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}
	if err != nil {
		log.Warn("Marshal data to json error: ", err)
		return ""
	}
	return string(b)
}

func Parse(str string, t any) error {
	return json.Unmarshal([]byte(str), t)
}
