package main

import (
	"context"
	"fmt"
	"log/slog"
	apprun "run-group-test/pkg/app_run"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.Default()

	g := apprun.New(log)

	start1 := g.AddAfter(
		func(started *apprun.StartSignal) error {
			fmt.Println("start gorutine 1")
			time.Sleep(1 * time.Second)

			started.Success()

			<-ctx.Done()
			return nil
		},
		func(_ error) {
			fmt.Println("interrupt gorutine 1")
			cancel()
		},
		nil,
	)
	start2 := g.AddAfter(
		func(started *apprun.StartSignal) error {
			fmt.Println("start gorutine 2")
			time.Sleep(1 * time.Second)

			started.Success()

			<-ctx.Done()
			return nil
		},
		func(_ error) {
			fmt.Println("interrupt gorutine 2")
			cancel()
		},
		start1,
	)

	start2.AferStartFunc = func() {
		fmt.Println("application started")
	}

	g.Run(ctx)

	fmt.Println("application stopped")
}
