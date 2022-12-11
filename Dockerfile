ARG alpine_version=3.16
ARG go_version=1.18.2
ARG node_version=14.15.3

# Server binary builder
FROM golang:${go_version}-alpine${alpine_version} as server_builder
ARG go_swagger_version=0.27.0

ARG SSH_PRIVATE_KEY
ARG KEN_CRT

RUN apk add --no-cache git
RUN apk add --update --no-cache curl build-base
RUN curl -sSLo /usr/local/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/v${go_swagger_version}/swagger_linux_amd64
RUN chmod +x /usr/local/bin/swagger

RUN git config --global "url.ssh://git@gitlab.kenda.com.tw:4222".insteadOf "https://gitlab.kenda.com.tw"

RUN apk update && apk add openssh

RUN mkdir ~/.ssh && \
    echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa && \
    chmod 600 ~/.ssh/id_rsa && \
    ssh-keyscan -Ht ecdsa -p 4222 gitlab.kenda.com.tw,192.1.1.159 >> ~/.ssh/known_hosts

RUN echo "${KEN_CRT}" >> /etc/ssl/certs/ca-certificates.crt

ENV REPO_DIR ${GOPATH}/src/gitlab.com.kenda.com.tw/kenda/mui
ENV SERVER_DIR ${REPO_DIR}/server
ENV APP_NAME mui

COPY ./swagger.yml ${REPO_DIR}/
COPY ./assets/mesage/openapi.yaml ${REPO_DIR}/assets/mesage/
COPY ./assets/mes/services.swagger.json ${REPO_DIR}/assets/mes/

COPY ./server ${SERVER_DIR}/

WORKDIR ${SERVER_DIR}

ENV GOPRIVATE *.kenda.com.tw
RUN go generate .
RUN go mod download
RUN go vet ./...
RUN go test -race -coverprofile .testCoverage.txt ./...
RUN go tool cover -func .testCoverage.txt
RUN go build -race -ldflags "-extldflags '-static'" -o /opt/mui/server ./swagger/cmd/${APP_NAME}-server

CMD ["/bin/sh"]

# UI dist builder
FROM node:${node_version}-alpine3.12 as ui_builder

ENV UI_DIR /app
COPY ./ui ${UI_DIR}

WORKDIR ${UI_DIR}

RUN rm -rf node_modules
RUN npm install yarn
RUN yarn install

# Add lint, unit test, etc.
RUN yarn build

CMD ["/bin/sh"]

# Deployable with server binary and UI dist
FROM alpine:${alpine_version}

WORKDIR /root/
COPY --from=server_builder /opt/mui/server /root/server
COPY --from=ui_builder /app/dist /root/ui
