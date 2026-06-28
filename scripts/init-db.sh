#!/bin/bash
set -e

# This script runs when the PostgreSQL container is first initialized
# It creates the test database alongside the main database

echo "Creating additional databases..."

# Create test database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create test database if it doesn't exist
    SELECT 'CREATE DATABASE gosql_test'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'gosql_test')\gexec

    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE gosql_test TO $POSTGRES_USER;
EOSQL

echo "Additional databases created successfully."
