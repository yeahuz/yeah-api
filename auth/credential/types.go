package credential

import (
	"encoding/json"
)

type AttestationType string
type PubKeyCredType string
type authFlag byte

const (
	AttestationDirect   AttestationType = "direct"
	AttestationIndirect AttestationType = "indirect"
	AttestationNone     AttestationType = "none"

	PubKeyCredPubKey   PubKeyCredType = "public-key"
	PubKeyCredPassword PubKeyCredType = "password"
)

// https://github.com/go-webauthn/webauthn/blob/master/protocol/authenticator.go#L175
const (
	// FlagUserPresent Bit 00000001 in the byte sequence. Tells us if user is present. Also referred to as the UP flag.
	FlagUserPresent authFlag = 1 << iota

	// FlagRFU1 is a reserved for future use flag.
	FlagRFU1
	// FlagUserVerified Bit 00000100 in the byte sequence. Tells us if user is verified
	// by the authenticator using a biometric or PIN. Also referred to as the UV flag.
	FlagUserVerified

	// FlagBackupEligible Bit 00001000 in the byte sequence. Tells us if a backup is eligible for device. Also referred
	// to as the BE flag.
	FlagBackupEligible // Referred to as BE

	// FlagBackupState Bit 00010000 in the byte sequence. Tells us if a backup state for device. Also referred to as the
	// BS flag.
	FlagBackupState

	// FlagRFU2 is a reserved for future use flag.
	FlagRFU2

	// FlagAttestedCredentialData Bit 01000000 in the byte sequence. Indicates whether
	// the authenticator added attested credential data. Also referred to as the AT flag.
	FlagAttestedCredentialData

	// FlagHasExtensions Bit 10000000 in the byte sequence. Indicates if the authenticator data has extensions. Also
	// referred to as the ED flag.
	FlagHasExtensions
)

type urlEncodedB64 []byte

type PubKeyCredParam struct {
	Alg  int            `json:"alg"`
	Type PubKeyCredType `json:"type"`
}

type Opts struct {
	CredentialID string
	Title        string
	PubKey       string
	Counter      int
	UserID       string
	Transports   []string
}

type Rp struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
}

type Request struct {
	Type      string
	Challenge string
	Used      bool
}

type CreateRequest struct {
	ID               string            `json:"id"`
	Challenge        string            `json:"challenge"`
	Rp               Rp                `json:"rp"`
	User             User              `json:"user"`
	Timeout          int               `json:"timeout"`
	Attestation      AttestationType   `json:"attestation"`
	PubKeyCredParams []PubKeyCredParam `json:"pubkey_cred_params"`
}

type allowedCredential struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Transports []string `json:"transports"`
}

type GetRequest struct {
	ID               string              `json:"id"`
	Challenge        string              `json:"challenge"`
	RpID             string              `json:"rpId"`
	Timeout          int                 `json:"timeout"`
	AllowCredentials []allowedCredential `json:"allow_credentials"`
}

type Credential struct {
	ID string
	Opts
}

type parsedClientData struct {
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
	Origin    string `json:"origin"`
}

type attestedCredentialData struct {
	aaguid           []byte
	credentialID     []byte
	credentialPubkey []byte
}

type authenticatorData struct {
	rpIdHash []byte
	flags    authFlag
	counter  uint32
	attData  attestedCredentialData
	extData  []byte
}

type parsedAttestationObject struct {
	authData authenticatorData
}

type AttestationResponse struct {
	ClientDataJSON    string `json:"client_data_json"`
	AttestationObject string `json:"attestation_object"`
}

type parsedAttestationResponse struct {
	clientData        parsedClientData
	attestationObject parsedAttestationObject
}

type CreateResponse struct {
	ID         string               `json:"id"`
	RawID      string               `json:"raw_id"`
	Response   *AttestationResponse `json:"response"`
	Type       PubKeyCredType       `json:"type"`
	Transports []string             `json:"transports"`
}

type CreateCredentialData struct {
	ReqID      string          `json:"req_id"`
	Credential *CreateResponse `json:"credential"`
	Title      string          `json:"title"`
}

type Credentials struct {
	Credentials []Credential `json:"credentials"`
	Count       int          `json:"count"`
}

func (gr GetRequest) MarshalJSON() ([]byte, error) {
	type Alias GetRequest
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.credentialGetRequest",
		Alias: Alias(gr),
	})
}

func (cr CreateRequest) MarshalJSON() ([]byte, error) {
	type Alias CreateRequest
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.credentialCreateRequest",
		Alias: Alias(cr),
	})
}

func (c Credential) MarshalJSON() ([]byte, error) {
	type Alias Credential
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.credential",
		Alias: Alias(c),
	})
}

func (cs Credentials) MarshalJSON() ([]byte, error) {
	type Alias Credentials
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.credentials",
		Alias: Alias(cs),
	})
}
