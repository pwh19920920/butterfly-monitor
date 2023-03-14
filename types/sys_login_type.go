package types

type SysLoginAuthorizeRequest struct {
	ClientId string `json:"clientId"`
}

type SysLoginAuthorizeResponse struct {
	Code        string `json:"code"`
	RedirectUrl string `json:"redirectUrl"`
}

type SysLoginTokenRequest struct {
	Code string `form:"code"`
}

type SysLoginTokenResponse struct {
	AccessToken string  `json:"access_token"`
	TokenType   string  `json:"token_type"` // 类型Bearer
	ExpiryIn    float64 `json:"expiry_in"`  // 剩余过期时间
}

type SysLoginUserInfoResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
