package auth

import (
	"github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

func (pcd PhoneCodeData) validate(l *localizer.Localizer) error {
	if len(pcd.PhoneNumber) != 13 {
		return common.ErrBadRequest(l.T("Phone number is invalid"))
	}

	return nil
}

func (ecd EmailCodeData) validate() error {
	if len(ecd.email) != 13 {
		return common.ErrBadRequest("Email is invalid")
	}

	return nil
}
