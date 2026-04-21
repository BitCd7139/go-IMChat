package request

type RegisterRequest struct {
	Nickname  string `json:"nickname"`
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
	SmScode   string `json:"sms_code"`
}
