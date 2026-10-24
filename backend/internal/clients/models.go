package clients

type Client struct {
	ClientID        string    `json:"client_id"`
	ClientShortName string    `json:"client_short_name"`
	Services        []Service `json:"services"`
}

type Service struct {
	ClientServiceID    string  `json:"client_service_id"`
	ServiceID          string  `json:"service_id"`
	ServName           string  `json:"serv_name"`
	UATStatus          string  `json:"uat_status"`
	TestedBy           *string `json:"tested_by"`
	POSOSNSS           *string `json:"pos_os_nss"`
	TestcaseResultLink *string `json:"testcase_result_link"`
	Week               *string `json:"week"`
	UATSubmittedDate   *string `json:"uat_submitted_date"`
	ServiceType        *string `json:"service_type"`
	ProductionStatus   string  `json:"production_status"`
}

type CreateClientRequest struct {
	ClientID        string                 `json:"client_id"`
	ClientShortName string                 `json:"client_short_name"`
	Services        []CreateServiceRequest `json:"services"`
}

type CreateServiceRequest struct {
	ClientServiceID    string  `json:"client_service_id"`
	ServiceID          string  `json:"service_id"`
	ServName           string  `json:"serv_name"`
	UATStatus          string  `json:"uat_status"`
	TestedBy           *string `json:"tested_by"`
	POSOSNSS           *string `json:"pos_os_nss"`
	TestcaseResultLink *string `json:"testcase_result_link"`
	Week               *string `json:"week"`
	UATSubmittedDate   *string `json:"uat_submitted_date"`
	ServiceType        *string `json:"service_type"`
	ProductionStatus   string  `json:"production_status"`
}

type UpdateClientRequest struct {
	ClientShortName *string `json:"client_short_name"`
}

type UpdateServiceRequest struct {
	ServName           *string `json:"serv_name"`
	UATStatus          *string `json:"uat_status"`
	TestedBy           *string `json:"tested_by"`
	POSOSNSS           *string `json:"pos_os_nss"`
	TestcaseResultLink *string `json:"testcase_result_link"`
	Week               *string `json:"week"`
	UATSubmittedDate   *string `json:"uat_submitted_date"`
	ServiceType        *string `json:"service_type"`
	ProductionStatus   *string `json:"production_status"`
}
