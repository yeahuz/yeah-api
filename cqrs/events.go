package cqrs

const (
	emailCodeSent = "EmailCodeSent"
	phoneCodeSent = "PhoneCodeSent"
)

type EmailCodeSentEvent struct {
	name  string
	Email string
}

type PhoneCodeSentEvent struct {
	name        string
	PhoneNumber string
}

func NewEmailCodeSentEvent(email string) EmailCodeSentEvent {
	return EmailCodeSentEvent{
		name:  emailCodeSent,
		Email: email,
	}
}

func NewPhoneCodeSentEvent(phoneNumber string) PhoneCodeSentEvent {
	return PhoneCodeSentEvent{
		name:        phoneCodeSent,
		PhoneNumber: phoneNumber,
	}
}

func (ecse EmailCodeSentEvent) Name() string {
	return ecse.name
}

func (pcse PhoneCodeSentEvent) Name() string {
	return pcse.name
}
