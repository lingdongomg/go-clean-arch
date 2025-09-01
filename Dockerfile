# Builder
FROM golang:1.20.7-alpine3.17 as builder

RUN apk update && apk upgrade && \
    apk --update add git make bash build-base

WORKDIR /app
COPY . .
RUN make build

# Distribution
FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    mkdir /app 

WORKDIR /app 

ENV TZ=Asia/Shanghai
EXPOSE 9090

COPY --from=builder /app/engine /app/

CMD /app/engine



FROM tlinux/tlinux2.2:v1.2
MAINTAINER kaichaohu<kaichaohu@tencent.com>

ARG svc_env

ENV ProjectName_SVC_ENV $svc_env
ENV ProjectName_SVC_PATH_PREFIX /data/projectname
ENV ProjectName_SVC_BINARY_PATH $ProjectName_SVC_PATH_PREFIX/bin
ENV ProjectName_SVC_LOGS_PATH $ProjectName_SVC_PATH_PREFIX/logs
ENV ProjectName_SVC_CONF_PATH $ProjectName_SVC_PATH_PREFIX/conf

RUN mkdir -p $ProjectName_SVC_BINARY_PATH
RUN mkdir -p $ProjectName_SVC_LOGS_PATH
RUN mkdir -p $ProjectName_SVC_CONF_PATH
RUN chmod 777 $ProjectName_SVC_LOGS_PATH
RUN chmod 777 $ProjectName_SVC_BINARY_PATH
RUN chmod 777 $ProjectName_SVC_CONF_PATH

COPY ./src/projectname $ProjectName_SVC_BINARY_PATH/
WORKDIR $ProjectName_SVC_BINARY_PATH
ENTRYPOINT ./projectname -c /data/projectname/conf/config.yaml