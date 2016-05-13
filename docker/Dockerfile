FROM ubuntu:wily
MAINTAINER caktux

ENV DEBIAN_FRONTEND noninteractive

# Usual update / upgrade
RUN apt-get update
RUN apt-get upgrade -q -y
RUN apt-get dist-upgrade -q -y

# Install UR
RUN apt-get install -q -y software-properties-common
RUN add-apt-repository ppa:ur/ur
RUN add-apt-repository ppa:ur/ur-dev
RUN apt-get update
RUN apt-get install -q -y gur

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/usr/bin/gur"]
