package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yeahuz/yeah-api/auth"
	"github.com/yeahuz/yeah-api/common"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("auth.sendPhoneCode", common.MakeHandlerFunc(auth.HandleSendPhoneCode, http.MethodPost))
	mux.HandleFunc("auth.sendEmailCode", common.MakeHandlerFunc(auth.HandleSendEmailCode, http.MethodPost))
	mux.HandleFunc("auth.signIn", common.MakeHandlerFunc(auth.HandleSignIn, http.MethodPost))
	mux.HandleFunc("auth.signUp", common.MakeHandlerFunc(auth.HandleSignUp, http.MethodPost))

	fmt.Printf("Starting to listen on port :3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
