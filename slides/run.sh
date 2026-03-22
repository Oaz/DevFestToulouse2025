#!/bin/sh

# config.js define variables to be used in scripts part
cat <<EOF > /slidesk/assets/code/config.js
const API_URL = "${API_URL:=http://localhost:8086}";
const ADMIN_PASSWORD = "${ADMIN_PASSWORD:=foobar}";
EOF

# .env define variables for slides part, usage ++KEY++
cat <<EOF > /slidesk/.env
FRONT_URL="${FRONT_URL:=http://localhost:5173}"
EOF

# execute CMD defined in Dockerfile
exec "$@"
