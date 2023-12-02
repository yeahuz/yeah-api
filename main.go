package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeahuz/yeah-api/auth"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/cqrs"
	"github.com/yeahuz/yeah-api/db"
)

func main() {
	var err error
	config := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*60)
	defer cancel()

	cleanup, err := cqrs.Setup(ctx, cqrs.SetupOpts{
		NatsURL:       config.NatsURL,
		NatsAuthToken: config.NatsAuthToken,
		AwsKey:        config.AwsKey,
		AwsSecret:     config.AwsSecret,
	})

	defer cleanup()
	if err != nil {
		log.Fatal(err)
	}

	db.Pool, err = pgxpool.New(context.Background(), config.PostgresURI)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Pool.Close()
	mux := http.NewServeMux()
	mux.Handle("/auth.sendPhoneCode", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSendPhoneCode, http.MethodPost)))
	mux.Handle("/auth.sendEmailCode", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSendEmailCode, http.MethodPost)))
	mux.Handle("/auth.signInWithEmail", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignInWithEmail, http.MethodPost)))
	mux.Handle("/auth.signInWithPhone", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignInWithPhone, http.MethodPost)))
	mux.Handle("/auth.signUpWithEmail", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignUpWithEmail, http.MethodPost)))
	mux.Handle("/auth.signUpWithPhone", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignUpWithPhone, http.MethodPost)))
	fmt.Printf("Server started at %s\n", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, mux))
}
