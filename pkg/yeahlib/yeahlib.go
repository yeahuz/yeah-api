package yeahlib

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/inmem"
	"github.com/yeahuz/yeah-api/postgres"
)

const (
	defaultConfigPath = "~/yeahlib.conf"
)

type Lib struct {
	AuthService     yeahapi.AuthService
	UserService     yeahapi.UserService
	ListingService  yeahapi.ListingService
	KVService       yeahapi.KVService
	ClientService   yeahapi.ClientService
	CategoryService yeahapi.CategoryService
}

type Config struct {
	DB struct {
		Postgres string `toml:"postgres"`
	} `toml:"db"`

	HighwayHash struct {
		Key string `toml:"key"`
	} `toml:"highwayhash"`
}

func New(ctx context.Context) (*Lib, error) {
	config, err := ReadConfigFile(defaultConfigPath)
	pool, err := pgxpool.New(ctx, config.DB.Postgres)
	if err != nil {
		return nil, err
	}

	argonHasher := inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	highwayHasher := inmem.NewHighwayHasher(config.HighwayHash.Key)

	authService := postgres.NewAuthService(pool, argonHasher, highwayHasher)
	userService := postgres.NewUserService(pool)
	listingService := postgres.NewListingService(pool)
	kvService := postgres.NewKVService(pool)
	clientService := postgres.NewClientService(pool, argonHasher)
	categoryService := postgres.NewCategoryService(pool)

	lib := &Lib{
		AuthService:     authService,
		UserService:     userService,
		ListingService:  listingService,
		KVService:       kvService,
		ClientService:   clientService,
		CategoryService: categoryService,
	}

	return lib, nil
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
