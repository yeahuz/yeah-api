package config

import (
	"os"

	"github.com/joho/godotenv"
	c "github.com/yeahuz/yeah-api/common"
)

type config struct {
	PostgresURI    string
	Addr           string
	HighwayHashKey string
}

var Config config

func Load() config {
	env := c.GetEnvStr("YEAH_API_ENV", "development")

	godotenv.Load(".env." + env)

	Config = config{
		PostgresURI:    os.Getenv("POSTGRES_URI"),
		Addr:           c.GetEnvStr("ADDR", ":3000"),
		HighwayHashKey: os.Getenv("HIGHWAY_HASH_KEY"),
	}

	return Config
}
