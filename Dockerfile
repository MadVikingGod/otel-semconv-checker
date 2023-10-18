FROM golang:alpine as builder

WORKDIR /app

# Download and cache dependancies
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd
COPY pkg/ ./pkg
RUN go build -v -o semantic_convention_checker ./cmd

FROM alpine

COPY --from=builder /app/semantic_convention_checker /app/semantic_convention_checker
COPY ./config.yaml /app/config.yaml
WORKDIR /app
CMD ["/app/semantic_convention_checker"]
