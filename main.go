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
	"github.com/yeahuz/yeah-api/client"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/cqrs"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/localizer"
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

	cleanup, err := cqrs.Setup(ctx, cqrs.SetupOpts{
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
	mux := http.NewServeMux()
	mux.Handle("/auth.createOAuthFlow", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleCreateOAuthFlow, http.MethodPost)),
	))
	mux.Handle("/auth.processOAuthCallback", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleOAuthCallback, http.MethodPost)),
	))
	mux.Handle("/auth.sendPhoneCode", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleSendPhoneCode, http.MethodPost)),
	))
	mux.Handle("/auth.sendEmailCode", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleSendEmailCode, http.MethodPost)),
	))
	mux.Handle("/auth.signInWithEmail", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleSignInWithEmail, http.MethodPost)),
	))
	mux.Handle("/auth.signInWithPhone", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleSignInWithPhone, http.MethodPost)),
	))
	mux.Handle("/auth.signUpWithEmail", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleSignUpWithEmail, http.MethodPost)),
	))
	mux.Handle("/auth.signUpWithPhone", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleSignUpWithPhone, http.MethodPost)),
	))
	mux.Handle("/auth.createLoginToken", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleCreateLoginToken, http.MethodPost)),
	))
	mux.Handle("/auth.acceptLoginToken", localizer.Middleware(
		auth.Middleware(c.MakeHandler(auth.HandleAcceptLoginToken, http.MethodPost)),
	))
	mux.Handle("/auth.rejectLoginToken", localizer.Middleware(
		auth.Middleware(c.MakeHandler(auth.HandleRejectLoginToken, http.MethodPost)),
	))
	mux.Handle("/auth.scanLoginToken", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleScanLoginToken, http.MethodPost)),
	))
	mux.Handle("/auth.logOut", localizer.Middleware(
		auth.Middleware(c.MakeHandler(auth.HandleLogOut, http.MethodPost)),
	))
	mux.Handle("/credentials.pubKeyCreateRequest", localizer.Middleware(
		auth.Middleware(c.MakeHandler(auth.HandlePubKeyCreateRequest, http.MethodPost)),
	))
	mux.Handle("/credentials.pubKeyGetRequest", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandlePubKeyGetRequest, http.MethodPost)),
	))
	mux.Handle("/credentials.createPubKey", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleCreatePubKey, http.MethodPost)),
	))
	mux.Handle("/credentials.verifyPubKey", localizer.Middleware(
		client.Middleware(c.MakeHandler(auth.HandleVerifyPubKey, http.MethodPost)),
	))

	fmt.Printf("Server started at %s\n", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, mux))
}
