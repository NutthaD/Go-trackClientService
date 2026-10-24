USE trackclient;

-- Run this only when upgrading a database created with the old 001 migration.
-- New databases should use the updated 001_create_clients_services.sql instead.
ALTER TABLE services
  ADD COLUMN uat_status ENUM(
    'Waiting test',
    'Block',
    'Block by defect',
    'Defect',
    'Pass',
    'Cancel',
    'Hold testing'
  ) NOT NULL DEFAULT 'Waiting test',
  ADD COLUMN tested_by VARCHAR(150) NULL,
  ADD COLUMN pos_os_nss VARCHAR(100) NULL,
  ADD COLUMN testcase_result_link VARCHAR(2048) NULL,
  ADD COLUMN `week` VARCHAR(20) NULL,
  ADD COLUMN uat_submitted_date DATE NULL,
  ADD COLUMN service_type VARCHAR(100) NULL,
  ADD COLUMN production_status ENUM(
    'PROD',
    'NOT_PROD',
    'NOT_APPLICABLE'
  ) NOT NULL DEFAULT 'NOT_PROD';
