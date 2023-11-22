package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yeahuz/yeah-api/auth"
	"github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/db"
)

var (
	POSTGRES_URI = os.Getenv("POSTGRES_URI")
	ADDR         = common.GetEnvStr("ADDR", ":3000")
)

func main() {
	var err error
	db.Pool, err = pgxpool.New(context.Background(), POSTGRES_URI)
	if err != nil {
		panic(err)
	}

	defer db.Pool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("auth.sendPhoneCode", common.MakeHandlerFunc(auth.HandleSendPhoneCode, http.MethodPost))
	mux.HandleFunc("auth.sendEmailCode", common.MakeHandlerFunc(auth.HandleSendEmailCode, http.MethodPost))
	mux.HandleFunc("auth.signIn", common.MakeHandlerFunc(auth.HandleSignIn, http.MethodPost))
	mux.HandleFunc("auth.signUp", common.MakeHandlerFunc(auth.HandleSignUp, http.MethodPost))

	fmt.Printf("Server started at %s\n", ADDR)
	log.Fatal(http.ListenAndServe(ADDR, mux))
}
