package models

type ABDMSessionRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	GrantType    string `json:"grantType"`
}

type ABDMTokenResponse struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
}

type ABDMSessionErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type ABDMPublicKeyResponse struct {
	PublicKey           string `json:"publicKey"`
	EncryptionAlgorithm string `json:"encryptionAlgorithm"`
}

type ABDMOtpRequest struct {
	Scope     []string `json:"scope"`
	LoginHint string   `json:"loginHint"`
	LoginId   string   `json:"loginId"`
	OtpSystem string   `json:"otpSystem"`
}

type ABDMOtpResponse struct {
	TxnID   string `json:"txnId"`
	Message string `json:"message"`
}

type ABDMVerifyOtpRequest struct {
	Scope    []string `json:"scope"`
	AuthData struct {
		AuthMethods []string `json:"authMethods"`
		Otp         struct {
			TxnID    string `json:"txnId"`
			OtpValue string `json:"otpValue"`
		} `json:"otp"`
	} `json:"authData"`
}

type ABDMOtpVerifyResponse struct {
	TxnID      string `json:"txnId"`
	AuthResult string `json:"authResult"`
	Message    string `json:"message"`
	Token      string `json:"token"`
	Accounts   []struct {
		AbhaNumber           string `json:"ABHANumber"`
		PreferredAbhaAddress string `json:"preferredAbhaAddress"`
		Name                 string `json:"name"`
		Gender               string `json:"gender"`
		Dob                  string `json:"dob"`
	} `json:"accounts"`
}

type ABDMUserVerifyRequest struct {
	ABHANumber string `json:"ABHANumber"`
	TxnID      string `json:"txnId"`
}

type ABDMUserVerifyResponse struct {
	Token            string `json:"token"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

type AbdmCreateAbhaByAadhaarResponse struct {
	Message string `json:"message"`
	TxnId   string `json:"txnId"`
	Tokens  struct {
		Token            string `json:"token"`
		ExpiresIn        int    `json:"expiresIn"`
		RefreshToken     string `json:"refreshToken"`
		RefreshExpiresIn int    `json:"refreshExpiresIn"`
	} `json:"tokens"`
	ABHAProfile struct {
		FirstName    string   `json:"firstName"`
		MiddleName   string   `json:"middleName"`
		LastName     string   `json:"lastName"`
		Dob          string   `json:"dob"`
		Gender       string   `json:"gender"`
		Mobile       string   `json:"mobile"`
		ABHANumber   string   `json:"ABHANumber"`
		AbhaStatus   string   `json:"abhaStatus"`
		PhrAddress   []string `json:"phrAddress"`
		Address      string   `json:"address"`
		StateName    string   `json:"stateName"`
		DistrictName string   `json:"districtName"`
		Photo        string   `json:"photo"`
	} `json:"ABHAProfile"`
	IsNew bool `json:"isNew"`
}

type AbdmAbhaAddressSuggestionResponse struct {
	TxnId           string   `json:"txnId"`
	AbhaAddressList []string `json:"abhaAddressList"`
}

type AbdmVerifyAadhaarOtpResponse struct {
	TxnId           string   `json:"txnId"`
	AbhaAddressList []string `json:"abhaAddressList"`
	Token           string   `json:"token"`
	ABHAProfile     struct {
		FirstName    string   `json:"firstName"`
		MiddleName   string   `json:"middleName"`
		LastName     string   `json:"lastName"`
		Dob          string   `json:"dob"`
		Gender       string   `json:"gender"`
		Mobile       string   `json:"mobile"`
		ABHANumber   string   `json:"ABHANumber"`
		AbhaStatus   string   `json:"abhaStatus"`
		PhrAddress   []string `json:"phrAddress"`
		Address      string   `json:"address"`
		StateName    string   `json:"stateName"`
		DistrictName string   `json:"districtName"`
		Photo        string   `json:"photo"`
	} `json:"ABHAProfile"`
	IsNew bool `json:"isNew"`
}

type AbdmCreateAbhaAddressResponse struct {
	TxnId                string `json:"txnId"`
	HealthIdNumber       string `json:"healthIdNumber"`
	PreferredAbhaAddress string `json:"preferredAbhaAddress"`
}
