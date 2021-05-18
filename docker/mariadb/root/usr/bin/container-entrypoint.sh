#!/bin/bash
#
# Adfinis SyGroup AG
# openshift-mariadb-galera: Container entrypoint
#

set -e
set -x

# Locations
CONTAINER_SCRIPTS_DIR="/usr/share/container-scripts/mysql"
EXTRA_DEFAULTS_FILE="/etc/my.cnf.d/galera.cnf"
# Check if the container runs in Kubernetes/OpenShift
if [ -z "$GALLERA_MODE" ]; then
	# Single container runs in docker
	echo "GALLERA_MODE not set, spin up single node"
else
	# Is running in Kubernetes/OpenShift, so find all other pods
	# belonging to the namespace
	echo "Galera: Finding peers"
	echo "Using service name: ${K8S_SVC_NAME}"
  HOSTNAME=$(echo "$MY_POD_IP" | sed -r 's/\./-/g')
  HOSTNAME=$(echo "$HOSTNAME.$K8S_SVC_NAME.${MY_POD_NAMESPACE}.svc.cluster.local")
	mkdir -p /etc/my.cnf.d
	mkdir -p /etc/mysql/conf.d
	echo "Using hostname: ${HOSTNAME}"
	echo "------------------------"
	cp ${CONTAINER_SCRIPTS_DIR}/galera.cnf ${EXTRA_DEFAULTS_FILE}
	cp ${CONTAINER_SCRIPTS_DIR}/galera.cnf /etc/mysql/conf.d/galera.cnf
	if [  -f "/var/lib/mysql/grastate.dat" ]; then
	  if  grep -q "safe_to_bootstrap: 1" "/var/lib/mysql/grastate.dat"; then
      echo "safe_to_bootstrap present"
    else
	    echo "safe_to_bootstrap: 1" >> /var/lib/mysql/grastate.dat
	    chown mysql:mysql /var/lib/mysql/grastate.dat
    fi
  else
	    echo "safe_to_bootstrap: 1" > /var/lib/mysql/grastate.dat
	    chown mysql:mysql /var/lib/mysql/grastate.dat
	fi
	/usr/bin/peer-finder -on-start="${CONTAINER_SCRIPTS_DIR}/configure-galera.sh" -labels="${LABEL_SELECTOR}" -ns=${MY_POD_NAMESPACE}
fi

chmod 0444 /etc/mysql/conf.d/server.cnf
chmod 0444 /etc/mysql/conf.d/client.cnf

# We assume that mysql needs to be setup if this directory is not present
if [ ! -d "/var/lib/mysql/mysql" ]; then
	echo "Configure first time mysql"
	${CONTAINER_SCRIPTS_DIR}/configure-mysql.sh
fi


cp /usr/share/container-scripts/mysql/readiness-probe.sh /usr/bin/readiness-probe.sh
# Run mysqld
exec mysqld