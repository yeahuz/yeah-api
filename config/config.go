package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	c "github.com/yeahuz/yeah-api/common"
)

type config struct {
	PostgresURI    string
	Addr           string
	HighwayHashKey string
	AwsKey         string
	AwsSecret      string
	NatsURL        string
}

var Config *config

func Load() *config {
	env := c.GetEnvStr("YEAH_API_ENV", "development")

	godotenv.Load(".env." + env)

	Config = &config{
		PostgresURI:    os.Getenv("POSTGRES_URI"),
		Addr:           c.GetEnvStr("ADDR", ":3000"),
		HighwayHashKey: os.Getenv("HIGHWAY_HASH_KEY"),
		AwsKey:         os.Getenv("AWS_KEY"),
		AwsSecret:      os.Getenv("AWS_SECRET"),
		NatsURL:        c.GetEnvStr("NATS_URL", nats.DefaultURL),
	}

	return Config
}
