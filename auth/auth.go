package auth

import (
	"github.com/yeahuz/yeah-api/common"
)

func (pcd PhoneCodeData) validate() error {
	if len(pcd.PhoneNumber) != 13 {
		return common.ErrBadRequest("Phone number is invalid")
	}

	return nil
}

func (ecd EmailCodeData) validate() error {
	if len(ecd.email) != 13 {
		return common.ErrBadRequest("Email is invalid")
	}

	return nil
}
