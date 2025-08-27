#!/bin/sh

echo "Running database migrations..."
echo "Migration started at: $(date)"

./migrate -path migrations -database "$DATABASE_URL" up 2>&1

# Check migration exit code
migration_exit_code=$?
echo "Migration command exit code: $migration_exit_code"

if [ $migration_exit_code -eq 0 ]; then
    echo "Migration completed successfully at: $(date)"
    echo "Starting application..."
    exec ./main
else
    echo "Migration failed with exit code: $migration_exit_code at: $(date)"
    echo "Exiting without starting main application"
    exit $migration_exit_code
fi
