package cqrs

const (
	emailCodeSent      = "auth.emailCodeSent"
	phoneCodeSent      = "auth.phoneCodeSent"
	loginTokenRejected = "auth.loginTokenRejected"
	loginTokenAccepted = "auth.loginTokenAccepted"
)

type LoginTokenRejectedEvent struct {
	name  string
	Token string `json:"token"`
}

type LoginTokenAcceptedEvent struct {
	name  string
	Token string `json:"token"`
}

type EmailCodeSentEvent struct {
	name  string
	Email string `json:"email"`
}

type PhoneCodeSentEvent struct {
	name        string
	PhoneNumber string `json:"phone_number"`
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

func NewLoginTokenRejectedEvent(token string) LoginTokenRejectedEvent {
	return LoginTokenRejectedEvent{
		name:  loginTokenRejected,
		Token: token,
	}
}

func NewLoginTokenAcceptedEvent(token string) LoginTokenAcceptedEvent {
	return LoginTokenAcceptedEvent{
		name:  loginTokenAccepted,
		Token: token,
	}
}

func (ecse EmailCodeSentEvent) Name() string {
	return ecse.name
}

func (pcse PhoneCodeSentEvent) Name() string {
	return pcse.name
}

func (l LoginTokenAcceptedEvent) Name() string {
	return l.name
}

func (l LoginTokenRejectedEvent) Name() string {
	return l.name
}
