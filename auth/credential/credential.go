package credential

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func NewPubKeyCreateRequest(userID, displayName string) (*PubKeyCreateRequest, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}

	request := &PubKeyCreateRequest{
		kind: "webauthn.create",
		PubKey: &pubKeyCredentialCreationOpts{
			Challenge: challenge,
			Rp: pubKeyCredentialRpEntity{
				Name: config.Config.RpName,
				ID:   config.Config.RpID,
			},
			User: pubKeyCredentialUserEntity{
				id:          userID,
				EncodedID:   base64.RawURLEncoding.EncodeToString([]byte(userID)),
				DisplayName: displayName,
			},
			Timeout: 60000,
			AuthenticatorSelection: authenticatorSelectionCriteria{
				UserVerification: UserVerificationRequired,
			},
			Attestation: AttestationNone,
			PubKeyCredParams: []pubKeyCredentialParameters{
				{Type: "public-key", Alg: COSEAlgES256},
			},
		},
	}

	return request, nil
}

func NewPubKeyGetRequest(userID string, allowCredentials []pubKeyCredentialDescriptor) (*pubKeyGetRequest, error) {
	challenge, err := generateChallenge()

	if err != nil {
		return nil, err
	}

	request := &pubKeyGetRequest{
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
		"select id, credential_id, title, transports, user_id from credentials where credential_id = $1", credId).Scan(
		&credential.ID, &credential.Title, &credential.Transports, &credential.UserID)

	if err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	return &credential, nil
}

func GetAll(ctx context.Context, userID string) ([]pubKeyCredentialDescriptor, error) {
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
	err := db.Pool.QueryRow(ctx,
		"insert into credential_requests (challenge, type, user_id) values ($1, $2, $3) returning id",
		r.PubKey.Challenge, r.kind, r.PubKey.User.id,
	).Scan(&r.ID)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return errors.Internal
	}

	return nil
}

func (r *pubKeyGetRequest) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credential_requests (challenge, type, user_id) values ($1, $2, $3) returning id",
		r.PubKey.Challenge, r.kind, r.userID,
	).Scan(&r.ID)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func (k *PubKeyCredential) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		`insert into credentials (credential_id, title, pubkey, pubkey_algo, transports, user_id, counter, credential_request_id)
		 values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`,
		k.CredentialID, k.Title, k.PubKey, k.PubKeyAlg, k.Transports, k.UserID, k.Counter, k.CredentialRequestID,
	).Scan(&k.ID)

	if err != nil {
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

func ValidateClientData(rawClientData string, req *Request) error {
	decoded, err := base64.RawURLEncoding.DecodeString(rawClientData)
	if err != nil {
		return errors.Internal
	}

	clientData := &collectedClientData{}

	if err := json.Unmarshal(decoded, &clientData); err != nil {
		return errors.Internal
	}

	if clientData.Challenge != req.Challenge {
		return errors.NewBadRequest(l.T("Challenges don't match"))
	}

	if clientData.Origin != config.Config.Origin {
		return errors.NewBadRequest(l.T("Invalid origin"))
	}

	if clientData.Type != req.Type {
		return errors.NewBadRequest(l.T("Invalid credential type"))
	}

	return nil
}

func generateChallenge() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func parsePubKey(pubKey string) (string, error) {
	keyBytes, err := base64.RawURLEncoding.DecodeString(pubKey)
	if err != nil {
		return "", err
	}
	_ = keyBytes

	return "", nil
}

func (k *ec2PubKeyData) verify(data []byte, sig []byte) (bool, error) {
	var curve elliptic.Curve
	switch COSEAlgorithmIdentifier(k.Alg) {
	case COSEAlgES256:
		curve = elliptic.P256()
	case COSEAlgES384:
		curve = elliptic.P384()
	case COSEAlgES512:
		curve = elliptic.P521()
	default:
		return false, fmt.Errorf("Unsupported")
	}

	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     big.NewInt(0).SetBytes(k.xcoord),
		Y:     big.NewInt(0).SetBytes(k.ycoord),
	}
	_ = pubKey

	return false, nil
}

func parseAuthenticatorData(authnData string) (*authenticatorData, error) {
	data, err := base64.RawURLEncoding.DecodeString(authnData)
	if err != nil {
		return nil, errors.Internal
	}

	if len(data) < 37 {
		return nil, errors.NewBadRequest("authenticator data: unexpected EOF")
	}

	r := &authenticatorData{}
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
