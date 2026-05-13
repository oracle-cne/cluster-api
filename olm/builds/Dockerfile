FROM ghcr.io/oracle/oraclelinux:8-slim

ARG managerImage=/cluster-api/manager
ARG RPM

COPY rpms/${RPM} /tmp/

RUN microdnf update && \
    microdnf clean all && \
    rpm -i /tmp/*.rpm && \
    mv ${managerImage} /manager && \
    rm /tmp/*.rpm && \
    groupadd -r ocne && \
    useradd --no-log-init -r -m -d /ocne -g ocne -u 65532 ocne && \
    mkdir -p /home/ocne && \
    chown -R 65532:ocne /manager /home/ocne && \
    chmod 500 /manager

USER 65532
ENTRYPOINT ["/manager"]
