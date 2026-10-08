ALTER TABLE update_events
  MODIFY client_type ENUM('dyndns','duckdns','agent','windows','shell','docker','unknown') NOT NULL DEFAULT 'unknown';
