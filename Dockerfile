FROM ubuntu:latest
LABEL authors="vacheslavterentev"

ENTRYPOINT ["top", "-b"]