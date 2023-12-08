package credential

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func NewPubKeyCreateRequest(userId, displayName string) (*PubKeyCreateRequest, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}

	request := &PubKeyCreateRequest{
		kind: "webauthn.create",
		PubKey: &PublicKeyCredentialCreationOpts{
			Challenge: challenge,
			Rp: PublicKeyCredentialRpEntity{
				Name: config.Config.RpName,
				ID:   config.Config.RpID,
			},
			User: PublicKeyCredentialUserEntity{
				ID:          []byte(userId),
				DisplayName: displayName,
			},
			Timeout: 60000,
			AuthenticatorSelection: AuthenticatorSelectionCriteria{
				UserVerification: UserVerificationRequired,
			},
			Attestation: AttestationNone,
			PubKeyCredParams: []PublicKeyCredentialParameters{
				{Type: "public-key", Alg: COSEAlgorithmES256},
			},
		},
	}

	return request, nil
}

func NewPubKeyGetRequest(allowCredentials []PublicKeyCredentialDescriptor) (*PubKeyGetRequest, error) {
	challenge, err := generateChallenge()

	if err != nil {
		return nil, err
	}

	request := &PubKeyGetRequest{
		kind: "webauthn.get",
		PubKey: &PublicKeyCredentialRequestOpts{
			Challenge:        challenge,
			RpID:             config.Config.RpID,
			Timeout:          60000,
			AllowCredentials: allowCredentials,
			UserVerification: UserVerificationRequired,
		},
	}

	return request, nil
}

// func New(opts Opts) *Credential {
// 	return &Credential{
// 		Opts: opts,
// 	}
// }

// func (c *Credential) Save(ctx context.Context) error {
// 	err := db.Pool.QueryRow(ctx,
// 		"insert into credentials (title, credential_id, pubkey, counter, user_id, transports) values ($1, $2, $3, $4, $5, $6) returning id",
// 		c.Title, c.CredentialID, c.PubKey, c.Counter, c.UserID, c.Transports,
// 	).Scan(&c.ID)

// 	if err != nil {
// 		// TODO: handle errors properly
// 		return errors.Internal
// 	}
// 	return nil
// }

// func GetById(ctx context.Context, credId string) (*Credential, error) {
// 	var credential Credential
// 	err := db.Pool.QueryRow(ctx,
// 		"select id, credential_id, title, transports, user_id from credentials where credential_id = $1", credId).Scan(
// 		&credential.ID, &credential.Title, &credential.Transports, &credential.UserID)

// 	if err != nil {
// 		// TODO: handle errors properly
// 		return nil, errors.Internal
// 	}

// 	return &credential, nil
// }

// func GetAllowedCredentials(ctx context.Context, userID string) ([]allowedCredential, error) {
// 	credentials := make([]allowedCredential, 0)
// 	rows, err := db.Pool.Query(ctx, "select credential_id, transports type from credentials where user_id = $1", userID)
// 	defer rows.Close()
// 	if err != nil {
// 		return nil, errors.Internal
// 	}

// 	for rows.Next() {
// 		var crd allowedCredential
// 		if err := rows.Scan(&crd.ID, &crd.Transports, &crd.Type); err != nil {
// 			return nil, errors.Internal
// 		}

// 		credentials = append(credentials, crd)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, errors.Internal
// 	}

// 	return credentials, nil
// }

// func GetAll(ctx context.Context) (*Credentials, error) {
// 	creds := make([]Credential, 0)

// 	rows, err := db.Pool.Query(ctx, "select id, credential_id, title, transports, user_id from credentials order by id desc")

// 	defer rows.Close()
// 	if err != nil {
// 		// TODO: handle errors properly
// 		return nil, errors.Internal
// 	}

// 	for rows.Next() {
// 		var crd Credential
// 		if err := rows.Scan(&crd.ID, &crd.CredentialID, &crd.Title, &crd.Transports, &crd.UserID); err != nil {
// 			// TODO: handle errors properly
// 			return nil, errors.Internal
// 		}
// 		creds = append(creds, crd)
// 	}

// 	if err := rows.Err(); err != nil {
// 		// TODO: handle errors properly
// 		return nil, errors.Internal
// 	}

// 	list := &Credentials{Credentials: creds, Count: len(creds)}

// 	return list, nil
// }

func (r *PubKeyCreateRequest) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credential_requests (challenge, type) values ($1, $2) returning id",
		r.PubKey.Challenge.String(), r.kind,
	).Scan(&r.ID)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return errors.Internal
	}

	return nil
}

func (r *PubKeyGetRequest) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credential_requests (challenge, type) values ($1, $2) returning id",
		string(r.PubKey.Challenge), r.kind,
	).Scan(&r.ID)

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

// func (r Request) VerifyClientData(data string) (*parsedClientData, error) {
// 	decoded, err := base64.RawURLEncoding.DecodeString(data)
// 	if err != nil {
// 		return nil, errors.Internal
// 	}

// 	var clientData parsedClientData

// 	if err := json.NewDecoder(bytes.NewReader(decoded)).Decode(&clientData); err != nil {
// 		return nil, errors.Internal
// 	}

// 	if clientData.Type != r.Type {
// 		return nil, errors.NewBadRequest(l.T("Invalid credential type"))
// 	}

// 	decodedChallenge, err := base64.RawURLEncoding.DecodeString(clientData.Challenge)
// 	if err != nil {
// 		return nil, errors.Internal
// 	}

// 	if string(decodedChallenge) != r.Challenge {
// 		return nil, errors.NewBadRequest(l.T("Challenges don't match"))
// 	}

// 	if clientData.Origin != config.Config.Origin {
// 		return nil, errors.NewBadRequest(l.T("Origin is invalid"))
// 	}

// 	return &clientData, nil
// }

// func ParseAttestation(b []byte) (*parsedAttestationObject, error) {
// 	var p parsedAttestationObject

// 	if err := cbor.Unmarshal(b, &p); err != nil {
// 		return nil, err
// 	}

// 	return &p, nil
// }

func generateChallenge() ([]byte, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
