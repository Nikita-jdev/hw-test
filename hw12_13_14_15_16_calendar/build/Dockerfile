# --- Сборка ---
FROM golang:1.22 AS build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CMD_PATH=./cmd/calendar
ARG LDFLAGS

RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o /out/app ${CMD_PATH}

# --- Рантайм ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

ENV BIN_FILE=/opt/calendar/app
COPY --from=build /out/app ${BIN_FILE}

ENV CONFIG_FILE=/etc/calendar/config.toml
COPY deployments/config.docker.toml ${CONFIG_FILE}

EXPOSE 8080
CMD ["/opt/calendar/app", "-config", "/etc/calendar/config.toml"]
