package localizer

import (
	"context"
	"net/http"

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

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Header.Get("Accept-Language")
		l := Get(lang)
		ctx := context.WithValue(r.Context(), "localizer", l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
