#!/bin/bash
set -e

# Check if the database file exists
if [ -f ./pb_data/data.db ]; then
  echo "Database already exists, skipping restore"
else
  echo "No database found, restoring from replica if exists"
  # Attempt to restore from the replica
  ./litestream restore -if-replica-exists -o /pb_data/data.db "${REPLICA_URL}"
  
  # Check the exit status of the restore command
  if [ $? -eq 0 ]; then
    echo "Restore failed"
  else
    echo "Restore successful"
    # Do not exit the script here
  fi
fi

# Run litestream with your app as the subprocess.
# Use port 8080 for deploying to Fly.io, GCP Cloud Run, or AWS App Runner easily.
exec ./litestream replicate -exec "./pocketbase serve --http 0.0.0.0:8080"