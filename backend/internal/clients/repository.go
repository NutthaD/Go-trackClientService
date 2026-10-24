package clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound          = errors.New("client or service not found")
	ErrConflict          = errors.New("record already exists")
	ErrClientHasServices = errors.New("client still has services")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, request CreateClientRequest) (Client, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Client{}, fmt.Errorf("begin create client transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO clients (client_id, client_short_name) VALUES (?, ?)`,
		request.ClientID, request.ClientShortName,
	)
	if err != nil {
		return Client{}, translateMySQLError(err)
	}

	client := Client{
		ClientID:        request.ClientID,
		ClientShortName: request.ClientShortName,
		Services:        make([]Service, 0, len(request.Services)),
	}
	for _, requestedService := range request.Services {
		service := serviceFromCreateRequest(request.ClientID, requestedService)
		_, err = tx.ExecContext(ctx,
			`INSERT INTO services (
				client_service_id, client_id, service_id, serv_name, uat_status,
				tested_by, pos_os_nss, testcase_result_link, week,
				uat_submitted_date, service_type, production_status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			service.ClientServiceID, request.ClientID, service.ServiceID, service.ServName,
			service.UATStatus, nullableDBValue(service.TestedBy), nullableDBValue(service.POSOSNSS),
			nullableDBValue(service.TestcaseResultLink), nullableDBValue(service.Week),
			nullableDBValue(service.UATSubmittedDate), nullableDBValue(service.ServiceType),
			service.ProductionStatus,
		)
		if err != nil {
			return Client{}, translateMySQLError(err)
		}
		client.Services = append(client.Services, service)
	}

	if err := tx.Commit(); err != nil {
		return Client{}, fmt.Errorf("commit create client transaction: %w", err)
	}
	return client, nil
}

func (r *Repository) CreateService(ctx context.Context, clientID string, request CreateServiceRequest) (Service, error) {
	service := serviceFromCreateRequest(clientID, request)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO services (
			client_service_id, client_id, service_id, serv_name, uat_status,
			tested_by, pos_os_nss, testcase_result_link, week,
			uat_submitted_date, service_type, production_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		service.ClientServiceID, clientID, service.ServiceID, service.ServName,
		service.UATStatus, nullableDBValue(service.TestedBy), nullableDBValue(service.POSOSNSS),
		nullableDBValue(service.TestcaseResultLink), nullableDBValue(service.Week),
		nullableDBValue(service.UATSubmittedDate), nullableDBValue(service.ServiceType),
		service.ProductionStatus,
	)
	if err != nil {
		return Service{}, translateMySQLError(err)
	}
	return service, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Client, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.client_id, c.client_short_name,
		       s.client_service_id, s.service_id, s.serv_name, s.uat_status,
		       s.tested_by, s.pos_os_nss, s.testcase_result_link, s.week,
		       s.uat_submitted_date, s.service_type, s.production_status
		FROM (
			SELECT client_id, client_short_name
			FROM clients
			ORDER BY client_id
			LIMIT ? OFFSET ?
		) AS c
		LEFT JOIN services AS s ON s.client_id = c.client_id
		ORDER BY c.client_id, s.service_id`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	defer rows.Close()
	return scanClients(rows)
}

func (r *Repository) Get(ctx context.Context, clientID string) (Client, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.client_id, c.client_short_name,
		       s.client_service_id, s.service_id, s.serv_name, s.uat_status,
		       s.tested_by, s.pos_os_nss, s.testcase_result_link, s.week,
		       s.uat_submitted_date, s.service_type, s.production_status
		FROM clients AS c
		LEFT JOIN services AS s ON s.client_id = c.client_id
		WHERE c.client_id = ?
		ORDER BY s.service_id`, clientID,
	)
	if err != nil {
		return Client{}, fmt.Errorf("get client: %w", err)
	}
	defer rows.Close()

	clients, err := scanClients(rows)
	if err != nil {
		return Client{}, err
	}
	if len(clients) == 0 {
		return Client{}, ErrNotFound
	}
	return clients[0], nil
}

func (r *Repository) UpdateClient(ctx context.Context, clientID, shortName string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE clients SET client_short_name = ? WHERE client_id = ?`,
		shortName, clientID,
	)
	if err != nil {
		return fmt.Errorf("update client: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated client: %w", err)
	}
	if affected == 0 {
		var exists int
		err := r.db.QueryRowContext(ctx,
			`SELECT 1 FROM clients WHERE client_id = ?`, clientID,
		).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check client after update: %w", err)
		}
	}
	return nil
}

func (r *Repository) UpdateService(ctx context.Context, clientID, clientServiceID string, request UpdateServiceRequest) error {
	updates := make([]string, 0, 9)
	args := make([]any, 0, 11)
	if request.ServName != nil {
		updates = append(updates, "serv_name = ?")
		args = append(args, *request.ServName)
	}
	if request.UATStatus != nil {
		updates = append(updates, "uat_status = ?")
		args = append(args, *request.UATStatus)
	}
	if request.TestedBy != nil {
		updates = append(updates, "tested_by = ?")
		args = append(args, nullableDBValue(request.TestedBy))
	}
	if request.POSOSNSS != nil {
		updates = append(updates, "pos_os_nss = ?")
		args = append(args, nullableDBValue(request.POSOSNSS))
	}
	if request.TestcaseResultLink != nil {
		updates = append(updates, "testcase_result_link = ?")
		args = append(args, nullableDBValue(request.TestcaseResultLink))
	}
	if request.Week != nil {
		updates = append(updates, "`week` = ?")
		args = append(args, nullableDBValue(request.Week))
	}
	if request.UATSubmittedDate != nil {
		updates = append(updates, "uat_submitted_date = ?")
		args = append(args, nullableDBValue(request.UATSubmittedDate))
	}
	if request.ServiceType != nil {
		updates = append(updates, "service_type = ?")
		args = append(args, nullableDBValue(request.ServiceType))
	}
	if request.ProductionStatus != nil {
		updates = append(updates, "production_status = ?")
		args = append(args, *request.ProductionStatus)
	}
	if len(updates) == 0 {
		return errors.New("no service fields to update")
	}
	args = append(args, clientID, clientServiceID)
	query := fmt.Sprintf(
		"UPDATE services SET %s WHERE client_id = ? AND client_service_id = ?",
		strings.Join(updates, ", "),
	)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated service: %w", err)
	}
	if affected == 0 {
		var exists int
		err := r.db.QueryRowContext(ctx,
			`SELECT 1 FROM services WHERE client_id = ? AND client_service_id = ?`,
			clientID, clientServiceID,
		).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check service after update: %w", err)
		}
	}
	return nil
}

func (r *Repository) DeleteService(ctx context.Context, clientID, clientServiceID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM services WHERE client_id = ? AND client_service_id = ?`,
		clientID, clientServiceID,
	)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted service: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteClient(ctx context.Context, clientID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete client transaction: %w", err)
	}
	defer tx.Rollback()

	var found string
	err = tx.QueryRowContext(ctx,
		`SELECT client_id FROM clients WHERE client_id = ? FOR UPDATE`, clientID,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock client before delete: %w", err)
	}

	var serviceCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM services WHERE client_id = ?`, clientID,
	).Scan(&serviceCount); err != nil {
		return fmt.Errorf("count client services before delete: %w", err)
	}
	if serviceCount > 0 {
		return ErrClientHasServices
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM clients WHERE client_id = ?`, clientID,
	); err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete client transaction: %w", err)
	}
	return nil
}

func scanClients(rows *sql.Rows) ([]Client, error) {
	clients := make([]Client, 0)
	indexByID := make(map[string]int)

	for rows.Next() {
		var clientID, shortName string
		var clientServiceID, serviceID, serviceName sql.NullString
		var uatStatus, productionStatus sql.NullString
		var testedBy, posOSNSS, testcaseResultLink, week, serviceType sql.NullString
		var submittedDate sql.NullTime
		if err := rows.Scan(
			&clientID, &shortName, &clientServiceID, &serviceID, &serviceName,
			&uatStatus, &testedBy, &posOSNSS, &testcaseResultLink, &week,
			&submittedDate, &serviceType, &productionStatus,
		); err != nil {
			return nil, fmt.Errorf("scan client row: %w", err)
		}

		index, exists := indexByID[clientID]
		if !exists {
			index = len(clients)
			indexByID[clientID] = index
			clients = append(clients, Client{
				ClientID:        clientID,
				ClientShortName: shortName,
				Services:        make([]Service, 0),
			})
		}
		if clientServiceID.Valid {
			clients[index].Services = append(clients[index].Services, Service{
				ClientServiceID:    clientServiceID.String,
				ServiceID:          serviceID.String,
				ServName:           serviceName.String,
				UATStatus:          uatStatus.String,
				TestedBy:           nullableString(testedBy),
				POSOSNSS:           nullableString(posOSNSS),
				TestcaseResultLink: nullableString(testcaseResultLink),
				Week:               nullableString(week),
				UATSubmittedDate:   nullableDate(submittedDate),
				ServiceType:        nullableString(serviceType),
				ProductionStatus:   productionStatus.String,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate client rows: %w", err)
	}
	return clients, nil
}

func translateMySQLError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrConflict
		case 1452:
			return ErrNotFound
		}
	}
	return err
}

func clientServiceID(clientID, serviceID string) string {
	return clientID + "_" + serviceID
}

func serviceFromCreateRequest(clientID string, request CreateServiceRequest) Service {
	return Service{
		ClientServiceID:    clientServiceID(clientID, request.ServiceID),
		ServiceID:          request.ServiceID,
		ServName:           request.ServName,
		UATStatus:          request.UATStatus,
		TestedBy:           request.TestedBy,
		POSOSNSS:           request.POSOSNSS,
		TestcaseResultLink: request.TestcaseResultLink,
		Week:               request.Week,
		UATSubmittedDate:   request.UATSubmittedDate,
		ServiceType:        request.ServiceType,
		ProductionStatus:   request.ProductionStatus,
	}
}

func nullableDBValue(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return *value
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableDate(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format("2006-01-02")
	return &formatted
}
