FROM golang:1.25.11-alpine AS build
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o app .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /build/app app
COPY configs configs
USER nonroot:nonroot
ENTRYPOINT ["./app"]
