package credential

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func NewCreateRequest(userID string, display, name string) (*CreateRequest, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}

	createRequest := &CreateRequest{
		Challenge: challenge,
		Rp: Rp{
			ID:   config.Config.RpID,
			Name: config.Config.RpName,
		},
		User: User{
			ID:          base64.RawURLEncoding.EncodeToString([]byte(userID)),
			DisplayName: display,
			Name:        name,
		},
		Timeout:     60000,
		Attestation: AttestationNone,
		PubKeyCredParams: []PubKeyCredParam{
			{Type: PubKeyCredPubKey, Alg: -7},
		},
	}

	return createRequest, nil
}

func NewGetRequest(credentials []allowedCredential) (*GetRequest, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}

	getRequest := &GetRequest{
		Challenge:        challenge,
		RpID:             config.Config.RpID,
		Timeout:          60000,
		AllowCredentials: credentials,
	}

	return getRequest, nil
}

func New(opts Opts) *Credential {
	return &Credential{
		Opts: opts,
	}
}

func (c *Credential) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credentials (title, credential_id, pubkey, counter, user_id, transports) values ($1, $2, $3, $4, $5, $6) returning id",
		c.Title, c.CredentialID, c.PubKey, c.Counter, c.UserID, c.Transports,
	).Scan(&c.ID)

	if err != nil {
		// TODO: handle errors properly
		return errors.Internal
	}
	return nil
}

func GetById(ctx context.Context, credId string) (*Credential, error) {
	var credential Credential
	err := db.Pool.QueryRow(ctx,
		"select id, credential_id, title, transports, user_id from credentials where credential_id = $1", credId).Scan(
		&credential.ID, &credential.Title, &credential.Transports, &credential.UserID)

	if err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	return &credential, nil
}

func GetAllowedCredentials(ctx context.Context, userID string) ([]allowedCredential, error) {
	credentials := make([]allowedCredential, 0)
	rows, err := db.Pool.Query(ctx, "select credential_id, transports type from credentials where user_id = $1", userID)
	defer rows.Close()
	if err != nil {
		return nil, errors.Internal
	}

	for rows.Next() {
		var crd allowedCredential
		if err := rows.Scan(&crd.ID, &crd.Transports, &crd.Type); err != nil {
			return nil, errors.Internal
		}

		credentials = append(credentials, crd)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal
	}

	return credentials, nil
}

func GetAll(ctx context.Context) (*Credentials, error) {
	creds := make([]Credential, 0)

	rows, err := db.Pool.Query(ctx, "select id, credential_id, title, transports, user_id from credentials order by id desc")

	defer rows.Close()
	if err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	for rows.Next() {
		var crd Credential
		if err := rows.Scan(&crd.ID, &crd.CredentialID, &crd.Title, &crd.Transports, &crd.UserID); err != nil {
			// TODO: handle errors properly
			return nil, errors.Internal
		}
		creds = append(creds, crd)
	}

	if err := rows.Err(); err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	list := &Credentials{Credentials: creds, Count: len(creds)}

	return list, nil
}

func (cr *CreateRequest) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credential_requests (challenge, type) values ($1, $2) returning id",
		cr.Challenge, "create",
	).Scan(&cr.ID)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func (gr *GetRequest) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credential_requests (challenge, type) values ($1, $2) returning id",
		gr.Challenge, "get",
	).Scan(&gr.ID)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func GetRequestById(ctx context.Context, id string) (*Request, error) {
	var request Request

	err := db.Pool.QueryRow(ctx,
		"select type, challenge, used from credential_requests where id = $1",
		id,
	).Scan(&request.Type, &request.Challenge, &request.Used)

	if err != nil {
		return nil, errors.Internal
	}

	return &request, nil
}

func (r Request) VerifyClientData(data string) (*clientData, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.Internal
	}

	var clientData clientData

	if err := json.NewDecoder(bytes.NewReader(decoded)).Decode(&clientData); err != nil {
		return nil, errors.Internal
	}

	if clientData.typ != r.Type {
		return nil, errors.NewBadRequest(l.T("Invalid credential type"))
	}

	if clientData.challenge != r.Challenge {
		return nil, errors.NewBadRequest(l.T("Challenges don't match"))
	}

	if clientData.origin != config.Config.RpID {
		return nil, errors.NewBadRequest(l.T("Origin is invalid"))
	}

	return &clientData, nil
}

func generateChallenge() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}