package cqrs

const (
	sendPhoneCode = "auth.sendPhoneCode"
	sendEmailCode = "auth.sendEmailCode"
)

type SendEmailCodeCommand struct {
	name  string
	Email string `json:"email"`
	Code  string `json:"code"`
}

type SendPhoneCodeCommand struct {
	name        string
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

func NewSendEmailCodeCommand(email string, code string) SendEmailCodeCommand {
	return SendEmailCodeCommand{
		name:  sendEmailCode,
		Email: email,
		Code:  code,
	}
}

func NewSendPhoneCodeCommand(phoneNumber string, code string) SendPhoneCodeCommand {
	return SendPhoneCodeCommand{
		name:        sendPhoneCode,
		PhoneNumber: phoneNumber,
		Code:        code,
	}
}

func (secc SendEmailCodeCommand) Name() string {
	return secc.name
}

func (spcc SendPhoneCodeCommand) Name() string {
	return spcc.name
}
