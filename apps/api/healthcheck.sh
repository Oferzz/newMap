#!/bin/sh
# Health check script that uses the PORT environment variable
wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/api/health || exit 1