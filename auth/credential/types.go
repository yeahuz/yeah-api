package credential

import "encoding/json"

type AttestationType string
type PubKeyCredType string

const (
	AttestationDirect   AttestationType = "direct"
	AttestationIndirect AttestationType = "indirect"
	AttestationNone     AttestationType = "none"
	PubKeyCredPubKey    PubKeyCredType  = "public-key"
	PubKeyCredPassword  PubKeyCredType  = "password"
)

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

type clientData struct {
	challenge string
	typ       string
	origin    string
}

type AttestationResponse struct {
	ClientDataJSON    string `json:"client_data_json"`
	AttestationObject string `json:"attestation_object"`
}

type CreateResponse struct {
	ID       string               `json:"id"`
	RawID    string               `json:"raw_id"`
	Response *AttestationResponse `json:"response"`
	Type     PubKeyCredType       `json:"type"`
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
