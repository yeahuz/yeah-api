package yeahapi

import (
	_ "github.com/yeahuz/yeah-api/internal/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Localizer struct {
	fallback string
	printers []printer
}

type printer struct {
	code    string
	printer *message.Printer
}

type LocalizerService interface {
	Get(args ...string) *printer
}

func NewLocalizerService(fallback string) *Localizer {
	return &Localizer{
		printers: []printer{
			{code: "en", printer: message.NewPrinter(language.MustParse("en"))},
			{code: "ru", printer: message.NewPrinter(language.MustParse("ru"))},
			{code: "uz", printer: message.NewPrinter(language.MustParse("uz"))},
		},
		fallback: fallback,
	}
}

func (l *Localizer) Get(args ...string) (p printer) {
	locale := l.fallback
	if len(args) > 0 {
		locale = args[0]
	}

	for _, printer := range l.printers {
		if printer.code == locale {
			p = printer
		}
	}

	return p
}

func (p *printer) T(key message.Reference, args ...interface{}) string {
	return p.printer.Sprintf(key, args...)
}
