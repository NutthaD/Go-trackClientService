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
  uat_status ENUM(
    'Waiting test',
    'Block',
    'Block by defect',
    'Defect',
    'Pass',
    'Cancel',
    'Hold testing'
  ) NOT NULL DEFAULT 'Waiting test',
  tested_by VARCHAR(150) NULL,
  pos_os_nss VARCHAR(100) NULL,
  testcase_result_link VARCHAR(2048) NULL,
  `week` VARCHAR(20) NULL,
  uat_submitted_date DATE NULL,
  service_type VARCHAR(100) NULL,
  production_status ENUM(
    'PROD',
    'NOT_PROD',
    'NOT_APPLICABLE'
  ) NOT NULL DEFAULT 'NOT_PROD',
  PRIMARY KEY (client_service_id),
  UNIQUE KEY uq_services_client_service (client_id, service_id),
  CONSTRAINT fk_services_client
    FOREIGN KEY (client_id) REFERENCES clients(client_id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
