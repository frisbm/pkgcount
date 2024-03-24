package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/frisbm/pkgcount/internal/pkgcount"

	"github.com/frisbm/pkgcount/internal/models"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var args models.Args

	flag.BoolVar(&args.Help, "h", false, "Display help messages with argument list and descriptions.")
	flag.BoolVar(&args.Unrendered, "u", false, "Retrieve markdown in unrendered format.")
	flag.StringVar(&args.Out, "o", "", "Save output to file. Used with -u to preserve format.")
	flag.StringVar(&args.Dir, "d", ".", "Specify directory or file path. Default is current directory.")
	flag.StringVar(&args.Exclude, "exclude", "", "Exclude specific files, directories, or other entities with regex.")
	flag.IntVar(&args.Lte, "lte", math.MaxInt, "Display package counts less than or equal to specified integer.")
	flag.IntVar(&args.Gte, "gte", 0, "Display package counts greater than or equal to specified integer.")

	flag.Parse()
	if args.Help {
		flag.PrintDefaults()
		return
	}

	// Handle signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		select {
		case <-signalChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	err := pkgcount.Run(ctx, args)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to run pkgcount: %w", err))
	}
}
