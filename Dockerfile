FROM czarsimon/godep:1.11.2-alpine3.8 as build

# Copy source
WORKDIR /go/src/directory
COPY . .

# Install dependencies
RUN dep ensure

# Build application
WORKDIR /go/src/directory/cmd
RUN go build

FROM alpine:3.8 as run
WORKDIR /opt/app
RUN mkdir /etc/mimir /etc/mimir/directory
COPY --from=build /go/src/directory/cmd/cmd directory
CMD ["./directory"]