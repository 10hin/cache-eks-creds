package main

import (
	"context"
	"fmt"
	"github.com/10hin/cache-eks-creds/cmd"
	"github.com/10hin/cache-eks-creds/pkg/cache"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	emptyCtx := context.Background()
	ctxWithComponents := storeComponents(emptyCtx, map[string]interface{}{
		"github.com/10hin/cache-eks-creds/pkg/cache.CredentialCache": cache.NewFileCache(),
	})

	ctx, cancel := context.WithCancel(ctxWithComponents)
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

func storeComponents(ctx context.Context, components map[string]interface{}) context.Context {
	var newCtx context.Context = nil
	for key, comp := range components {
		if newCtx == nil {
			newCtx = context.WithValue(ctx, key, comp)
		} else {
			newCtx = context.WithValue(newCtx, key, comp)
		}
	}
	return newCtx
}
