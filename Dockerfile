FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /llm-assessor

FROM gcr.io/distroless/base-debian12
COPY --from=builder /llm-assessor /llm-assessor
ENTRYPOINT ["/llm-assessor"]
