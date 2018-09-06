FROM        quay.io/prometheus/busybox:glibc
MAINTAINER  The Presto_exporter Authors <presto-pj@mail.yahoo.co.jp>

COPY presto_exporter /bin/presto_exporter

EXPOSE     9483
ENTRYPOINT ["/bin/presto_exporter"]
