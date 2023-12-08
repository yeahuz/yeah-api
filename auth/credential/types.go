package credential

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
)

// type authFlag byte

// // https://github.com/go-webauthn/webauthn/blob/master/protocol/authenticator.go#L175
// const (
// 	// FlagUserPresent Bit 00000001 in the byte sequence. Tells us if user is present. Also referred to as the UP flag.
// 	FlagUserPresent authFlag = 1 << iota

// 	// FlagRFU1 is a reserved for future use flag.
// 	FlagRFU1
// 	// FlagUserVerified Bit 00000100 in the byte sequence. Tells us if user is verified
// 	// by the authenticator using a biometric or PIN. Also referred to as the UV flag.
// 	FlagUserVerified

// 	// FlagBackupEligible Bit 00001000 in the byte sequence. Tells us if a backup is eligible for device. Also referred
// 	// to as the BE flag.
// 	FlagBackupEligible // Referred to as BE

// 	// FlagBackupState Bit 00010000 in the byte sequence. Tells us if a backup state for device. Also referred to as the
// 	// BS flag.
// 	FlagBackupState

// 	// FlagRFU2 is a reserved for future use flag.
// 	FlagRFU2

// 	// FlagAttestedCredentialData Bit 01000000 in the byte sequence. Indicates whether
// 	// the authenticator added attested credential data. Also referred to as the AT flag.
// 	FlagAttestedCredentialData

// 	// FlagHasExtensions Bit 10000000 in the byte sequence. Indicates if the authenticator data has extensions. Also
// 	// referred to as the ED flag.
// 	FlagHasExtensions
// )

// type urlEncodedB64 []byte

// type PubKeyCredParam struct {
// 	Alg  int            `json:"alg"`
// 	Type PubKeyCredType `json:"type"`
// }

// type Opts struct {
// 	CredentialID string
// 	Title        string
// 	PubKey       string
// 	Counter      int
// 	UserID       string
// 	Transports   []string
// }

// type Rp struct {
// 	Name string `json:"name"`
// 	ID   string `json:"id"`
// }

// type User struct {
// 	ID          string `json:"id"`
// 	DisplayName string `json:"display_name"`
// 	Name        string `json:"name"`
// }

type Request struct {
	Type      string
	Challenge string
	Used      bool
}

// type CreateRequest struct {
// 	ID               string            `json:"id"`
// 	Challenge        string            `json:"challenge"`
// 	Rp               Rp                `json:"rp"`
// 	User             User              `json:"user"`
// 	Timeout          int               `json:"timeout"`
// 	Attestation      AttestationType   `json:"attestation"`
// 	PubKeyCredParams []PubKeyCredParam `json:"pubkey_cred_params"`
// }

// type allowedCredential struct {
// 	ID         string   `json:"id"`
// 	Type       string   `json:"type"`
// 	Transports []string `json:"transports"`
// }

// type GetRequest struct {
// 	ID               string              `json:"id"`
// 	Challenge        string              `json:"challenge"`
// 	RpID             string              `json:"rpId"`
// 	Timeout          int                 `json:"timeout"`
// 	AllowCredentials []allowedCredential `json:"allow_credentials"`
// }

// type Credential struct {
// 	ID string
// 	Opts
// }

// type parsedClientData struct {
// 	Challenge string `json:"challenge"`
// 	Type      string `json:"type"`
// 	Origin    string `json:"origin"`
// }

// type attestedCredentialData struct {
// 	aaguid           []byte
// 	credentialID     []byte
// 	credentialPubkey []byte
// }

// type authenticatorData struct {
// 	RpIDHash     []byte
// 	UserPresent  bool
// 	UserVerified bool
// 	Counter      uint32
// 	AAGUID       []byte
// 	CredentialID []byte
// }

// type parsedAttestationObject struct {
// 	AuthnData []byte          `cbor:"authData"`
// 	Fmt       string          `cbor:"fmt"`
// 	AttStmt   cbor.RawMessage `cbor:"attStmt"`
// }

// type AttestationResponse struct {
// 	ClientDataJSON    string `json:"client_data_json"`
// 	AttestationObject string `json:"attestation_object"`
// }

// type parsedAttestationResponse struct {
// 	clientData        parsedClientData
// 	attestationObject parsedAttestationObject
// }

// type CreateResponse struct {
// 	ID         string               `json:"id"`
// 	RawID      string               `json:"raw_id"`
// 	Response   *AttestationResponse `json:"response"`
// 	Type       PubKeyCredType       `json:"type"`
// 	Transports []string             `json:"transports"`
// }

// type CreateCredentialData struct {
// 	ReqID      string          `json:"req_id"`
// 	Credential *CreateResponse `json:"credential"`
// 	Title      string          `json:"title"`
// }

// type Credentials struct {
// 	Credentials []Credential `json:"credentials"`
// 	Count       int          `json:"count"`
// }

// func (gr GetRequest) MarshalJSON() ([]byte, error) {
// 	type Alias GetRequest
// 	return json.Marshal(struct {
// 		Type string `json:"_"`
// 		Alias
// 	}{
// 		Type:  "auth.credentialGetRequest",
// 		Alias: Alias(gr),
// 	})
// }

// func (cr CreateRequest) MarshalJSON() ([]byte, error) {
// 	type Alias CreateRequest
// 	return json.Marshal(struct {
// 		Type string `json:"_"`
// 		Alias
// 	}{
// 		Type:  "auth.credentialCreateRequest",
// 		Alias: Alias(cr),
// 	})
// }

// func (c Credential) MarshalJSON() ([]byte, error) {
// 	type Alias Credential
// 	return json.Marshal(struct {
// 		Type string `json:"_"`
// 		Alias
// 	}{
// 		Type:  "auth.credential",
// 		Alias: Alias(c),
// 	})
// }

// func (cs Credentials) MarshalJSON() ([]byte, error) {
// 	type Alias Credentials
// 	return json.Marshal(struct {
// 		Type string `json:"_"`
// 		Alias
// 	}{
// 		Type:  "auth.credentials",
// 		Alias: Alias(cs),
// 	})
// }

type AuthenticatorTransport string

const (
	Authenticatorusb      AuthenticatorTransport = "usb"
	Authenticatornfc      AuthenticatorTransport = "nfc"
	Authenticatorble      AuthenticatorTransport = "ble"
	Authenticatorinternal AuthenticatorTransport = "internal"
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
	COSEAlgorithmES256 COSEAlgorithmIdentifier = -7
	COSEAlgorithmES384 COSEAlgorithmIdentifier = -35
	COSEAlgorithmES512 COSEAlgorithmIdentifier = -36
	COSEAlgorithmEdDSA COSEAlgorithmIdentifier = -8
)

type b64URLEncoded []byte

type PubKeyCreateRequest struct {
	ID     string                           `json:"id"`
	PubKey *PublicKeyCredentialCreationOpts `json:"pubkey"`
	kind   string
}

type PubKeyGetRequest struct {
	ID     string
	PubKey *PublicKeyCredentialRequestOpts `json:"pubkey"`
	kind   string
}

type AuthenticatorResponse struct {
	ClientData *CollectedClientData
	// ClientDataJSON b64URLEncoded `json:"client_data_json"`
}

type AuthenticatorData struct {
	RpIDHash     []byte
	UserPresent  bool
	UserVerified bool
	Counter      uint32
	AAGUID       []byte
	CredentialID []byte
}

type AuthenticatorAttestationResponse struct {
	AuthenticatorResponse
	AuthenticatorData *AuthenticatorData       `json:"authenticator_data"`
	Transports        []AuthenticatorTransport `json:"transports"`
	PubKey            b64URLEncoded            `json:"pubkey"`
	PubKeyAlgo        int                      `json:"pubkey_algo"`
}

type AuthenticatorAssertionResponse struct {
	AuthenticatorResponse
	AuthenticatorData *AuthenticatorData `json:"authenticator_data"`
	Signature         b64URLEncoded      `json:"signature"`
	UserHandle        b64URLEncoded      `json:"user_handle"`
}

type PublicKeyCredential struct {
	ID       string                           `json:"id"`
	RawID    b64URLEncoded                    `json:"raw_id"`
	Response AuthenticatorAttestationResponse `json:"response"`
}

type CreatePubKeyData struct {
	ReqID      string              `json:"req_id"`
	Credential PublicKeyCredential `json:"credential"`
}

type AssertPubKeyData struct {
	ReqID     string                         `json:"req_id"`
	Assertion AuthenticatorAssertionResponse `json:"assertion"`
}

type PublicKeyCredentialRpEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublicKeyCredentialUserEntity struct {
	ID          b64URLEncoded `json:"id"`
	DisplayName string        `json:"display_name"`
}

type AuthenticatorSelectionCriteria struct {
	AuthenticatorAttachment AuthenticatorAttachment     `json:"authenticator_attachment,omitempty"`
	ResidentKey             ResidentKeyRequirement      `json:"resident_key,omitempty"`
	RequireResidentKey      bool                        `json:"require_resident_key,omitempty"`
	UserVerification        UserVerificationRequirement `json:"user_verification,omitempty"`
}

type PublicKeyCredentialDescriptor struct {
	Type       string                   `json:"type"`
	ID         b64URLEncoded            `json:"id"`
	Transports []AuthenticatorTransport `json:"transports"`
}

type PublicKeyCredentialParameters struct {
	Type string                  `json:"type"`
	Alg  COSEAlgorithmIdentifier `json:"alg"`
}

type PublicKeyCredentialCreationOpts struct {
	Rp                     PublicKeyCredentialRpEntity     `json:"rp"`
	User                   PublicKeyCredentialUserEntity   `json:"user"`
	Challenge              b64URLEncoded                   `json:"challenge"`
	PubKeyCredParams       []PublicKeyCredentialParameters `json:"pubkey_cred_params"`
	Timeout                int                             `json:"timeout"`
	ExcludeCredentials     []PublicKeyCredentialDescriptor `json:"exclude_credentials,omitempty"`
	AuthenticatorSelection AuthenticatorSelectionCriteria  `json:"authenticator_selection,omitempty"`
	Attestation            AttestationConveyancePreference `json:"attestation"`
}

type PublicKeyCredentialRequestOpts struct {
	Challenge        b64URLEncoded                   `json:"challenge"`
	Timeout          int                             `json:"timeout"`
	RpID             string                          `json:"rp_id"`
	AllowCredentials []PublicKeyCredentialDescriptor `json:"allow_credentials"`
	UserVerification UserVerificationRequirement     `json:"user_verification"`
}

type CollectedClientData struct {
	Type      string        `json:"type"`
	Challenge b64URLEncoded `json:"challenge"`
	Origin    string        `json:"origin"`
}

func (b b64URLEncoded) String() string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func (b b64URLEncoded) MarshalJSON() ([]byte, error) {
	s := base64.RawURLEncoding.EncodeToString(b)
	return []byte("\"" + s + "\""), nil
}

func (b *b64URLEncoded) UnmarshalJSON(data []byte) (err error) {
	if len(data) < 2 {
		return errors.New("json: illegal data: " + string(data))
	}

	if data[0] != '"' {
		return errors.New("json: illegal at input byte 0")
	}

	if data[len(data)-1] != '"' {
		return errors.New("json: illegal data at input byte " + strconv.Itoa(len(data)-1))
	}

	*b, err = base64.RawURLEncoding.DecodeString(string(data[1 : len(data)-1]))
	return err
}

func (r PubKeyCreateRequest) MarshalJSON() ([]byte, error) {
	type Alias PubKeyCreateRequest
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyCreateRequest",
		Alias: Alias(r),
	})
}

func (r PubKeyGetRequest) MarshalJSON() ([]byte, error) {
	type Alias PubKeyGetRequest
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyGetRequest",
		Alias: Alias(r),
	})
}

func (o PublicKeyCredentialCreationOpts) MarshalJSON() ([]byte, error) {
	type Alias PublicKeyCredentialCreationOpts
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyCreateOpts",
		Alias: Alias(o),
	})
}

func (o PublicKeyCredentialRequestOpts) MarshalJSON() ([]byte, error) {
	type Alias PublicKeyCredentialRequestOpts
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyGetOpts",
		Alias: Alias(o),
	})
}

func (r AuthenticatorResponse) UnmarshalJSON(data []byte) error {
}
