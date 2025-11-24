# syntax=docker/dockerfile:1

FROM golang:alpine3.19

LABEL org.opencontainers.image.authors="Josh Carlson <837837+magneticstain@users.noreply.github.com>"

WORKDIR /app

ENV SVC_ACCT_USER_NAME=ip2cr
ENV SVC_ACCT_USER_ID=1001
ENV SVC_ACCT_GROUP_NAME=ip2cr
ENV SVC_ACCT_GROUP_ID=1001

RUN addgroup -g $SVC_ACCT_GROUP_ID $SVC_ACCT_GROUP_NAME && \
    adduser --shell /sbin/nologin --disabled-password --no-create-home \
    --uid $SVC_ACCT_USER_ID --ingroup $SVC_ACCT_GROUP_NAME $SVC_ACCT_USER_NAME

RUN chown -R $SVC_ACCT_USER_NAME:$SVC_ACCT_GROUP_NAME /app

USER $SVC_ACCT_USER_NAME:$SVC_ACCT_GROUP_NAME

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o ./ip2cr

ENTRYPOINT [ "./ip2cr" ]
CMD [ "--help" ]
