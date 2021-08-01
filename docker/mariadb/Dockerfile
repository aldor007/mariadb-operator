FROM golang:1.16-alpine AS peer-finder-build
LABEL stage=gobuilder
LABEL app=peer-finder
RUN apk add --no-cache git
ARG TARGETARCH amd64
WORKDIR /go/src/
ADD peer-finder/ /go/src/apps/peer-finder
RUN  cd /go/src/apps/peer-finder \
  &&  go get -d -v ./... \
  && CGO_ENABLED=0 go build -ldflags '-w  -extldflags "-static"' -v  -o peer-finder ./peer-finder.go

FROM ubuntu:20.04
ARG DB_VER
ENV DEBIAN_FRONTEND noninteractive
COPY --from=peer-finder-build /go/src/apps/peer-finder/peer-finder /usr/bin/peer-finder
LABEL io.k8s.description="MariaDB is a multi-user, multi-threaded SQL database server" \
      io.k8s.display-name="MariaDB ${DB_VER}" \
      io.openshift.expose-services="3306:mysql" \
      io.openshift.tags="database,mysql,mariadb10,rh-mariadb10"
EXPOSE 3306/tcp

# NOTES:
# galera-4 is installing with mariadb-server as a dependency
# mariadb-backup is in ise since MariaDB 10.3
# in earlier versions use percona-xtrabackup and change galera.cnf accordingly
RUN apt update \
  && apt install -y -qq \
    software-properties-common \
  && apt-key adv --fetch-keys 'https://mariadb.org/mariadb_release_signing_key.asc' \
  && add-apt-repository 'deb [arch=amd64,arm64,ppc64el] https://ftp.icm.edu.pl/pub/unix/database/mariadb/repo/10.5/ubuntu focal main' \
  && apt update \
  && apt install -y -qq \
    dnsutils \
    mariadb-server \
    mariadb-backup \
    rsync \
    rclone \
    socat \
    tzdata \
  && apt clean

RUN echo '!include /etc/config/my_extra.cnf' >> /etc/mysql/my.cnf

RUN rm /etc/mysql/mariadb.conf.d/50-server.cnf \
    && chmod -R  0444 /etc/mysql/conf.d/ && chmod a+x /usr/bin/container-entrypoint.sh

COPY root /
RUN bash /usr/libexec/container-setup.sh

ENTRYPOINT ["/usr/bin/container-entrypoint.sh"]