FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y gcc git

RUN apt-get update && apt-get install -y wget && \
    wget https://mirrors.aliyun.com/golang/go1.21.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz && \
    apt-get update -y && apt-get install s3fs -y && \
    rm go1.21.3.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /build

COPY . /build

RUN go build -o ycsi ./cmd

ENTRYPOINT ["/ycsi"]