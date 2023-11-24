package auth

import (
	"regexp"

	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var (
	emailRegex = regexp.MustCompile(`(?i)^(([^<>()[\].,;:\s@"]+(\.[^<>()[\].,;:\s@"]+)*)|(".+"))@(([^<>()[\].,;:\s@"]+\.)+[^<>()[\].,;:\s@"]{2,})$`)
)

func (pcd PhoneCodeData) validate(l *localizer.Localizer) error {
	if len(pcd.PhoneNumber) == 0 {
		return c.ErrBadRequest(l.T("Phone number is required"))
	}
	if len(pcd.PhoneNumber) != 13 {
		return c.ErrBadRequest(l.T("Phone number is invalid"))
	}

	return nil
}

func (ecd EmailCodeData) validate(l *localizer.Localizer) error {
	if len(ecd.Email) == 0 {
		return c.ErrBadRequest(l.T("Email is required"))
	}

	if !emailRegex.MatchString(ecd.Email) {
		return c.ErrBadRequest(l.T("Email is invalid"))
	}

	return nil
}
