USE trackclient;

CREATE TABLE IF NOT EXISTS clients (
  client_id VARCHAR(30) NOT NULL,
  client_short_name VARCHAR(100) NOT NULL,
  PRIMARY KEY (client_id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS services (
  client_service_id VARCHAR(51) NOT NULL,
  client_id VARCHAR(30) NOT NULL,
  service_id VARCHAR(20) NOT NULL,
  serv_name VARCHAR(255) NOT NULL,
  PRIMARY KEY (client_service_id),
  UNIQUE KEY uq_services_client_service (client_id, service_id),
  CONSTRAINT fk_services_client
    FOREIGN KEY (client_id) REFERENCES clients(client_id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
