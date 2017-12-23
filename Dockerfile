# build stage
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go get -v -d && go build -o goapp

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp