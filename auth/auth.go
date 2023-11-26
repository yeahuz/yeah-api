package auth

import (
	"net/http"
	"regexp"

	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var (
	emailRegex = regexp.MustCompile(`(?i)^(([^<>()[\].,;:\s@"]+(\.[^<>()[\].,;:\s@"]+)*)|(".+"))@(([^<>()[\].,;:\s@"]+\.)+[^<>()[\].,;:\s@"]{2,})$`)
	l          = localizer.GetDefault()
)

func (pcd PhoneCodeData) validate() error {
	if len(pcd.PhoneNumber) == 0 {
		return errors.ErrBadRequest{Message: l.T("Phone number is required"), StatusCode: http.StatusBadRequest}
	}

	if len(pcd.PhoneNumber) != 13 {
		return errors.ErrBadRequest{Message: l.T("Phone number is invalid"), StatusCode: http.StatusBadRequest}
	}

	return nil
}

func (ecd EmailCodeData) validate() error {
	if len(ecd.Email) == 0 {
		return errors.ErrBadRequest{Message: l.T("Email is required"), StatusCode: http.StatusBadRequest}
	}

	if !emailRegex.MatchString(ecd.Email) {
		return errors.ErrBadRequest{Message: l.T("Email is invalid"), StatusCode: http.StatusBadRequest}
	}

	return nil
}

func (sipd SignInPhoneData) validate() error {
	errs := make(map[string]string)

	if len(sipd.PhoneNumber) == 0 {
		errs["phone_number"] = l.T("Phone number is required")
	}

	if len(sipd.PhoneNumber) != 13 {
		errs["phone_number"] = l.T("Phone number is invalid")
	}

	if len(sipd.Code) == 0 {
		errs["code"] = l.T("Phone code is required")
	}

	if len(sipd.Hash) == 0 {
		errs["hash"] = l.T("Hash is required")
	}

	if len(errs) > 0 {
		return errors.ErrValidation{Message: l.T("Validation failed"), Errors: errs, StatusCode: http.StatusUnprocessableEntity}
	}

	return nil
}

func (sied SignInEmailData) validate() error {
	errs := make(map[string]string)

	if !emailRegex.MatchString(sied.Email) {
		errs["email"] = l.T("Email is invalid")
	}

	if len(sied.Email) == 0 {
		errs["email"] = l.T("Email is required")
	}

	if len(sied.Code) == 0 {
		errs["code"] = l.T("Email code is required")
	}

	if len(sied.Hash) == 0 {
		errs["hash"] = l.T("Hash is required")
	}

	if len(errs) > 0 {
		return errors.ErrValidation{Message: l.T("Validation failed"), Errors: errs, StatusCode: http.StatusUnprocessableEntity}
	}

	return nil
}
