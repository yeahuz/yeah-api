package frontend

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

func generateQRDataURL(data string) (string, error) {
	q, err := qrcode.New(data, qrcode.Low)
	if err != nil {
		return "", err
	}

	q.DisableBorder = true
	png, err := q.PNG(256)
	if err != nil {
		return "", err
	}

	b64 := base64.RawStdEncoding.EncodeToString(png)
	return "data:image/png;base64," + b64, nil
}

func fallbackStr(str, fallback string) string {
	if str == "" {
		return fallback
	}
	return str
}
