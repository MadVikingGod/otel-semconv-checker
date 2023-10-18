FROM golang:alpine as builder

WORKDIR /app

COPY . ./

RUN go build -v -o semantic_convention_checker ./cmd

FROM alpine

COPY --from=builder /app/semantic_convention_checker /app/semantic_convention_checker
WORKDIR /app
CMD ["/app/semantic_convention_checker"]
