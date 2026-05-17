-- migrations/001_initial_schema.sql

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- packages table stores different service plans ISP offers
CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    -- Mikrotik-Rate-Limit format e.g., 10M/10M
    rate_limit VARCHAR(100) NOT NULL,
    -- Price for reference
    price NUMERIC(10, 2) DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Status enum for users
CREATE TYPE user_status AS ENUM ('active', 'expired', 'suspended');

-- users table for subscriber information
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    -- Using bcrypt for password hashing
    password VARCHAR(255) NOT NULL,
    package_id UUID NOT NULL REFERENCES packages(id),
    status user_status NOT NULL DEFAULT 'active',
    -- Expiry date for the user's package
    expiry_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_expiry_date ON users(expiry_date);

-- sessions table for accounting data
CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    -- Acct-Session-Id from RADIUS packet
    session_id VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    nas_ip_address VARCHAR(45),
    calling_station_id VARCHAR(50), -- User's MAC address
    session_start_time TIMESTAMPTZ,
    session_stop_time TIMESTAMPTZ,
    session_total_time INT, -- in seconds
    input_octets BIGINT,
    output_octets BIGINT,
    -- Acct-Terminate-Cause
    terminate_cause VARCHAR(255)
);

-- Create indexes for faster lookups on sessions table
CREATE UNIQUE INDEX idx_sessions_session_id_username ON sessions(session_id, username);
CREATE INDEX idx_sessions_username ON sessions(username);
CREATE INDEX idx_sessions_start_time ON sessions(session_start_time);
CREATE INDEX idx_sessions_stop_time ON sessions(session_stop_time);

-- Insert some sample data for testing
INSERT INTO packages (name, rate_limit, price) VALUES
('Basic 5M', '5M/5M', 20.00),
('Standard 10M', '10M/10M', 35.00),
('Premium 50M', '50M/50M', 60.00);

-- Pass: password123
-- Note: You would generate this hash in your application, this is just an example
INSERT INTO users (username, password, package_id, status, expiry_date) VALUES
('testuser', '$2a$10$f.wT8zY.d.e.j/k.l.m.n.o.p.q.r.s.t.u.v.w.x.y.z.A', (SELECT id FROM packages WHERE name = 'Standard 10M'), 'active', NOW() + INTERVAL '30 days');