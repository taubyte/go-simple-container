FROM debian:stable-slim

RUN apt-get clean && apt-get update && \
    apt-get install -y wget gcc gcc-avr avr-libc

ENV GO_RELEASE=1.17.6
RUN wget https://dl.google.com/go/go${GO_RELEASE}.linux-amd64.tar.gz && \
	tar xfv go${GO_RELEASE}.linux-amd64.tar.gz -C /usr/local && \
	rm go${GO_RELEASE}.linux-amd64.tar.gz && \
	find /usr/local/go -mindepth 1 -maxdepth 1 ! -name 'src' ! -name 'VERSION' ! -name 'bin' ! -name 'pkg' -exec rm -rf {} +
ENV PATH=${PATH}:/usr/local/go/bin

RUN apt-get remove -y wget && \
    apt-get autoremove -y && \
    apt-get clean

RUN apt-get install -y bash

RUN apt-get install -y wabt

ENV PATH=${PATH}:/tools:/usr/bin:/bin

RUN echo $PATH

ENV GOPATH=/go

RUN mkdir /src /out

VOLUME /src
VOLUME /out
