package clients

type Client struct {
	ClientID        string    `json:"client_id"`
	ClientShortName string    `json:"client_short_name"`
	Services        []Service `json:"services"`
}

type Service struct {
	ClientServiceID string `json:"client_service_id"`
	ServiceID       string `json:"service_id"`
	ServName        string `json:"serv_name"`
}

type CreateClientRequest struct {
	ClientID        string                 `json:"client_id"`
	ClientShortName string                 `json:"client_short_name"`
	Services        []CreateServiceRequest `json:"services"`
}

type CreateServiceRequest struct {
	ClientServiceID string `json:"client_service_id"`
	ServiceID       string `json:"service_id"`
	ServName        string `json:"serv_name"`
}

type UpdateClientRequest struct {
	ClientShortName *string `json:"client_short_name"`
}

type UpdateServiceRequest struct {
	ServName *string `json:"serv_name"`
}
