#!/bin/sh

cd /app
envsubst < reporter.cfg.template > reporter.cfg
#/usr/local/sbin/upgrade_db.sh
./reporter
