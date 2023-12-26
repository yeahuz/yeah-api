package yeahapi

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"math/big"

	"github.com/gofrs/uuid"
)

type PubKeyCredentialOpts struct {
	CredentialID        string
	Title               string
	PubKey              string
	PubKeyAlg           int
	Transports          []AuthenticatorTransport
	UserID              uuid.UUID
	Counter             uint32
	CredentialRequestID string
}

type PubKeyCredential struct {
	ID                  uuid.UUID
	CredentialID        string
	Title               string
	PubKey              string
	PubKeyAlg           int
	Transports          []AuthenticatorTransport
	UserID              uuid.UUID
	Counter             uint32
	CredentialRequestID uuid.UUID
}

type CredentialRequest struct {
	ID        string
	Type      string
	Challenge string
	Used      bool
	UserID    uuid.UUID
}

type AuthenticatorTransport string

const (
	AuthenticatorUSB      AuthenticatorTransport = "usb"
	AuthenticatorNFC      AuthenticatorTransport = "nfc"
	AuthenticatorBLE      AuthenticatorTransport = "ble"
	AuthenticatorInternal AuthenticatorTransport = "internal"
)

type AuthenticatorAttachment string

const (
	AuthenticatorPlatform      AuthenticatorAttachment = "platform"
	AuthenticatorCrossPlatform AuthenticatorAttachment = "cross-platform"
)

type ResidentKeyRequirement string

const (
	ResidentKeyDiscouraged ResidentKeyRequirement = "discouraged"
	ResidentKeyPreferred   ResidentKeyRequirement = "preferred"
	ResidentKeyRequired    ResidentKeyRequirement = "required"
)

type UserVerificationRequirement string

const (
	UserVerificationRequired    UserVerificationRequirement = "required"
	UserVerificationPreferred   UserVerificationRequirement = "preferred"
	UserVerificationDiscouraged UserVerificationRequirement = "discouraged"
)

type AttestationConveyancePreference string

const (
	AttestationNone       AttestationConveyancePreference = "none"
	AttestationIndirect   AttestationConveyancePreference = "indirect"
	AttestationDirect     AttestationConveyancePreference = "direct"
	AttestationEnterprise AttestationConveyancePreference = "enterprise"
)

type COSEAlgorithmIdentifier int

const (
	COSEAlgES256 COSEAlgorithmIdentifier = -7
	COSEAlgEdDSA COSEAlgorithmIdentifier = -8
	COSEAlgRS256 COSEAlgorithmIdentifier = -257
)

type PubKeyCreateRequest struct {
	ID     uuid.UUID                     `json:"id"`
	PubKey *PubKeyCredentialCreationOpts `json:"pubkey"`
	Kind   string
}

type PubKeyGetRequest struct {
	ID     uuid.UUID                    `json:"id"`
	PubKey *PubKeyCredentialRequestOpts `json:"pubkey"`
	UserID UserID
	Kind   string
}

type CollectedClientData struct {
	Raw       []byte
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

type authenticatorResponse struct {
	ClientDataJSON string `json:"client_data_json"`
}

type AuthenticatorData struct {
	Raw          []byte
	RpIDHash     []byte
	UserPresent  bool
	UserVerified bool
	Counter      uint32
	AAGUID       []byte
	CredentialID []byte
}

type AuthenticatorAttestationResponse struct {
	authenticatorResponse
	AuthenticatorData string                   `json:"authenticator_data"`
	Transports        []AuthenticatorTransport `json:"transports"`
	PubKey            string                   `json:"pubkey"`
	PubKeyAlg         int                      `json:"pubkey_alg"`
}

type AuthenticatorAssertionResponse struct {
	authenticatorResponse
	AuthenticatorData string `json:"authenticator_data"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"user_handle"`
}

type rawPubKeyCredential struct {
	ID    string `json:"id"`
	RawID string `json:"raw_id"`
}

type RawPubKeyCredentialAttestation struct {
	rawPubKeyCredential
	Response AuthenticatorAttestationResponse `json:"response"`
}

type RawPubKeyCredentialAssertion struct {
	rawPubKeyCredential
	Response AuthenticatorAssertionResponse `json:"response"`
}

type CreatePubKeyData struct {
	ReqID      string                         `json:"req_id"`
	Credential RawPubKeyCredentialAttestation `json:"credential"`
	Title      string
}

type AssertPubKeyData struct {
	ReqID      string                       `json:"req_id"`
	Credential RawPubKeyCredentialAssertion `json:"credential"`
}

type PubKeyCredentialRpEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PubKeyCredentialUserEntity struct {
	ID          UserID
	EncodedID   string `json:"id"`
	DisplayName string `json:"display_name"`
}

type AuthenticatorSelectionCriteria struct {
	AuthenticatorAttachment AuthenticatorAttachment     `json:"authenticator_attachment,omitempty"`
	ResidentKey             ResidentKeyRequirement      `json:"resident_key,omitempty"`
	RequireResidentKey      bool                        `json:"require_resident_key,omitempty"`
	UserVerification        UserVerificationRequirement `json:"user_verification,omitempty"`
}

type PubKeyCredentialDescriptor struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Transports []AuthenticatorTransport `json:"transports"`
}

type PubKeyCredentialParameters struct {
	Type string                  `json:"type"`
	Alg  COSEAlgorithmIdentifier `json:"alg"`
}

type PubKeyCredentialCreationOpts struct {
	Rp                     PubKeyCredentialRpEntity        `json:"rp"`
	User                   PubKeyCredentialUserEntity      `json:"user"`
	Challenge              string                          `json:"challenge"`
	PubKeyCredParams       []PubKeyCredentialParameters    `json:"pubkey_cred_params"`
	Timeout                int                             `json:"timeout"`
	ExcludeCredentials     []PubKeyCredentialDescriptor    `json:"exclude_credentials,omitempty"`
	AuthenticatorSelection AuthenticatorSelectionCriteria  `json:"authenticator_selection,omitempty"`
	Attestation            AttestationConveyancePreference `json:"attestation"`
}

type PubKeyCredentialRequestOpts struct {
	Challenge        string                       `json:"challenge"`
	Timeout          int                          `json:"timeout"`
	RpID             string                       `json:"rp_id"`
	AllowCredentials []PubKeyCredentialDescriptor `json:"allow_credentials"`
	UserVerification UserVerificationRequirement  `json:"user_verification"`
}

func (c *PubKeyCredential) Verify(clientData []byte, authnData []byte, sig string) error {
	sigbytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return errors.Internal
	}

	clientDataHash := sha256.Sum256(clientData)
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

type CredentialService interface {
	PubKeyCreateRequest(ctx context.Context, user *User) (*PubKeyCreateRequest, error)
	PubKeyGetRequest(ctx context.Context, userID UserID) (*PubKeyGetRequest, error)
	CreatePubKey(ctx context.Context, credential *PubKeyCredential) error
	VerifyPubKey(ctx context.Context) error
	Request(ctx context.Context, id uuid.UUID) (*CredentialRequest, error)
	Credentials(ctx context.Context, userID UserID) ([]PubKeyCredentialDescriptor, error)
	Credential(ctx context.Context, id string) (*PubKeyCredential, error)
	ValidateClientData(data string, req *CredentialRequest) (*CollectedClientData, error)
	ValidateAuthnData(data string) (*AuthenticatorData, error)
}
