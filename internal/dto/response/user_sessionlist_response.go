package response

type UserSessionListResponse struct {
	SessionId string `json:"session_id"`
	Avatar    string `json:"avatar"`
	UserId    string `json:"user_id"`
	Username  string `json:"user_name"`
}
