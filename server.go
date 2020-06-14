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
	"cloud.google.com/go/logging"
	"context"
	"github.com/google/uuid"
	"net"
	"net/http"
	"time"
)

func provideServer(lg Logger, handler http.Handler) *http.Server {
	result := http.Server{
		Handler:           handler,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Second * 60,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		ErrorLog:          lg.StandardLogger(logging.Error),
		BaseContext: func(_ net.Listener) (ctx context.Context) {
			ctx, _ = AliveContext()
			return
		},
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, RequestContextKey, uuid.New().String())
		},
	}
	go func() {
		alive, _ := AliveContext()
		<-alive.Done()
		_ = result.Shutdown(context.Background())
	}()
	return &result
}

type requestContextKey string

const RequestContextKey = requestContextKey(`gkeRequestContextKey`)
