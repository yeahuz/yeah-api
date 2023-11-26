package common

import (
	"net/http"
)

type ApiFunc func(w http.ResponseWriter, r *http.Request) error
