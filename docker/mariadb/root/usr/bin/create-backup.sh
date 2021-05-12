#!/bin/bash
#
#

set -x

if [ -z "$BACKUP_URL" ]; then
  echo "\$BACKUP_URL is empty"
  exit 1
fi

if [ -z "$BACKUP_USER" ]; then
  echo "\$BACKUP_USER is empty"
  exit 1
fi

if [ -z "$BACKUP_PASSWORD" ]; then
  echo "\$BACkUP_PASSWORD is empty"
  exit 1
fi

if [ -z "$CLUSTER_NAME" ]; then
  echo "\$CLUSTER_NAME is empty"
  exit 1
fi

if [ -z "$HOST" ]; then
  echo "\$HOST is empty"
  exit 1
fi

if [ -z "$PORT" ]; then
  echo "\$PORT is empty"
  exit 1
fi
echo "Create Google Drive service-account.json file."
echo "${GDRIVE_SERVICE_ACCOUNT}" > /tmp/gdrive-service-account.json

echo "Create rclone.conf file."
cat <<EOF > /tmp/rclone.conf
[gd]
type = drive
scope = drive
service_account_file = /tmp/gdrive-service-account.json
client_id = ${GDRIVE_CLIENT_ID}
root_folder_id = ${GDRIVE_ROOT_FOLDER_ID}
impersonate = ${GDRIVE_IMPERSONATOR}
[s3]
type = s3
env_auth = false
provider = ${S3_PROVIDER:-"AWS"}
access_key_id = ${AWS_ACCESS_KEY_ID}
secret_access_key = ${AWS_SECRET_ACCESS_KEY:-$AWS_SECRET_KEY}
region = ${AWS_REGION:-"us-east-1"}
endpoint = ${S3_ENDPOINT}
acl = ${AWS_ACL}
storage_class = ${AWS_STORAGE_CLASS}
[gs]
type = google cloud storage
project_number = ${GCS_PROJECT_ID}
service_account_file = /tmp/google-credentials.json
object_acl = ${GCS_OBJECT_ACL}
bucket_acl = ${GCS_BUCKET_ACL}
location =  ${GCS_LOCATION}
storage_class = ${GCS_STORAGE_CLASS:-"MULTI_REGIONAL"}
[http]
type = http
url = ${HTTP_URL}
[azure]
type = azureblob
account = ${AZUREBLOB_ACCOUNT}
key = ${AZUREBLOB_KEY}
EOF

if [[ -n "${GCS_SERVICE_ACCOUNT_JSON_KEY:-}" ]]; then
    echo "Create google-credentials.json file."
    cat <<EOF > /tmp/google-credentials.json
    ${GCS_SERVICE_ACCOUNT_JSON_KEY}
EOF
else
    touch /tmp/google-credentials.json
fi



BACKUP_DIR=/tmp/backup/backup_$(date +%F_%T)
mkdir -p $BACKUP_DIR

#mariabackup --compress --backup -H $HOST  -P${PORT} --export -u${BACKUP_USER} -p${BACKUP_PASSWORD} --target-dir=${BACKUP_DIR}
mysqldump -  --lock-tables --all-databases --h $HOST -P${PORT} -u${BACKUP_USER} -p${BACKUP_PASSWORD} > $BACKUP_DIR/full.sql

BACKUP_PATH=/tmp/$CLUSTER_NAME-$(date +%F_T).tar.gz
tar -czvf $BACKUP_PATH $BACKUP_DIR

rclone  --config /tmp/rclone.conf copy $BACKUP_PATH $BACKUP_URL