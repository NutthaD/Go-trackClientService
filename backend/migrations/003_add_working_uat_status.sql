USE trackclient;

-- Run this after migration 002 on an existing database that already has uat_status.
-- New databases created with migration 001 already include Working.
ALTER TABLE services
  MODIFY COLUMN uat_status ENUM(
    'Waiting test',
    'Working',
    'Block',
    'Block by defect',
    'Defect',
    'Pass',
    'Cancel',
    'Hold testing'
  ) NOT NULL DEFAULT 'Waiting test';
