package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeahuz/yeah-api/auth"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
)

func main() {
	var err error
	config := config.Load()
	db.Pool, err = pgxpool.New(context.Background(), config.PostgresURI)
	if err != nil {
		panic(err)
	}

	defer db.Pool.Close()
	mux := http.NewServeMux()
	mux.Handle("/auth.sendPhoneCode", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSendPhoneCode, http.MethodPost)))
	mux.Handle("/auth.sendEmailCode", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSendEmailCode, http.MethodPost)))
	mux.Handle("/auth.signInWithEmail", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignInWithEmail, http.MethodPost)))
	mux.Handle("/auth.signInWithPhone", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignInWithPhone, http.MethodPost)))
	mux.Handle("/auth.signUp", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSignUp, http.MethodPost)))
	fmt.Printf("Server started at %s\n", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, mux))
}
