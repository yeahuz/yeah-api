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
