FROM golang:1.25-alpine AS build

ENV GOTOOLCHAIN=auto

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /build/app app
COPY configs configs
USER nonroot:nonroot
ENTRYPOINT ["./app"]
