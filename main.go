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
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/cqrs"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/listing"
	"github.com/yeahuz/yeah-api/postgres"
	"github.com/yeahuz/yeah-api/smsclient"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func main() {
	var err error
	config := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*60)
	defer cancel()

	smsClient := smsclient.New(smsclient.Opts{
		SmsApiBaseUrl:   config.SmsApiBaseUrl,
		SmsApiEmail:     config.SmsApiEmail,
		SmsApiPassword:  config.SmsApiPassword,
		TimeoutDuration: time.Second * 30,
	})

	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AwsKey, config.AwsSecret, "")),
		awsconfig.WithRegion("eu-north-1"),
	)

	sesClient := ses.NewFromConfig(cfg)

	cleanup, cqrsClient, err := cqrs.Setup(ctx, cqrs.SetupOpts{
		NatsURL:       config.NatsURL,
		NatsAuthToken: config.NatsAuthToken,
		SmsClient:     smsClient,
		SesClient:     sesClient,
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
	userService := &postgres.UserService{
		Pool: db.Pool,
	}

	mux := http.NewServeMux()

	mux.Handle("POST /auth.createOAuthFlow", localized(
		clientOnly(auth.HandleCreateOAuthFlow()),
	))
	mux.Handle("POST /auth.signInWithGoogle", localized(
		clientOnly(auth.HandleSignInWithGoogle(userService)),
	))
	mux.Handle("POST /auth.signInWithTelegram", localized(
		clientOnly(auth.HandleSignInWithTelegram()),
	))
	mux.Handle("POST /auth.sendPhoneCode", localized(
		clientOnly(auth.HandleSendPhoneCode(cqrsClient)),
	))
	mux.Handle("/auth.sendEmailCode", localized(
		clientOnly(auth.HandleSendEmailCode(cqrsClient)),
	))
	mux.Handle("/auth.signInWithEmail", localized(
		clientOnly(auth.HandleSignInWithEmail(userService)),
	))
	mux.Handle("/auth.signInWithPhone", localized(
		clientOnly(auth.HandleSignInWithPhone(userService)),
	))
	mux.Handle("/auth.signUpWithEmail", localized(
		clientOnly(auth.HandleSignUpWithEmail(userService)),
	))
	mux.Handle("/auth.signUpWithPhone", localized(
		clientOnly(auth.HandleSignUpWithPhone(userService)),
	))
	mux.Handle("POST /auth.createLoginToken", localized(
		clientOnly(auth.HandleCreateLoginToken()),
	))
	mux.Handle("POST /auth.acceptLoginToken", localized(
		userOnly(auth.HandleAcceptLoginToken(cqrsClient)),
	))
	mux.Handle("POST /auth.rejectLoginToken", localized(
		userOnly(auth.HandleRejectLoginToken(cqrsClient)),
	))
	mux.Handle("POST /auth.scanLoginToken", localized(
		userOnly(auth.HandleScanLoginToken()),
	))
	mux.Handle("POST /auth.logOut", localized(
		userOnly(auth.HandleLogOut()),
	))

	mux.Handle("POST /credentials.pubKeyCreateRequest", localized(
		userOnly(auth.HandlePubKeyCreateRequest(userService)),
	))
	mux.Handle("POST /credentials.pubKeyGetRequest", localized(
		clientOnly(auth.HandlePubKeyGetRequest(userService)),
	))

	mux.Handle("POST /credentials.createPubKey", localized(
		clientOnly(auth.HandleCreatePubKey()),
	))
	mux.Handle("POST /credentials.verifyPubKey", localized(
		clientOnly(auth.HandleVerifyPubKey()),
	))

	mux.Handle("POST /listing.createListing", localized(
		userOnly(listing.HandleCreateListing()),
	))

	fmt.Printf("Server started at %s\n", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, mux))
}
