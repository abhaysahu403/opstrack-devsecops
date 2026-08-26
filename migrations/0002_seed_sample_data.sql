-- 0002_seed_sample_data.sql
-- Optional demo-friendly sample data. Safe to run multiple times because
-- the migration runner records applied files and only runs this once.
-- To skip seeding entirely, delete this file before first run.

INSERT INTO services (name, description, owner, environment, status)
VALUES
    ('Payment Service', 'Handles payment authorization and capture', 'payments-team', 'production', 'ACTIVE'),
    ('User Service', 'Manages user accounts and authentication', 'identity-team', 'production', 'ACTIVE'),
    ('Notification Service', 'Sends email/SMS/push notifications', 'platform-team', 'production', 'DEGRADED'),
    ('Order Service', 'Manages order lifecycle', 'commerce-team', 'staging', 'ACTIVE')
ON CONFLICT DO NOTHING;

INSERT INTO incidents (service_id, title, description, severity, status, assigned_to)
SELECT id, 'Payment API latency', 'Elevated p99 latency on the payment authorization endpoint', 'HIGH', 'INVESTIGATING', 'oncall-payments'
FROM services WHERE name = 'Payment Service'
ON CONFLICT DO NOTHING;

INSERT INTO incidents (service_id, title, description, severity, status, assigned_to)
SELECT id, 'Notification delivery failure', 'Push notifications failing to deliver for a subset of users', 'MEDIUM', 'OPEN', 'oncall-platform'
FROM services WHERE name = 'Notification Service'
ON CONFLICT DO NOTHING;

INSERT INTO incidents (service_id, title, description, severity, status, assigned_to)
SELECT id, 'Database connection saturation', 'Connection pool exhaustion under peak load', 'CRITICAL', 'MITIGATED', 'oncall-commerce'
FROM services WHERE name = 'Order Service'
ON CONFLICT DO NOTHING;
