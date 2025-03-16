FROM golang:1.23.5 as build
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o app .

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /build/app app
COPY configs configs
ENTRYPOINT ["./app"]