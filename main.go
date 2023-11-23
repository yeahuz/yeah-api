package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeahuz/yeah-api/auth"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/db"
)

var (
	POSTGRES_URI = os.Getenv("POSTGRES_URI")
	ADDR         = c.GetEnvStr("ADDR", ":3000")
)

func main() {
	var err error
	db.Pool, err = pgxpool.New(context.Background(), POSTGRES_URI)
	if err != nil {
		panic(err)
	}

	defer db.Pool.Close()

	mux := http.NewServeMux()

	mux.Handle("/auth.sendPhoneCode", c.LocalizerMiddleware(c.MakeHandler(auth.HandleSendPhoneCode, http.MethodPost)))
	mux.Handle("/auth.sendEmailCode", c.MakeHandler(auth.HandleSendEmailCode, http.MethodPost))
	mux.Handle("/auth.signIn", c.MakeHandler(auth.HandleSignIn, http.MethodPost))
	mux.Handle("/auth.signUp", c.MakeHandler(auth.HandleSignUp, http.MethodPost))
	fmt.Printf("Server started at %s\n", ADDR)
	log.Fatal(http.ListenAndServe(ADDR, mux))
}
