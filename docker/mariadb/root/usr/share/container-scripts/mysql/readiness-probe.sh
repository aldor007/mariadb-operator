#!/bin/bash
#
# Adfinis SyGroup AG,
# openshift-mariadb-galera: mysqld readinessProbe
#

MYSQL_USER="readinessProbe"
MYSQL_PASS="readinessProbe"
MYSQL_HOST="localhost"

mysql --protocol=socket --socket=/var/run/mysqld/mysql.sock --connect-timeout=10 -u${MYSQL_USER} -p${MYSQL_PASS} -h${MYSQL_HOST} -e"SHOW DATABASES;"

if [ $? -ne 0 ]; then
  exit 1
else
  exit 0
fi
