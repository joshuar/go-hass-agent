// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package translations

import (
	"context"
	"log/slog"

	"github.com/jeandeaual/go-locale"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/joshuar/go-hass-agent/internal/logging"
)

//go:generate gotext -srclang=en update -out=catalog.go -lang=en,fr,de github.com/joshuar/go-hass-agent

// Translator provides a msgPrinter that can display localized strings for
// translation of the UI.
type Translator struct {
	msgPrinter *message.Printer
}

// NewTranslator creates a new Translator in the locale of the system. Strings
// translator by this Translator instance will be localized if a translation is
// available.
func NewTranslator(ctx context.Context) *Translator {
	var msgPrinter *message.Printer

	userLocales, err := locale.GetLocales()
	if err != nil {
		logging.FromContext(ctx).Warn("Could not find any installed locales. Using English.")

		msgPrinter = message.NewPrinter(message.MatchLanguage(language.English.String()))
	} else {
		logging.FromContext(ctx).Debug("Setting language.", slog.Any("lang", userLocales))

		msgPrinter = message.NewPrinter(message.MatchLanguage(userLocales...))
	}

	return &Translator{
		msgPrinter: msgPrinter,
	}
}

// Translate will take a string defined in English and apply the appropriate
// translation (if available) of the defined Translator.
func (t *Translator) Translate(key string, args ...any) string {
	return t.msgPrinter.Sprintf(key, args...)
}
