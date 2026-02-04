package protocol

type RegisterRequest struct {
	RequestedSubdomain string `json:"requested_subdomain"`
}

type RegisterResponse struct {
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
}
