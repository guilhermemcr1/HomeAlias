UPDATE update_events SET client_type = 'unknown' WHERE client_type IN ('shell','docker');
ALTER TABLE update_events
  MODIFY client_type ENUM('dyndns','duckdns','agent','windows','unknown') NOT NULL DEFAULT 'unknown';
