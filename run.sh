#!/bin/sh

service cron restart
exec python3 fetch.py &
exec ./server
