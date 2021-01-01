# The first stage copies the source code and builds the zag application
FROM golang:1.15-alpine AS build
RUN apk add --update \
    git \
    gcc \
    libc-dev
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
# Copy source code
COPY . /go/src/github.com/mewil/zag
# Download dependencies and build the application
WORKDIR /go/src/github.com/mewil/zag
RUN go mod download
RUN go install .
RUN adduser -D -g '' user

# The second stage uses a scratch image to reduce the image size and improve security
FROM scratch AS zag
LABEL Author="Michael Wilson"
# Copy the statically compiled Go binary and use it as our entrypoint
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/zag /bin/zag
USER user

ENTRYPOINT ["/bin/zag"]
EXPOSE 6464