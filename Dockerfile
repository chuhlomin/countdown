FROM golang:1.23 AS builder
WORKDIR /countdown
COPY . .

# download fonts
RUN apt-get update && apt-get install -y unzip && \
    mkdir /countdown/cmd/server/fonts && \
    mkdir /tmp/fonts && \
    # Gidole
    wget https://github.com/larsenwork/Gidole/raw/refs/heads/master/gidole.zip -O /tmp/fonts/gidole.zip && \
    unzip /tmp/fonts/gidole.zip -d /tmp/fonts/gidole && \
    mv /tmp/fonts/gidole/GidoleFont/*.ttf /countdown/cmd/server/fonts && \
    mv /tmp/fonts/gidole/GidoleFont/*.otf /countdown/cmd/server/fonts && \
    # Overpass
    wget https://github.com/RedHatOfficial/Overpass/releases/download/v3.0.5/overpass-3.0.5.zip -O /tmp/fonts/overpass.zip && \
    unzip /tmp/fonts/overpass.zip -d /tmp/fonts/overpass && \
    mv /tmp/fonts/overpass/Overpass-3.0.5/desktop-fonts/overpass-mono /countdown/cmd/server/fonts && \
    rm -rf /tmp/fonts

ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM
ARG TARGETPLATFORM

# build server
RUN echo "Building for ${TARGETOS}/${TARGETARCH}; build platform ${BUILDPLATFORM}; target platform ${TARGETPLATFORM}" && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -mod=vendor -buildvcs \
    -o /countdown/server /countdown/cmd/server

FROM scratch
COPY --from=builder /countdown/server /server
ENTRYPOINT ["/server"]
