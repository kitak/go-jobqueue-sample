package main

import "os"
import "os/signal"
import "fmt"

func main() {
	done := make(chan struct{})
	quit := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func(quit chan struct{}, done chan struct{}) {
		for {
			select {
			case _ = <-quit:
				done <- struct{}{}
			}
		}
	}(quit, done)

	go func() {
		for {
			s := <-sig
			fmt.Println(s)
			done <- struct{}{}
		}
	}()

	<-done
	fmt.Println("done")
}
