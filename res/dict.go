package res

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Dictionary struct {
	bundle *i18n.Bundle
}

func NewDictionary() *Dictionary {
	d := &Dictionary{}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFile("cn.toml")

	i18n.NewLocalizer(bundle, "en")

	d.bundle = bundle

	return d
}

func (d *Dictionary) Load() {

}

func (d *Dictionary) Switch() {

}
