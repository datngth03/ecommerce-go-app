-- Initialize PostgreSQL databases for E-commerce microservices
-- This script creates separate databases for each service

-- Create databases
CREATE DATABASE users_db;
CREATE DATABASE products_db;
CREATE DATABASE orders_db;
CREATE DATABASE payments_db;
CREATE DATABASE inventory_db;
CREATE DATABASE notifications_db;

-- Grant all privileges to postgres user (already owner, but explicit)
GRANT ALL PRIVILEGES ON DATABASE users_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE products_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE orders_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE payments_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE inventory_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE notifications_db TO postgres;

-- Connect to each database and create extensions if needed
\c users_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c products_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c orders_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c payments_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c inventory_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c notifications_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
