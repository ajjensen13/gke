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
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	pkgAlive         context.Context
	pkgAliveCancel   context.CancelFunc
	pkgErrGroup      *errgroup.Group
	pkgErrGroupCtx   context.Context
	pkgSyncWaitGroup sync.WaitGroup
)

func init() {
	pkgAlive, pkgAliveCancel = context.WithCancel(context.Background())
	pkgErrGroup, pkgErrGroupCtx = errgroup.WithContext(pkgAlive)

	go func() {
		defer pkgAliveCancel()

		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case <-c:
		case <-pkgErrGroupCtx.Done():
		}

		signal.Stop(c)
	}()
}

func Do(f func(context.Context) error) {
	pkgSyncWaitGroup.Add(1)
	pkgErrGroup.Go(func() error {
		defer pkgSyncWaitGroup.Done()
		return f(pkgAlive)
	})
}

// AliveContext returns a context that is used to communicate a shutdown to various parts of an application.
func AliveContext() (context.Context, context.CancelFunc) {
	return pkgAlive, pkgAliveCancel
}

func WaitForCleanup(timeout time.Duration) error {
	done := make(chan error)
	go func() {
		defer close(done)
		<-pkgAlive.Done()
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		<-ctx.Done()
		done <- ctx.Err()
	}()

	go func() {
		defer close(done)
		pkgSyncWaitGroup.Wait()
	}()

	return <-done
}
