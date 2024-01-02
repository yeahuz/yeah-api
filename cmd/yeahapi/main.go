package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"

	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/aws"
	"github.com/yeahuz/yeah-api/eskiz"
	"github.com/yeahuz/yeah-api/http"
	"github.com/yeahuz/yeah-api/inmem"
	"github.com/yeahuz/yeah-api/nats"
	"github.com/yeahuz/yeah-api/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() { <-c; cancel() }()

	m := NewMain()

	if err := m.ParseFlags(ctx, os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := m.Run(ctx); err != nil {
		m.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	Config     *Config
	ConfigPath string
	Pool       *pgxpool.Pool
	Server     *http.Server
}

const (
	defaultConfigPath = "~/yeahapi.conf"
)

func NewMain() *Main {
	return &Main{
		Config:     &Config{},
		Server:     http.NewServer(),
		ConfigPath: defaultConfigPath,
	}
}

func (m *Main) Run(ctx context.Context) (err error) {
	if m.Pool, err = pgxpool.New(ctx, m.Config.DB.Postgres); err != nil {
		return err
	}

	argonHasher := inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	highwayHasher := inmem.NewHighwayHasher(m.Config.HighwayHash.Key)

	authService := postgres.NewAuthService(m.Pool, argonHasher, highwayHasher)
	userService := postgres.NewUserService(m.Pool)
	listingService := postgres.NewListingService(m.Pool)
	kvService := postgres.NewKVService(m.Pool)
	localizerService := yeahapi.NewLocalizerService("en")

	cqrsService, err := nats.NewCQRSService(ctx, yeahapi.CQRSConfig{
		NatsURL:       m.Config.Nats.URL,
		NatsAuthToken: m.Config.Nats.AuthToken,
		Streams:       m.Config.Nats.Streams,
	})

	if err != nil {
		return err
	}

	awsconfig, err := awsconf.LoadDefaultConfig(ctx,
		awsconf.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(m.Config.AWS.Key, m.Config.AWS.Secret, "")),
		awsconf.WithRegion("en-north-1"),
	)

	emailService := aws.NewEmailService(awsconfig, cqrsService)
	smsService := eskiz.NewSmsService(m.Config.Eskiz.Email, m.Config.Eskiz.Password, m.Config.Eskiz.BaseURL, cqrsService)

	cqrsService.Handle("auth.sendEmailCode", emailService.SendEmailCode)
	cqrsService.Handle("auth.sendPhoneCode", smsService.SendSmsCode)

	m.Server.Addr = m.Config.HTTP.Addr

	m.Server.UserService = userService
	m.Server.AuthService = authService
	m.Server.LocalizerService = localizerService
	m.Server.ListingService = listingService
	m.Server.CQRSService = cqrsService
	m.Server.KVService = kvService

	if err := m.Server.Open(); err != nil {
		return err
	}

	return nil
}

func (m *Main) Close() error {
	if m.Pool != nil {
		m.Pool.Close()
	}

	if m.Server != nil {
		if err := m.Server.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Main) ParseFlags(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("api", flag.ContinueOnError)

	fs.StringVar(&m.ConfigPath, "config", defaultConfigPath, "config path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	configPath, err := expand(m.ConfigPath)
	if err != nil {
		return err
	}

	config, err := ReadConfigFile(configPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\n", m.ConfigPath)
	} else if err != nil {
		return err
	}

	m.Config = config

	return nil
}

type Config struct {
	DB struct {
		Postgres string `toml:"postgres"`
	} `toml:"db"`

	HTTP struct {
		Addr string `toml:"addr"`
	} `toml:"http"`

	AWS struct {
		Secret string `toml:"secret"`
		Key    string `toml:"key"`
	} `toml:"aws"`

	Eskiz struct {
		Email    string `toml:"email"`
		Password string `toml:"password"`
		BaseURL  string `toml:"base-url"`
	} `toml:"eskiz"`

	HighwayHash struct {
		Key string `toml:"key"`
	} `toml:"highwayhash"`

	Nats struct {
		AuthToken string              `toml:"auth-token"`
		URL       string              `toml:"url"`
		Streams   map[string][]string `toml:"streams"`
	} `toml:"nats"`

	Google struct {
		ClientID     string `toml:"client-id"`
		ClientSecret string `toml:"client-secret"`
		RedirectURL  string `toml:"redirect-url"`
	} `toml:"google"`
}

func ReadConfigFile(filename string) (*Config, error) {
	var config Config
	if buf, err := os.ReadFile(filename); err != nil {
		return &config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return &config, err
	}

	return &config, nil
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
