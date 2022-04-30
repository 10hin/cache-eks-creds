package main

import (
	"context"
	"fmt"
	"github.com/10hin/cache-eks-creds/cmd"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 0)
	signal.Notify(sigChan, syscall.SIGINT)

	complete := make(chan struct{}, 0)
	go func(ctx1 context.Context, complete1 chan<- struct{}) {
		defer func() {
			close(complete1)
		}()
		err1 := cmd.ExecuteContext(ctx1)
		if err1 != nil {
			panic(err1)
		}
	}(ctx, complete)
	select {
	case <-sigChan:
		fmt.Println("SIGINT received!")
		panic(fmt.Errorf("interrpted"))
	case <-complete:
	}
}
