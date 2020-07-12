/*
Copyright © 2020 A. Jensen <jensen.aaro@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package gke_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ajjensen13/gke"
)

// ExampleAliveContext demonstrates how to use gke.Do() and
// gke.AfterAliveContext() together to coordinate a graceful
// shutdown
func ExampleAliveContext() {
	c := make(chan bool)
	fmt.Println("0 started")

	// Cleans up gracefully after aliveCtx is canceled
	gke.Go(func(ctx context.Context) error {
		fmt.Println("1 started")
		c <- false // send 1
		<-ctx.Done()
		fmt.Println("1 stopped: ready context canceled")
		return nil
	})

	<-c // receive 1

	// Cleans up gracefully after finishing work
	gke.Go(func(ctx context.Context) error {
		defer func() { c <- false }() // send 2
		fmt.Println("2 started")
		// Do work
		fmt.Println("2 stopped: work complete")
		return nil
	})

	<-c // receive 2

	// Returns error signalling the end of the ready phase
	gke.Go(func(ctx context.Context) error {
		fmt.Println("3 started")
		c <- false // send 3
		<-c        // receive 4
		fmt.Println("3 stopped: error")
		return errors.New("error")
	})

	<-c // receive 3

	fmt.Println("0 waiting")
	c <- false // send 4
	cleanupCtx := gke.AfterAliveContext(time.Second * 10)
	<-cleanupCtx.Done()
	fmt.Println("0 stopped: cleanup complete")

	// Output:
	// 0 started
	// 1 started
	// 2 started
	// 2 stopped: work complete
	// 3 started
	// 0 waiting
	// 3 stopped: error
	// 1 stopped: ready context canceled
	// 0 stopped: cleanup complete
}

// ExampleAliveContext_WithLogger demonstrates how to use gke.Do()
// gke.NewLogger, and gke.AfterAliveContext together.
func ExampleAliveContext_WithLogger() {
	lg, cleanup, err := gke.NewLogger(context.Background())
	if err != nil {
		panic(err)
	}
	defer cleanup()

	gke.LogEnv(lg)
	gke.LogMetadata(lg)

	gke.Go(func(aliveCtx context.Context) error {
		<-time.After(time.Second)
		return nil
	})

	cleanupCtx := gke.AfterAliveContext(time.Second * 10)
	<-cleanupCtx.Done()
}
