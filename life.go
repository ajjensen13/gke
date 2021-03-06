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

package gke

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

var (
	pkgAliveOnce     sync.Once // protects below
	pkgAlive         context.Context
	pkgAliveCancel   context.CancelFunc
	pkgErrGroup      *errgroup.Group
	pkgErrGroupCtx   context.Context
	pkgSyncWaitGroup sync.WaitGroup
)

func initAlive() {
	pkgAlive, pkgAliveCancel = context.WithCancel(context.Background())
	pkgErrGroup, pkgErrGroupCtx = errgroup.WithContext(pkgAlive)

	go func() {
		defer pkgAliveCancel()

		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case s := <-c:
			log.Printf("gke: signal recieved: %v", s)
		case <-pkgErrGroupCtx.Done():
		}

		signal.Stop(c)
	}()
}

// Go kicks off a function that will run while the application is alive. It is passed
// the AliveContext() context as a parameter. It should shutdown once the alive
// context has been canceled. If f returns a non-nil error, then the alive context
// will be canceled and other functions started via Go() will begin to shutdown.
func Go(f func(aliveCtx context.Context) error) {
	pkgAliveOnce.Do(initAlive)
	pkgSyncWaitGroup.Add(1)
	pkgErrGroup.Go(func() error {
		defer pkgSyncWaitGroup.Done()
		return f(pkgAlive)
	})
}

// Wait blocks until all function calls from the Go() function have returned, then
// returns the first non-nil error (if any) from them.
func Wait() error {
	pkgAliveOnce.Do(initAlive)
	return pkgErrGroup.Wait()
}

// AliveContext returns a context that is used to communicate a
// shutdown to various parts of an application.
func AliveContext() (context.Context, context.CancelFunc) {
	pkgAliveOnce.Do(initAlive)
	return pkgAlive, pkgAliveCancel
}

// AfterAliveContext returns a context that completes when the alive
// context has been canceled and all functions that were started by
// calling Go() have returned (or the timeout expires).
//
//		// Note: currently this will always be true
// 		errors.Is(AfterAliveContext(timeout).Err(), context.Canceled)
func AfterAliveContext(timeout time.Duration) context.Context {
	pkgAliveOnce.Do(initAlive)
	result, cancelFunc := context.WithCancel(context.Background())

	wf := int32(0)
	go func() {
		defer cancelFunc()
		<-pkgAlive.Done()
		<-time.After(timeout)
		if atomic.LoadInt32(&wf) == 0 {
			panic("program failed to shutdown gracefully")
		}
	}()

	go func() {
		defer cancelFunc()
		pkgSyncWaitGroup.Wait()
		atomic.StoreInt32(&wf, 1)
	}()

	return result
}
