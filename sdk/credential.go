package sdk

// Credential is used to sign the request
type Credential struct {
	SecretID  string
	SecretKey string
}

func NewCredentials(secretID, secretKey string) *Credential {
	return &Credential{secretID, secretKey}
}
