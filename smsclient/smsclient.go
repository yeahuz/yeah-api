package smsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	Email           string
	Password        string
	BaseUrl         string
	TimeoutDuration time.Duration
	token           string
	refreshing      atomic.Bool
	cond            *sync.Cond
}

type Opts struct {
	SmsApiBaseUrl   string
	SmsApiEmail     string
	SmsApiPassword  string
	TimeoutDuration time.Duration
}

func New(opts Opts) *Client {
	return &Client{
		Email:           opts.SmsApiEmail,
		BaseUrl:         opts.SmsApiBaseUrl,
		Password:        opts.SmsApiPassword,
		TimeoutDuration: opts.TimeoutDuration,
		cond:            sync.NewCond(&sync.Mutex{}),
	}
}

type tokenData struct {
	Token string `json:"token"`
}

type tokenResponse struct {
	Message   string    `json:"message"`
	Data      tokenData `json:"data"`
	TokenType string    `json:"token_type"`
}

func (c *Client) getToken(ctx context.Context) error {
	form := url.Values{
		"email":    {c.Email},
		"password": {c.Password},
	}
	req, err := http.NewRequest("POST", c.BaseUrl+"/auth/login", strings.NewReader(form.Encode()))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	var tokenResponse tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return err
	}

	c.token = tokenResponse.Data.Token
	c.cond.Signal()

	return nil
}

func (c *Client) request(method, path, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.TimeoutDuration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, c.BaseUrl+path, strings.NewReader(data))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if !c.refreshing.Load() {
			c.refreshing.Store(true)
			c.cond.L.Lock()
			go c.getToken(ctx)
		}
		c.cond.Wait()
		c.cond.L.Unlock()
		c.refreshing.Store(false)
		return c.request(method, path, data)
	}

	if err != nil {
		return err
	}

	//TODO: properly handle errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request(): Request failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Send(to, message string) error {
	form := url.Values{
		"message":      {message},
		"from":         {"4546"},
		"mobile_phone": {to},
	}

	if err := c.request("POST", "/message/sms/send", form.Encode()); err != nil {
		return err
	}

	return nil
}
