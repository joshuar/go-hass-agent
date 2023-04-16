// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package translations

import (
	"github.com/jeandeaual/go-locale"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:generate gotext -srclang=en update -out=catalog.go -lang=en github.com/joshuar/go-hass-agent

// Translator provides a msgPrinter that can display localised strings for
// translation of the UI
type Translator struct {
	msgPrinter *message.Printer
}

// NewTranslator creates a new Translator in the locale of the system. Strings
// translater by this Translator instance will be localised if a translation is
// available.
func NewTranslator() *Translator {
	t := &Translator{}
	userLocales, err := locale.GetLocales()
	if err != nil {
		log.Warn().Msg("Could not find a suitable locale. Defaulting to English.")
		t.msgPrinter = message.NewPrinter(message.MatchLanguage(language.English.String()))
	}
	log.Debug().Caller().Msgf("Setting language to %v.", userLocales)
	t.msgPrinter = message.NewPrinter(message.MatchLanguage(userLocales...))
	return t
}

// Translate will take a string defined in English and apply the appropriate
// translation (if available) of the defined Translator.
func (t *Translator) Translate(key string, args ...interface{}) string {
	return t.msgPrinter.Sprintf(key, args...)
}
