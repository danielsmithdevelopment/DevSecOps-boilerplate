package main

import (
	"context"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"worker-cluster/internal/otelsetup"
)

var taskTotal = atomic.Int32{}

var httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests in seconds",
	Buckets: prometheus.DefBuckets,
}, []string{"path"})

func initialize() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	initialize()
	ctx := log.Logger.WithContext(context.Background())
	shutdownTracer, err := otelsetup.InitTracer(ctx, "worker-cluster")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize tracing")
	}
	defer func() { _ = shutdownTracer(ctx) }()

	log.Ctx(ctx).Info().Msg("Starting work")

	app := fiber.New()
	app.Use(otelsetup.FiberMiddleware("worker-cluster"))

	app.Get("/", func(c *fiber.Ctx) error {
		startTime := time.Now()
		taskTotal.Add(1)
		log.Ctx(ctx).Info().Msg("Hello, World! #" + strconv.Itoa(int(taskTotal.Load())))

		httpRequestDuration.WithLabelValues("/").Observe(time.Since(startTime).Seconds())

		return c.SendString("Hello, World!")
	})

	prometheusHandler := promhttp.Handler()

	app.Get("/metrics", func(c *fiber.Ctx) error {
		handler := adaptor.HTTPHandler(prometheusHandler)
		return handler(c)
	})

	log.Ctx(ctx).Fatal().Err(app.Listen(":4200")).Msg("")
}
