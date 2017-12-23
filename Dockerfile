# build stage
FROM golang:1.9 AS build-env
ADD . /src
#disable crosscompiling
ENV CGO_ENABLED=0

#compile linux only
ENV GOOS=linux
RUN cd /src && go get -v -d && go build -ldflags '-w -s' -a -installsuffix cgo -o goapp

# final stage
FROM scratch
COPY --from=build-env /src/goapp /app/
ENTRYPOINT /app/goapp