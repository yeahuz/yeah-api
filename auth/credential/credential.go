package credential

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash"
	"math/big"

	"github.com/gofrs/uuid"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

var hashers = map[COSEAlgorithmIdentifier]func() hash.Hash{
	COSEAlgES256: sha256.New,
	COSEAlgEdDSA: sha512.New,
	COSEAlgRS256: sha256.New,
}

func NewPubKeyCredential(opts *PubKeyCredentialOpts) (*PubKeyCredential, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &PubKeyCredential{ID: id, PubKeyCredentialOpts: opts}, nil
}

func NewPubKeyCreateRequest(userID uuid.UUID, displayName string) (*PubKeyCreateRequest, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	request := &PubKeyCreateRequest{
		ID:   id,
		kind: "webauthn.create",
		PubKey: &pubKeyCredentialCreationOpts{
			Challenge: challenge,
			Rp: pubKeyCredentialRpEntity{
				Name: config.Config.RpName,
				ID:   config.Config.RpID,
			},
			User: pubKeyCredentialUserEntity{
				id:          userID,
				EncodedID:   base64.RawURLEncoding.EncodeToString(userID[:]),
				DisplayName: displayName,
			},
			Timeout: 60000,
			AuthenticatorSelection: authenticatorSelectionCriteria{
				UserVerification: UserVerificationRequired,
			},
			Attestation: AttestationNone,
			PubKeyCredParams: []pubKeyCredentialParameters{
				{Type: "public-key", Alg: COSEAlgES256},
				{Type: "public-key", Alg: COSEAlgEdDSA},
				{Type: "public-key", Alg: COSEAlgRS256},
			},
		},
	}

	return request, nil
}

func NewPubKeyGetRequest(userID uuid.UUID, allowCredentials []pubKeyCredentialDescriptor) (*pubKeyGetRequest, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	request := &pubKeyGetRequest{
		ID:   id,
		kind: "webauthn.get",
		PubKey: &pubKeyCredentialRequestOpts{
			Challenge:        challenge,
			RpID:             config.Config.RpID,
			Timeout:          60000,
			AllowCredentials: allowCredentials,
			UserVerification: UserVerificationRequired,
		},
		userID: userID,
	}

	return request, nil
}

func GetById(ctx context.Context, credId string) (*PubKeyCredential, error) {
	var credential PubKeyCredential
	err := db.Pool.QueryRow(ctx,
		"select id, credential_id, title, transports, user_id, pubkey, pubkey_alg from credentials where credential_id = $1", credId).Scan(
		&credential.ID, &credential.CredentialID, &credential.Title, &credential.Transports, &credential.UserID, &credential.PubKey, &credential.PubKeyAlg)

	if err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	return &credential, nil
}

func GetAll(ctx context.Context, userID uuid.UUID) ([]pubKeyCredentialDescriptor, error) {
	credentials := make([]pubKeyCredentialDescriptor, 0)
	rows, err := db.Pool.Query(ctx, "select credential_id, transports, type from credentials where user_id = $1", userID)
	defer rows.Close()
	if err != nil {
		return nil, errors.Internal
	}

	for rows.Next() {
		var crd pubKeyCredentialDescriptor
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

func (r *PubKeyCreateRequest) Save(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		"insert into credential_requests (id, challenge, type, user_id) values ($1, $2, $3, $4) returning id",
		r.ID, r.PubKey.Challenge, r.kind, r.PubKey.User.id,
	)

	if err != nil {
		//TODO: handle errors properly
		return errors.Internal
	}

	return nil
}

func (r *pubKeyGetRequest) Save(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		"insert into credential_requests (id, challenge, type, user_id) values ($1, $2, $3, $4) returning id",
		r.ID, r.PubKey.Challenge, r.kind, r.userID,
	)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func (k *PubKeyCredential) Save(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		`insert into credentials (id, credential_id, title, pubkey, pubkey_alg, transports, user_id, counter, credential_request_id)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`,
		k.ID, k.CredentialID, k.Title, k.PubKey, k.PubKeyAlg, k.Transports, k.UserID, k.Counter, k.CredentialRequestID,
	)

	if err != nil {
		// TODO: handle errors properly
		return errors.Internal
	}

	return nil
}

func GetRequestById(ctx context.Context, id string) (*Request, error) {
	var request Request

	err := db.Pool.QueryRow(ctx,
		"select type, challenge, used, user_id from credential_requests where id = $1",
		id,
	).Scan(&request.Type, &request.Challenge, &request.Used, &request.UserID)

	if err != nil {
		return nil, errors.Internal
	}

	return &request, nil
}

func ValidateAuthenticatorData(data string) (*authenticatorData, error) {
	authnData, err := parseAuthenticatorData(data)
	if err != nil {
		return nil, err
	}

	if !authnData.UserPresent {
		return nil, errors.NewBadRequest(l.T("User was not present during authenication"))
	}

	if !authnData.UserVerified {
		return nil, errors.NewBadRequest(l.T("User not verified during authenication"))
	}

	return authnData, err
}

func ValidateClientData(rawClientData string, req *Request) (*collectedClientData, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(rawClientData)
	if err != nil {
		return nil, errors.Internal
	}

	clientData := &collectedClientData{Raw: decoded}

	if err := json.Unmarshal(decoded, &clientData); err != nil {
		return nil, errors.Internal
	}

	if clientData.Challenge != req.Challenge {
		return nil, errors.NewBadRequest(l.T("Challenges don't match"))
	}

	if clientData.Origin != config.Config.Origin {
		return nil, errors.NewBadRequest(l.T("Invalid origin"))
	}

	if clientData.Type != req.Type {
		return nil, errors.NewBadRequest(l.T("Invalid credential type"))
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

func (c *PubKeyCredential) Verify(cdata []byte, authnData []byte, sig string) error {
	sigbytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return errors.Internal
	}

	clientDataHash := sha256.Sum256(cdata)
	message := make([]byte, len(authnData)+len(clientDataHash))
	copy(message, authnData)
	copy(message[len(authnData):], clientDataHash[:])

	return c.verifySignature(message, sigbytes)
}

func (c *PubKeyCredential) verifySignature(message []byte, sig []byte) error {
	bytes, err := base64.RawURLEncoding.DecodeString(c.PubKey)
	if err != nil {
		return errors.NewInternal(l.T("Couldn't decode pubkey"))
	}

	parsed, err := x509.ParsePKIXPublicKey(bytes)

	if err != nil {
		return errors.NewInternal(l.T("Unable to parse pubkey"))
	}

	hasher := hashers[COSEAlgorithmIdentifier(c.PubKeyAlg)]
	if hasher == nil {
		return errors.NewInternal(l.T("Unsupported hashing algorithm"))
	}

	h := hasher()
	_, err = h.Write(message)
	if err != nil {
		return errors.NewInternal(l.T("Couldn't hash the data"))
	}

	digest := h.Sum(nil)

	switch pk := parsed.(type) {
	case *ecdsa.PublicKey:
		type ecdsaSignature struct {
			R, S *big.Int
		}
		var ecdsaSig ecdsaSignature
		if rest, err := asn1.Unmarshal(sig, &ecdsaSig); err != nil {
			return errors.Internal
		} else if len(rest) != 0 {
			return errors.NewBadRequest(l.T("Trailing data after ECDSA signature"))
		}
		if ecdsaSig.R.Sign() <= 0 || ecdsaSig.S.Sign() <= 0 {
			return errors.NewBadRequest(l.T("ECDSA signature contained zero or negative values"))
		}
		if !ecdsa.Verify(pk, digest, ecdsaSig.R, ecdsaSig.S) {
			return errors.NewBadRequest(l.T("ECDSA signature verification failed"))
		}
	case *rsa.PublicKey:
		if err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, digest, sig); err != nil {
			return errors.NewBadRequest(l.T("RSA signature verification failed"))
		}
	default:
		return errors.NewInternal("Unsupported key type")

	}

	return nil
}

func parseAuthenticatorData(authnData string) (*authenticatorData, error) {
	data, err := base64.RawURLEncoding.DecodeString(authnData)
	if err != nil {
		return nil, errors.Internal
	}

	if len(data) < 37 {
		return nil, errors.NewBadRequest("authenticator data: unexpected EOF")
	}

	r := &authenticatorData{Raw: data}
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
			return nil, errors.NewBadRequest(l.T("Unexpected EOF of credential"))
		}

		r.AAGUID = make([]byte, 16)
		copy(r.AAGUID, rest)

		idlen := binary.BigEndian.Uint16(rest[16:18])

		if len(rest[18:]) < int(idlen) {
			return nil, errors.NewBadRequest(l.T("Unexpected EOF of credential"))
		}

		r.CredentialID = make([]byte, idlen)

		copy(r.CredentialID, rest[18:])
	}

	if extensionDataIncluded {
		return nil, errors.NewBadRequest(l.T("Unexpected credential extension"))
	}

	return r, nil
}
