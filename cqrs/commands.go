package cqrs

const (
	sendPhoneCode = "SendPhoneCode"
	sendEmailCode = "SendEmailCode"
)

type Command interface {
	Name() string
}

type SendEmailCodeCommand struct {
	name string
	Recv string
	Code string
}

type SendPhoneCodeCommand struct {
	name string
	Recv string
	Code string
}

func NewSendEmailCodeCommand(recv string, code string) SendEmailCodeCommand {
	return SendEmailCodeCommand{
		name: sendEmailCode,
		Recv: recv,
		Code: code,
	}
}

func NewSendPhoneCodeCommand(recv string, code string) SendPhoneCodeCommand {
	return SendPhoneCodeCommand{
		name: sendPhoneCode,
		Recv: recv,
		Code: code,
	}
}

func (secc SendEmailCodeCommand) Name() string {
	return secc.name
}

func (spcc SendPhoneCodeCommand) Name() string {
	return spcc.name
}
