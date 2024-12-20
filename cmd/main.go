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
			time.Sleep(1 * time.Second)
			fmt.Println("interrupt gorutine 1")
			cancel()
		},
		nil,
	)
	start2 := g.AddAfter(
		func(started *apprun.StartSignal) error {
			time.Sleep(1 * time.Second)
			fmt.Println("start gorutine 2")
			time.Sleep(1 * time.Second)

			started.Success()

			<-ctx.Done()
			return nil
		},
		func(_ error) {
			time.Sleep(1 * time.Second)
			fmt.Println("interrupt gorutine 2")
			cancel()
		},
		start1,
	)

	start2.AferStart = func() {
		fmt.Println("application started")
	}

	err := g.RunApp(ctx)
	if err != nil {
		fmt.Println("application start error:", err)
	}

	fmt.Println("application stopped")
}
