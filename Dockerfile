# build stage
FROM golang:1.9 AS build-env
ADD . /src
RUN cd /src && go get -v -d && go build -o goapp

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp