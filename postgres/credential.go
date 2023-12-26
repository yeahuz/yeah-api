package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type CredentialService struct {
	pool   *pgxpool.Pool
	rpName string
	rpID   string
	origin string
}

func NewCredentailService(pool *pgxpool.Pool) *CredentialService {
	return &CredentialService{
		pool: pool,
	}
}

var hashers = map[yeahapi.COSEAlgorithmIdentifier]func() hash.Hash{
	yeahapi.COSEAlgES256: sha256.New,
	yeahapi.COSEAlgEdDSA: sha512.New,
	yeahapi.COSEAlgRS256: sha256.New,
}

func (c *CredentialService) CreatePubKeyRequest(ctx context.Context, user *yeahapi.User) (*yeahapi.PubKeyCreateRequest, error) {
	const op yeahapi.Op = "credential.CreatePubKeyRequest"
	challenge, err := generateChallenge()
	if err != nil {
		return nil, yeahapi.E(op, err, "unable to generate a challenge")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err, "unable to generate uuid")
	}

	request := &yeahapi.PubKeyCreateRequest{
		ID:   id,
		Kind: "webauthn.create",
		PubKey: &yeahapi.PubKeyCredentialCreationOpts{
			Challenge: challenge,
			Rp: yeahapi.PubKeyCredentialRpEntity{
				Name: c.rpName,
				ID:   c.rpID,
			},
			User: yeahapi.PubKeyCredentialUserEntity{
				ID:          user.ID,
				EncodedID:   base64.RawURLEncoding.EncodeToString([]byte(user.ID)),
				DisplayName: user.FirstName,
			},
			Timeout: 60000,
			AuthenticatorSelection: yeahapi.AuthenticatorSelectionCriteria{
				UserVerification: yeahapi.UserVerificationRequired,
			},
			Attestation: yeahapi.AttestationNone,
			PubKeyCredParams: []yeahapi.PubKeyCredentialParameters{
				{Type: "public-key", Alg: yeahapi.COSEAlgES256},
				{Type: "public-key", Alg: yeahapi.COSEAlgEdDSA},
				{Type: "public-key", Alg: yeahapi.COSEAlgRS256},
			},
		},
	}

	_, err = c.pool.Exec(ctx,
		"insert into credential_requests (id, challenge, type, user_id) values ($1, $2, $3, $4)",
		request.ID, request.PubKey.Challenge, request.Kind, request.PubKey.User.ID,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return request, nil
}

func (c *CredentialService) GetPubKeyRequest(ctx context.Context, userID yeahapi.UserID) (*yeahapi.PubKeyGetRequest, error) {
	const op yeahapi.Op = "credential.GetPubKeyRequest"
	credentials, err := c.Credentials(ctx, userID)
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	challenge, err := generateChallenge()
	if err != nil {
		return nil, yeahapi.E(op, err, "unable to generate a challenge")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err, "unable to generate uuid")
	}

	request := &yeahapi.PubKeyGetRequest{
		ID:   id,
		Kind: "webauthn.get",
		PubKey: &yeahapi.PubKeyCredentialRequestOpts{
			Challenge:        challenge,
			RpID:             c.rpID,
			Timeout:          60000,
			AllowCredentials: credentials,
			UserVerification: yeahapi.UserVerificationRequired,
		},
		UserID: userID,
	}

	_, err = c.pool.Exec(ctx,
		"insert into credential_requests (id, challenge, type, user_id) values ($1, $2, $3, $4)",
		request.ID, request.PubKey.Challenge, request.Kind, request.UserID,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return request, nil
}

func (c *CredentialService) Credentials(ctx context.Context, userID yeahapi.UserID) ([]yeahapi.PubKeyCredentialDescriptor, error) {
	const op yeahapi.Op = "credential.credentials"
	credentials := make([]yeahapi.PubKeyCredentialDescriptor, 0)

	rows, err := c.pool.Query(ctx, "select credential_id, transports, type from credentials where user_id = $1", userID)
	defer rows.Close()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	for rows.Next() {
		var crd yeahapi.PubKeyCredentialDescriptor
		if err := rows.Scan(&crd.ID, &crd.Transports, &crd.Type); err != nil {
			return nil, yeahapi.E(op, err)
		}

		credentials = append(credentials, crd)
	}

	if err := rows.Err(); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return credentials, nil
}

func (c *CredentialService) Credential(ctx context.Context, id string) (*yeahapi.PubKeyCredential, error) {
	const op yeahapi.Op = "credential.credential"
	var credential yeahapi.PubKeyCredential
	err := c.pool.QueryRow(ctx,
		"select id, credential_id, title, transports, user_id, pubkey, pubkey_alg from credentials where credential_id = $1", id).Scan(
		&credential.ID, &credential.CredentialID, &credential.Title, &credential.Transports, &credential.UserID, &credential.PubKey, &credential.PubKeyAlg)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return &credential, nil
}

func (c *CredentialService) ValidateAuthnData(data string) (*yeahapi.AuthenticatorData, error) {
	const op yeahapi.Op = "credential.ValidateAuthnData"
	authnData, err := parseAuthenticatorData(data)
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	if !authnData.UserPresent {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "User was not present during authentication")
	}

	if !authnData.UserVerified {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "User not verified during authenication")
	}

	return authnData, err
}

func (c *CredentialService) ValidateClientData(data string, req *yeahapi.CredentialRequest) (*yeahapi.CollectedClientData, error) {
	const op yeahapi.Op = "credential.ValidateClientData"
	decoded, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, yeahapi.E(err, yeahapi.EInternal)
	}

	clientData := &yeahapi.CollectedClientData{Raw: decoded}

	if err := json.Unmarshal(decoded, &clientData); err != nil {
		return nil, yeahapi.E(op, err)
	}

	if clientData.Challenge != req.Challenge {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "Challenges don't match")
	}

	if clientData.Origin != c.origin {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "Invalid origin")
	}

	if clientData.Type != req.Type {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "Invalid credential type")
	}

	return clientData, nil
}

func generateChallenge() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func parseAuthenticatorData(authnData string) (*yeahapi.AuthenticatorData, error) {
	const op yeahapi.Op = "credential.parseAuthenticatorData"
	data, err := base64.RawURLEncoding.DecodeString(authnData)
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	if len(data) < 37 {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "unexpected EOF of authn data")
	}

	r := &yeahapi.AuthenticatorData{Raw: data}
	r.RpIDHash = make([]byte, 32)
	copy(r.RpIDHash, data)

	flags := data[32]
	r.UserPresent = (flags & 0x01) > 0
	r.UserVerified = (flags & 0x04) > 0
	credentialIncluded := (flags & 0x40) > 0
	extensionDataIncluded := (flags & 0x80) > 0

	r.Counter = binary.BigEndian.Uint32(data[33:37])

	rest := data[37:]

	if credentialIncluded {
		if len(rest) < 18 {
			return nil, yeahapi.E(op, yeahapi.EInvalid, "unexpected EOF of credential")
		}

		r.AAGUID = make([]byte, 16)
		copy(r.AAGUID, rest)

		idlen := binary.BigEndian.Uint16(rest[16:18])

		if len(rest[18:]) < int(idlen) {
			return nil, yeahapi.E(op, yeahapi.EInvalid, "unexpected EOF of credential")
		}

		r.CredentialID = make([]byte, idlen)

		copy(r.CredentialID, rest[18:])
	}

	if extensionDataIncluded {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "Unexpected credential extension")
	}

	return r, nil
}
