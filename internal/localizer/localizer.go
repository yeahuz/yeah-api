package localizer

import (
	_ "github.com/yeahuz/yeah-api/internal/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Localizer struct {
	ID      string
	printer *message.Printer
}

var locales = []Localizer{
	{
		ID:      "en",
		printer: message.NewPrinter(language.MustParse("en")),
	},
	{
		ID:      "ru",
		printer: message.NewPrinter(language.MustParse("ru")),
	},
	{
		ID:      "uz",
		printer: message.NewPrinter(language.MustParse("uz")),
	},
}

func GetDefault() Localizer {
	return locales[0]
}

func Get(id string) Localizer {
	for _, locale := range locales {
		if id == locale.ID {
			return locale
		}
	}

	return locales[0]
}

func (l Localizer) T(key message.Reference, args ...interface{}) string {
	return l.printer.Sprintf(key, args...)
}
