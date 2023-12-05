package credential

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"

	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
)

type AttestationType string

const (
	AttestationDirect   AttestationType = "direct"
	AttestationIndirect                 = "indirect"
	AttestationNone                     = "none"
)

type PubKeyCredParam struct {
	Alg  int    `json:"alg"`
	Kind string `json:"type"`
}
type Opts struct {
	CredentialID string
	Title        string
	PubKey       string
	Counter      int
	UserID       int
	Transports   []string
}

type Rp struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

type CreateRequest struct {
	Challenge        string            `json:"challenge"`
	Rp               Rp                `json:"rp"`
	User             User              `json:"user"`
	Timeout          int               `json:"timeout"`
	Attestation      AttestationType   `json:"attestation"`
	PubKeyCredParams []PubKeyCredParam `json:"pubKeyCredParams"`
}

type Credential struct {
	ID int
	Opts
}

func generateChallenge() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func NewCreateRequest(id int, display, name string) (*CreateRequest, error) {
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
			ID:          base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(id))),
			DisplayName: display,
			Name:        name,
		},
		Timeout:     60000,
		Attestation: AttestationNone,
		PubKeyCredParams: []PubKeyCredParam{
			{Kind: "public-key", Alg: -7},
		},
	}

	return createRequest, nil
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

func GetAll(ctx context.Context) ([]Credential, error) {
	credentials := make([]Credential, 0)

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
		credentials = append(credentials, crd)
	}

	if err := rows.Err(); err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	return credentials, nil
}
