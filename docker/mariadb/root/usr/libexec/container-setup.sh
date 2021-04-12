#!/bin/bash
#
# Adfinis SyGroup AG,
# openshift-mariadb-galera: Container setup script
#

# mysqld user:group
MYSQL_SYS_USER="mysql"
MYSQL_SYS_GROUP="mysql"

# Locations
MYSQL_DATA_DIR="/var/lib/mysql"
MYSQL_RUN_DIR="/var/run/mysqld"
MYSQL_LOG_DIR="/var/log/mysql"

# Fix data directory permissions
rm -rf ${MYSQL_DATA_DIR}
mkdir -p ${MYSQL_DATA_DIR}
chown -R ${MYSQL_SYS_USER}:${MYSQL_SYS_GROUP} ${MYSQL_DATA_DIR}
chmod -R g+w ${MYSQL_DATA_DIR}

# Create run directory
mkdir -p ${MYSQL_RUN_DIR}
chown -R ${MYSQL_SYS_USER}:${MYSQL_SYS_GROUP} ${MYSQL_RUN_DIR}
chmod -R g+w ${MYSQL_RUN_DIR}

# create log dir
mkdir -p ${MYSQL_LOG_DIR}
chown -R ${MYSQL_SYS_USER}:${MYSQL_SYS_GROUP} ${MYSQL_LOG_DIR}
chmod -R g+w ${MYSQL_LOG_DIR}