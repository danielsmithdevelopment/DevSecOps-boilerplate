package main

import (
	"context"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var taskCount = atomic.Int32{}
var taskTotal = atomic.Int32{}
var wg = sync.WaitGroup{}

// Create a histogram metric for HTTP request durations
var httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests in seconds",
	Buckets: prometheus.DefBuckets,
}, []string{"path"})

func work(ctx context.Context, tasks []func()) {
	log.Ctx(ctx).Info().Msg("Adding " + strconv.Itoa(len(tasks)) + " tasks to wait group")
	taskCount.Add(int32(len(tasks)))
	wg.Add(len(tasks))
	for _, task := range tasks {
		task()
		taskCount.Add(-1)
		taskTotal.Add(1)
		log.Ctx(ctx).Info().Msg("Task completed. " +
			strconv.Itoa(int(taskCount.Load())) + " tasks remaining. " +
			strconv.Itoa(int(taskTotal.Load())) + " tasks total.")
	}
	wg.Wait()
}

func initialize() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	initialize()
	ctx := log.Logger.WithContext(context.Background())
	log.Ctx(ctx).Info().Msg("Starting work")

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		startTime := time.Now()
		taskTotal.Add(1)
		log.Ctx(ctx).Info().Msg("Hello, World! #" + strconv.Itoa(int(taskTotal.Load())))

		// Record the request duration in the histogram
		httpRequestDuration.WithLabelValues("/").Observe(time.Since(startTime).Seconds())

		return c.SendString("Hello, World!")
	})

	// Create HTTP handler for Prometheus metrics
	prometheusHandler := promhttp.Handler()

	// Adapt the Prometheus handler for Fiber
	app.Get("/metrics", func(c *fiber.Ctx) error {
		handler := adaptor.HTTPHandler(prometheusHandler)
		return handler(c)
	})

	log.Ctx(ctx).Fatal().Err(app.Listen(":4200")).Msg("")

	// for i := 0; i < 600; i++ {
	// 	work(ctx, []func(){
	// 		func() {
	// 			fmt.Println("Hello, World 1!")
	// 			wg.Done()
	// 		},
	// 		func() {
	// 			fmt.Println("Hello, World 2!")
	// 			wg.Done()
	// 		},
	// 		func() {
	// 			time.Sleep(1 * time.Second)
	// 			fmt.Println("Hello, World 3!")
	// 			wg.Done()
	// 		},
	// 		func() {
	// 			fmt.Println("Hello, World 4!")
	// 			wg.Done()
	// 		},
	// 		func() {
	// 			fmt.Println("Hello, World 5!")
	// 			wg.Done()
	// 		},
	// 		func() {
	// 			fmt.Println("Hello, World 6!")
	// 			wg.Done()
	// 		},
	// 		func() {
	// 			fmt.Println("Hello, World 7!")
	// 			wg.Done()
	// 		},
	// 	})
	// }

}
