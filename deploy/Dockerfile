FROM golang:1.24-alpine AS build

RUN apk add --no-cache git sed

# Clone Aperture at a pinned tag for reproducible builds.
WORKDIR /src
RUN git clone --depth 1 --branch v0.3.6-beta https://github.com/lightninglabs/aperture.git .

# Copy in the Phoenixd challenger.
COPY challenger-phoenixd.go /src/challenger/phoenixd.go

# Patch config.go: add PhoenixdURL and PhoenixdPassword to AuthConfig.
RUN sed -i '/DevServer bool.*devserver/a\\n\t\/\/ PhoenixdURL is the URL of a Phoenixd instance (e.g. http:\/\/localhost:9740).\n\t\/\/ When set, Phoenixd is used for invoice creation instead of LND.\n\tPhoenixdURL string `long:"phoenixdurl" description:"URL of a Phoenixd instance"`\n\n\tPhoenixdPassword string `long:"phoenixdpassword" description:"Phoenixd HTTP API password"`' config.go \
    && grep -q 'PhoenixdURL' config.go

# Patch config.go validate(): add PhoenixdURL case before the default.
RUN sed -i '/case a.Passphrase != ""/i\\tcase a.PhoenixdURL != "":\n\t\tlog.Info("Using Phoenixd configuration")\n\t\treturn nil\n' config.go \
    && grep -q 'Using Phoenixd configuration' config.go

# Patch aperture.go: add Phoenixd case in the challenger switch.
RUN sed -i '/case authCfg.LndHost != ""/i\\t\tcase authCfg.PhoenixdURL != "":\n\t\t\tlog.Infof("Using Phoenixd authenticator")\n\n\t\t\ta.challenger, err = challenger.NewPhoenixdChallenger(\n\t\t\t\tauthCfg.PhoenixdURL,\n\t\t\t\tauthCfg.PhoenixdPassword,\n\t\t\t)\n\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n' aperture.go \
    && grep -q 'NewPhoenixdChallenger' aperture.go

# Build.
RUN go build -o /aperture ./cmd/aperture

# Build echo server.
COPY cmd/echo-server/main.go /echo-server/main.go
RUN cd /echo-server && go mod init echo-server && go build -o /echo-server/echo-server .

# --- Runtime ---
FROM alpine:3.21
RUN apk add --no-cache ca-certificates gettext
RUN adduser -D -u 1000 aperture \
    && mkdir -p /etc/aperture /home/aperture/.aperture \
    && chown -R 1000:1000 /etc/aperture /home/aperture
ENV HOME=/home/aperture
USER aperture

COPY --from=build /aperture /usr/local/bin/aperture
COPY --from=build /echo-server/echo-server /usr/local/bin/echo-server

EXPOSE 8080 8081
