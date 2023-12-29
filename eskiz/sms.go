package eskiz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	yeahapi "github.com/yeahuz/yeah-api"
)

type SmsService struct {
	email      string
	password   string
	baseUrl    string
	token      string
	refreshing atomic.Bool
	cond       *sync.Cond
	cqrssrv    yeahapi.CQRSService
}

type sendSmsOutput struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewSmsService(email, password, baseUrl string, cqrssrv yeahapi.CQRSService) *SmsService {
	return &SmsService{
		email:    email,
		password: password,
		baseUrl:  baseUrl,
		cond:     sync.NewCond(&sync.Mutex{}),
		cqrssrv:  cqrssrv,
	}
}

func (s *SmsService) getToken(ctx context.Context) error {
	form := url.Values{
		"email":    {s.email},
		"password": {s.password},
	}

	req, err := http.NewRequest("POST", s.baseUrl+"/auth/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer req.Body.Close()

	var tokenResp struct {
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
		TokenType string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	s.token = tokenResp.Data.Token
	s.cond.Signal()

	return nil
}

func (s *SmsService) request(ctx context.Context, method, path string, body io.Reader, v any) error {
	req, err := http.NewRequestWithContext(ctx, method, s.baseUrl+path, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if !s.refreshing.Load() {
			s.refreshing.Store(true)
			s.cond.L.Lock()
			go s.getToken(ctx)
		}
		s.cond.Wait()
		s.cond.L.Unlock()
		s.refreshing.Store(false)
		return s.request(ctx, method, path, body, v)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed wit status code of %d\n", resp.StatusCode)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func (s *SmsService) send(ctx context.Context, to, message string) (*sendSmsOutput, error) {
	form := url.Values{
		"message":      {message},
		"from":         {"4546"},
		"mobile_phone": {to},
	}
	var output sendSmsOutput
	if err := s.request(ctx, "POST", "/message/sms/send", strings.NewReader(form.Encode()), &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func (s *SmsService) SendSmsCode(m jetstream.Msg) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var cmd yeahapi.SendPhoneCodeCmd
	if err := json.Unmarshal(m.Data(), &cmd); err != nil {
		return err
	}

	_, err := s.send(ctx, cmd.PhoneNumber[1:], fmt.Sprintf("Your verification code is %s. It expires in 15 minutes. Do not share this code! @needs.uz #%s", cmd.Code, cmd.Code))
	if err != nil {
		return err
	}

	s.cqrssrv.Publish(ctx, yeahapi.NewPhoneCodeSentEvent(cmd.PhoneNumber))
	return nil
}
