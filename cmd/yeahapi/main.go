package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/http"
	"github.com/yeahuz/yeah-api/nats"
	"github.com/yeahuz/yeah-api/postgres"
)

func main() {
}

type Main struct {
	Config     Config
	ConfigPath string
	Pool       *pgxpool.Pool
	Server     *http.Server
}

const (
	defaultConfigPath = "~/yeahapi.conf"
)

func NewMain() *Main {
	return &Main{
		Config:     Config{},
		Server:     http.NewServer(),
		ConfigPath: defaultConfigPath,
	}
}

func (m *Main) Run(ctx context.Context) (err error) {
	if m.Pool, err = pgxpool.New(ctx, m.Config.DB.DSN); err != nil {
		return err
	}

	authService := postgres.NewAuthService(m.Pool)
	userService := postgres.NewUserService(m.Pool)
	cqrsService := nats.NewCQRSService(ctx, yeahapi.CQRSConfig{
		NatsURL:       m.Config.Nats.URL,
		NatsAuthToken: m.Config.Nats.AuthToken,
		Streams:       map[string][]string{},
	})

	// m.Server.AuthService = authService
	m.Server.UserService = userService
	m.Server.AuthService = authService
	m.Server.CQRSService = cqrsService

	return err
}

func (m *Main) Close() error {
	if m.Server != nil {
	}

	if m.Pool != nil {
		m.Pool.Close()
	}

	return nil
}

func (m *Main) ParseFlags(ctx context.Context, args []string) error {
	// fs := flag.NewFlagSet("")
	return nil
}

type Config struct {
	DB struct {
		DSN string `toml:"dsn"`
	} `toml:"db"`

	HTTP struct {
		Addr string `toml:"addr"`
	} `toml:"http"`

	AWS struct {
		Secret string `toml:"secret"`
		Key    string `toml:"key"`
	} `toml:"aws"`

	Nats struct {
		AuthToken string            `toml:"auth-token"`
		URL       string            `toml:"url"`
		Streams   map[string]string `toml:"streams"`
	} `toml:"nats"`

	Google struct {
		ClientID     string `toml:"client-id"`
		ClientSecret string `toml:"client-secret"`
		RedirectURL  string `toml:"redirect-url"`
	} `toml:"google"`
}

func ReadConfigFile(filename string) (Config, error) {
	var config Config
	if buf, err := ioutil.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

func expand(path string) (string, error) {
	// Ignore if path has no leading tilde.
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path, nil
	}

	// Fetch the current user to determine the home path.
	u, err := user.Current()
	if err != nil {
		return path, err
	} else if u.HomeDir == "" {
		return path, fmt.Errorf("home directory unset")
	}

	if path == "~" {
		return u.HomeDir, nil
	}
	return filepath.Join(u.HomeDir, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
}

// POSTGRES_URI=postgresql://jcbbb:2157132aA*codes@localhost:5432/needs-api-dev
// HIGHWAY_HASH_KEY=fad21f136da3d007d9b68bb7a161918e3067861e692175f76888c3210a0a485c
// ADDR=:3000
// AWS_KEY=AKIAT5IEKHYOJZKSGK6F
// AWS_SECRET=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
// NATS_AUTH_TOKEN=10cee6c8f524cab9795af019c62bdb60cdd30e7261df164adb577f857d93d2de
// SMS_API_URL=https://notify.eskiz.uz/api
// SMS_API_EMAIL=avaz28082000@gmail.com
// SMS_API_PASSWORD=WC0jM1ztmAdeudblHWYjP0dEI88A3AHS7EAZ3dvI
// SIGNING_SECRET=a0967db797a60d3ad5d72d906856386bbd1fa98526ff27d9891ad498656eb167
// ORIGIN=http://localhost:8080
// GOOGLE_OAUTH_CLIENT_SECRET=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
// GOOGLE_OAUTH_CLIENT_ID=600896005100-u88qclbcs8clv5mllc2h7dbl9cngppb5.apps.googleusercontent.com
// GOOGLE_OAUTH_REDIRECT_URL=http://localhost:3001/auth/google
