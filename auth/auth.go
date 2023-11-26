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
	if len(pcd.PhoneNumber) > 13 {
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

func (sipd SignInPhoneData) validate(l *localizer.Localizer) error {
	errors := make(map[string]string)

	if len(sipd.PhoneNumber) > 13 {
		errors["phone_number"] = l.T("Phone number is invalid")
	}

	if len(sipd.PhoneNumber) == 0 {
		errors["phone_number"] = l.T("Phone number is required")
	}

	if len(sipd.Code) == 0 {
		errors["code"] = l.T("Phone code is required")
	}

	if len(sipd.Hash) == 0 {
		errors["hash"] = l.T("Hash is required")
	}

	if len(errors) > 0 {
		return c.ErrValidation(l.T("Validation failed"), errors)
	}

	return nil
}

func (sied SignInEmailData) validate(l *localizer.Localizer) error {
	errors := make(map[string]string)

	if !emailRegex.MatchString(sied.Email) {
		errors["email"] = l.T("Email is invalid")
	}

	if len(sied.Email) == 0 {
		errors["email"] = l.T("Email is required")
	}

	if len(sied.Code) == 0 {
		errors["code"] = l.T("Email code is required")
	}

	if len(sied.Hash) == 0 {
		errors["hash"] = l.T("Hash is required")
	}

	if len(errors) > 0 {
		return c.ErrValidation(l.T("Validation failed"), errors)
	}

	return nil
}
