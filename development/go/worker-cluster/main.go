package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var taskCount = atomic.Int32{}
var taskTotal = atomic.Int32{}
var wg = sync.WaitGroup{}

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

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx := log.Logger.WithContext(context.Background())
	log.Ctx(ctx).Info().Msg("Starting work")
	work(ctx, []func(){
		func() {
			fmt.Println("Hello, World 1!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 2!")
			wg.Done()
		},
		func() {
			time.Sleep(1 * time.Second)
			fmt.Println("Hello, World 3!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 4!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 5!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 6!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 7!")
			wg.Done()
		},
	})

	work(ctx, []func(){
		func() {
			fmt.Println("Hello, World 8!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 9!")
			wg.Done()
		},
		func() {
			time.Sleep(1 * time.Second)
			fmt.Println("Hello, World 10!")
			wg.Done()
		},
	})

	work(ctx, []func(){
		func() {
			fmt.Println("Hello, World 11!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 12!")
			wg.Done()
		},
		func() {
			time.Sleep(1 * time.Second)
			fmt.Println("Hello, World 13!")
			wg.Done()
		},
		func() {
			fmt.Println("Hello, World 14!")
			wg.Done()
		},
	})
}
