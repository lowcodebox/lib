package client

type AuthResponse struct {
	XAuthToken  string `json:"x_auth_token"`
	UserUID     string `json:"user_uid"`
	ProfileUID  string `json:"profile_uid"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Ref         string `json:"ref"`
}
