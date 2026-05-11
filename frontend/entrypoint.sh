#!/bin/sh
set -e
npm run build
exec nginx -g 'daemon off;'
