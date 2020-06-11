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
	"os"
	"os/signal"
)

var (
	pkgAlive       context.Context
	pkgAliveCancel context.CancelFunc
)

func init() {
	pkgAlive, pkgAliveCancel = context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		signal.Stop(c)

		client, cleanup, err := NewLogClient(context.Background())
		if err != nil {
			panic(err)
		}
		defer cleanup()

		logger := client.Logger(os.Args[0] + "-signal-handler")
		logger.Noticef("signal received: %v", s)

		pkgAliveCancel()
	}()
}

// Alive returns a context that is used to communicate a shutdown to various parts of an application.
func Alive() (context.Context, context.CancelFunc) {
	return pkgAlive, pkgAliveCancel
}
