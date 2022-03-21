package res

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Dictionary provides key-value based mappings,
// and support i18n.
type Dictionary struct {
	bundle *i18n.Bundle
}

// NewDictionary creates a dictionary.
func NewDictionary() *Dictionary {
	d := &Dictionary{}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFile("cn.toml")

	i18n.NewLocalizer(bundle, "en")

	d.bundle = bundle

	return d
}

// Load loads kv mappings from file or db.
func (d *Dictionary) Load() {

}
