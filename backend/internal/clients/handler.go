package clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

type Store interface {
	Create(context.Context, CreateClientRequest) (Client, error)
	CreateService(context.Context, string, CreateServiceRequest) (Service, error)
	List(context.Context, int, int) ([]Client, error)
	Get(context.Context, string) (Client, error)
	UpdateClient(context.Context, string, string) error
	UpdateService(context.Context, string, string, UpdateServiceRequest) error
	DeleteService(context.Context, string, string) error
	DeleteClient(context.Context, string) error
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/clients", h.createClient)
	r.GET("/clients", h.listClients)
	r.GET("/clients/:client_id", h.getClient)
	r.PATCH("/clients/:client_id", h.updateClient)
	r.DELETE("/clients/:client_id", h.deleteClient)
	r.POST("/clients/:client_id/services", h.createService)
	r.PATCH("/clients/:client_id/services/:client_service_id", h.updateService)
	r.DELETE("/clients/:client_id/services/:client_service_id", h.deleteService)
}

func (h *Handler) createClient(c *gin.Context) {
	var request CreateClientRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body ต้องเป็น JSON ที่ถูกต้อง"})
		return
	}
	request.ClientShortName = strings.TrimSpace(request.ClientShortName)
	if err := validateClientFields(request.ClientID, request.ClientShortName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Services == nil {
		request.Services = make([]CreateServiceRequest, 0)
	}
	seenServiceIDs := make(map[string]struct{}, len(request.Services))
	for i := range request.Services {
		service := &request.Services[i]
		service.ServiceID = strings.TrimSpace(service.ServiceID)
		service.ServName = strings.TrimSpace(service.ServName)
		if err := normalizeCreateService(service); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("services[%d]: %s", i, err)})
			return
		}
		if _, exists := seenServiceIDs[service.ServiceID]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service_id ใน services ต้องไม่ซ้ำกัน"})
			return
		}
		seenServiceIDs[service.ServiceID] = struct{}{}
		if service.ClientServiceID != "" &&
			service.ClientServiceID != clientServiceID(request.ClientID, service.ServiceID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("services[%d].client_service_id ไม่ตรงกับ client_id และ service_id", i)})
			return
		}
	}

	client, err := h.store.Create(c.Request.Context(), request)
	if err != nil {
		h.writeStoreError(c, "create client", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": client})
}

func (h *Handler) createService(c *gin.Context) {
	clientID := strings.TrimSpace(c.Param("client_id"))
	if err := validateID(clientID, 30, "client_id"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request CreateServiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body ต้องเป็น JSON ที่ถูกต้อง"})
		return
	}
	request.ServiceID = strings.TrimSpace(request.ServiceID)
	request.ServName = strings.TrimSpace(request.ServName)
	if err := normalizeCreateService(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.ClientServiceID != "" &&
		request.ClientServiceID != clientServiceID(clientID, request.ServiceID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_service_id ไม่ตรงกับ client_id และ service_id"})
		return
	}

	service, err := h.store.CreateService(c.Request.Context(), clientID, request)
	if err != nil {
		h.writeStoreError(c, "create service", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": service})
}

func (h *Handler) listClients(c *gin.Context) {
	limit, err := parseQueryInt(c, "limit", defaultPageSize)
	if err != nil || limit < 1 || limit > maxPageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit ต้องเป็นตัวเลขระหว่าง 1 ถึง 100"})
		return
	}
	offset, err := parseQueryInt(c, "offset", 0)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset ต้องเป็นจำนวนเต็มตั้งแต่ 0 ขึ้นไป"})
		return
	}

	clients, err := h.store.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.writeStoreError(c, "list clients", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":   clients,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) getClient(c *gin.Context) {
	clientID := strings.TrimSpace(c.Param("client_id"))
	if err := validateID(clientID, 30, "client_id"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client, err := h.store.Get(c.Request.Context(), clientID)
	if err != nil {
		h.writeStoreError(c, "get client", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": client})
}

func (h *Handler) updateClient(c *gin.Context) {
	clientID := strings.TrimSpace(c.Param("client_id"))
	if err := validateID(clientID, 30, "client_id"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var request UpdateClientRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body ต้องเป็น JSON ที่ถูกต้อง"})
		return
	}
	if request.ClientShortName == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องส่ง client_short_name"})
		return
	}
	shortName := strings.TrimSpace(*request.ClientShortName)
	if err := validateClientFields(clientID, shortName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.UpdateClient(c.Request.Context(), clientID, shortName); err != nil {
		h.writeStoreError(c, "update client", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"client_id":         clientID,
		"client_short_name": shortName,
	}})
}

func (h *Handler) updateService(c *gin.Context) {
	clientID, serviceKey, ok := servicePath(c)
	if !ok {
		return
	}
	var request UpdateServiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body ต้องเป็น JSON ที่ถูกต้อง"})
		return
	}
	if err := normalizeUpdateService(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.UpdateService(c.Request.Context(), clientID, serviceKey, request); err != nil {
		h.writeStoreError(c, "update service", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": serviceUpdateResponse(clientID, serviceKey, request)})
}

func (h *Handler) deleteService(c *gin.Context) {
	clientID, serviceKey, ok := servicePath(c)
	if !ok {
		return
	}
	if err := h.store.DeleteService(c.Request.Context(), clientID, serviceKey); err != nil {
		h.writeStoreError(c, "delete service", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) deleteClient(c *gin.Context) {
	clientID := strings.TrimSpace(c.Param("client_id"))
	if err := validateID(clientID, 30, "client_id"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.DeleteClient(c.Request.Context(), clientID); err != nil {
		h.writeStoreError(c, "delete client", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func servicePath(c *gin.Context) (string, string, bool) {
	clientID := strings.TrimSpace(c.Param("client_id"))
	serviceKey := strings.TrimSpace(c.Param("client_service_id"))
	if err := validateID(clientID, 30, "client_id"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return "", "", false
	}
	if err := validateServiceKey(clientID, serviceKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return "", "", false
	}
	return clientID, serviceKey, true
}

func validateServiceKey(clientID, serviceKey string) error {
	if serviceKey == "" || utf8.RuneCountInString(serviceKey) > 51 {
		return errors.New("client_service_id ต้องมีความยาว 1 ถึง 51 ตัวอักษร")
	}
	prefix := clientID + "_"
	if !strings.HasPrefix(serviceKey, prefix) {
		return errors.New("client_service_id ต้องขึ้นต้นด้วย client_id ตามด้วย _")
	}
	return validateID(strings.TrimPrefix(serviceKey, prefix), 20, "service_id")
}

func validateClientFields(clientID, shortName string) error {
	if err := validateID(clientID, 30, "client_id"); err != nil {
		return err
	}
	if utf8.RuneCountInString(shortName) == 0 {
		return errors.New("client_short_name ห้ามว่าง")
	}
	if utf8.RuneCountInString(shortName) > 100 {
		return errors.New("client_short_name ยาวได้ไม่เกิน 100 ตัวอักษร")
	}
	return nil
}

func validateServiceFields(serviceID, servName string) error {
	if err := validateID(serviceID, 20, "service_id"); err != nil {
		return err
	}
	return validateServiceName(servName)
}

func normalizeCreateService(request *CreateServiceRequest) error {
	if err := validateServiceFields(request.ServiceID, request.ServName); err != nil {
		return err
	}
	if request.UATStatus == "" {
		request.UATStatus = "Waiting test"
	}
	if !validUATStatus(request.UATStatus) {
		return errors.New("uat_status ไม่อยู่ในรายการสถานะที่รองรับ")
	}
	if request.ProductionStatus == "" {
		request.ProductionStatus = "NOT_PROD"
	}
	if !validProductionStatus(request.ProductionStatus) {
		return errors.New("production_status ต้องเป็น PROD, NOT_PROD หรือ NOT_APPLICABLE")
	}
	return normalizeOptionalServiceFields(true,
		&request.TestedBy,
		&request.POSOSNSS,
		&request.TestcaseResultLink,
		&request.Week,
		&request.UATSubmittedDate,
		&request.ServiceType,
	)
}

func normalizeUpdateService(request *UpdateServiceRequest) error {
	if request.ServName != nil {
		*request.ServName = strings.TrimSpace(*request.ServName)
		if err := validateServiceName(*request.ServName); err != nil {
			return err
		}
	}
	if request.UATStatus != nil && !validUATStatus(*request.UATStatus) {
		return errors.New("uat_status ไม่อยู่ในรายการสถานะที่รองรับ")
	}
	if request.ProductionStatus != nil && !validProductionStatus(*request.ProductionStatus) {
		return errors.New("production_status ต้องเป็น PROD, NOT_PROD หรือ NOT_APPLICABLE")
	}
	if request.ServName == nil && request.UATStatus == nil &&
		request.TestedBy == nil && request.POSOSNSS == nil &&
		request.TestcaseResultLink == nil && request.Week == nil &&
		request.UATSubmittedDate == nil && request.ServiceType == nil &&
		request.ProductionStatus == nil {
		return errors.New("ต้องส่งฟิลด์ที่ต้องการแก้ไขอย่างน้อยหนึ่งฟิลด์")
	}
	return normalizeOptionalServiceFields(false,
		&request.TestedBy,
		&request.POSOSNSS,
		&request.TestcaseResultLink,
		&request.Week,
		&request.UATSubmittedDate,
		&request.ServiceType,
	)
}

func normalizeOptionalServiceFields(clearEmpty bool, fields ...**string) error {
	limits := []struct {
		field string
		limit int
	}{
		{field: "tested_by", limit: 150},
		{field: "pos_os_nss", limit: 100},
		{field: "testcase_result_link", limit: 2048},
		{field: "week", limit: 20},
		{field: "uat_submitted_date", limit: 10},
		{field: "service_type", limit: 100},
	}
	for i, field := range fields {
		if *field == nil {
			continue
		}
		value := strings.TrimSpace(**field)
		if value == "" {
			if clearEmpty {
				*field = nil
			} else {
				**field = ""
			}
			continue
		}
		if utf8.RuneCountInString(value) > limits[i].limit {
			return fmt.Errorf("%s ยาวได้ไม่เกิน %d ตัวอักษร", limits[i].field, limits[i].limit)
		}
		if limits[i].field == "uat_submitted_date" {
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return errors.New("uat_submitted_date ต้องใช้รูปแบบ YYYY-MM-DD")
			}
		}
		**field = value
	}
	return nil
}

func validUATStatus(status string) bool {
	switch status {
	case "Waiting test", "Block", "Block by defect", "Defect", "Pass", "Cancel", "Hold testing":
		return true
	default:
		return false
	}
}

func validProductionStatus(status string) bool {
	switch status {
	case "PROD", "NOT_PROD", "NOT_APPLICABLE":
		return true
	default:
		return false
	}
}

func serviceUpdateResponse(clientID, serviceKey string, request UpdateServiceRequest) gin.H {
	response := gin.H{
		"client_service_id": serviceKey,
		"service_id":        strings.TrimPrefix(serviceKey, clientID+"_"),
	}
	if request.ServName != nil {
		response["serv_name"] = *request.ServName
	}
	if request.UATStatus != nil {
		response["uat_status"] = *request.UATStatus
	}
	if request.TestedBy != nil {
		response["tested_by"] = nonEmptyStringPointer(request.TestedBy)
	}
	if request.POSOSNSS != nil {
		response["pos_os_nss"] = nonEmptyStringPointer(request.POSOSNSS)
	}
	if request.TestcaseResultLink != nil {
		response["testcase_result_link"] = nonEmptyStringPointer(request.TestcaseResultLink)
	}
	if request.Week != nil {
		response["week"] = nonEmptyStringPointer(request.Week)
	}
	if request.UATSubmittedDate != nil {
		response["uat_submitted_date"] = nonEmptyStringPointer(request.UATSubmittedDate)
	}
	if request.ServiceType != nil {
		response["service_type"] = nonEmptyStringPointer(request.ServiceType)
	}
	if request.ProductionStatus != nil {
		response["production_status"] = *request.ProductionStatus
	}
	return response
}

func nonEmptyStringPointer(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}

func validateServiceName(servName string) error {
	if utf8.RuneCountInString(servName) == 0 {
		return errors.New("serv_name ห้ามว่าง")
	}
	if utf8.RuneCountInString(servName) > 255 {
		return errors.New("serv_name ยาวได้ไม่เกิน 255 ตัวอักษร")
	}
	return nil
}

func validateID(value string, maxLength int, field string) error {
	if value == "" {
		return fmt.Errorf("%s ห้ามว่าง", field)
	}
	if utf8.RuneCountInString(value) > maxLength {
		return fmt.Errorf("%s ยาวได้ไม่เกิน %d ตัวอักษร", field, maxLength)
	}
	if strings.Contains(value, "_") || strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s ห้ามมี underscore หรือช่องว่าง", field)
	}
	return nil
}

func parseQueryInt(c *gin.Context, key string, fallback int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func (h *Handler) writeStoreError(c *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลที่ระบุ"})
	case errors.Is(err, ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "ข้อมูลซ้ำกับรายการที่มีอยู่"})
	case errors.Is(err, ErrClientHasServices):
		c.JSON(http.StatusConflict, gin.H{"error": "ลบผู้ว่าจ้างไม่ได้ กรุณาลบ service ที่เกี่ยวข้องก่อน"})
	default:
		log.Printf("%s failed: %v", operation, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
	}
}
