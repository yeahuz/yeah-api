package credential

import (
	"encoding/json"
)

type PubKeyCredential struct {
	ID                  string
	CredentialID        string
	Title               string
	PubKey              string
	PubKeyAlg           int
	Transports          []AuthenticatorTransport
	UserID              string
	Counter             uint32
	CredentialRequestID string
}

type Request struct {
	ID        string
	Type      string
	Challenge string
	Used      bool
	UserID    string
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
	ID     string                        `json:"id"`
	PubKey *pubKeyCredentialCreationOpts `json:"pubkey"`
	kind   string
}

type pubKeyGetRequest struct {
	ID     string                       `json:"id"`
	PubKey *pubKeyCredentialRequestOpts `json:"pubkey"`
	userID string
	kind   string
}

type collectedClientData struct {
	Raw       []byte
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

type authenticatorResponse struct {
	ClientDataJSON string `json:"client_data_json"`
}

type authenticatorData struct {
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

type rawPubKeyCredentialAttestation struct {
	rawPubKeyCredential
	Response AuthenticatorAttestationResponse `json:"response"`
}

type rawPubKeyCredentialAssertion struct {
	rawPubKeyCredential
	Response AuthenticatorAssertionResponse `json:"response"`
}

type CreatePubKeyData struct {
	ReqID      string                         `json:"req_id"`
	Credential rawPubKeyCredentialAttestation `json:"credential"`
	Title      string
}

type AssertPubKeyData struct {
	ReqID      string                       `json:"req_id"`
	Credential rawPubKeyCredentialAssertion `json:"credential"`
}

type pubKeyCredentialRpEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type pubKeyCredentialUserEntity struct {
	id          string
	EncodedID   string `json:"id"`
	DisplayName string `json:"display_name"`
}

type authenticatorSelectionCriteria struct {
	AuthenticatorAttachment AuthenticatorAttachment     `json:"authenticator_attachment,omitempty"`
	ResidentKey             ResidentKeyRequirement      `json:"resident_key,omitempty"`
	RequireResidentKey      bool                        `json:"require_resident_key,omitempty"`
	UserVerification        UserVerificationRequirement `json:"user_verification,omitempty"`
}

type pubKeyCredentialDescriptor struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Transports []AuthenticatorTransport `json:"transports"`
}

type pubKeyCredentialParameters struct {
	Type string                  `json:"type"`
	Alg  COSEAlgorithmIdentifier `json:"alg"`
}

type pubKeyCredentialCreationOpts struct {
	Rp                     pubKeyCredentialRpEntity        `json:"rp"`
	User                   pubKeyCredentialUserEntity      `json:"user"`
	Challenge              string                          `json:"challenge"`
	PubKeyCredParams       []pubKeyCredentialParameters    `json:"pubkey_cred_params"`
	Timeout                int                             `json:"timeout"`
	ExcludeCredentials     []pubKeyCredentialDescriptor    `json:"exclude_credentials,omitempty"`
	AuthenticatorSelection authenticatorSelectionCriteria  `json:"authenticator_selection,omitempty"`
	Attestation            AttestationConveyancePreference `json:"attestation"`
}

type pubKeyCredentialRequestOpts struct {
	Challenge        string                       `json:"challenge"`
	Timeout          int                          `json:"timeout"`
	RpID             string                       `json:"rp_id"`
	AllowCredentials []pubKeyCredentialDescriptor `json:"allow_credentials"`
	UserVerification UserVerificationRequirement  `json:"user_verification"`
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

func (r pubKeyGetRequest) MarshalJSON() ([]byte, error) {
	type Alias pubKeyGetRequest
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyGetRequest",
		Alias: Alias(r),
	})
}

func (o pubKeyCredentialCreationOpts) MarshalJSON() ([]byte, error) {
	type Alias pubKeyCredentialCreationOpts
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyCreateOpts",
		Alias: Alias(o),
	})
}

func (o pubKeyCredentialRequestOpts) MarshalJSON() ([]byte, error) {
	type Alias pubKeyCredentialRequestOpts
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "credentials.pubKeyGetOpts",
		Alias: Alias(o),
	})
}
