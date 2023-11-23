package translations

import (
	_ "golang.org/x/text/message"
)

//go:generate gotext -srclang=en update -out=catalog.go -lang=en,ru,uz github.com/yeahuz/yeah-api
