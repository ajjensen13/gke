// +build wireinject

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
	"github.com/google/wire"
	"net/http"
)

// NewServer returns a new server with settings defaulted for use in GKE. The server
// is initialized with sensible defaults for timeout values. It sets the base context
// to AliveContext(). It starts a go routine to call Shutdown() when the AliveContext()
// is canceled. It sets up a ConnContext function to initialize the RequestContextKey data.
func NewServer(ctx context.Context, handler http.Handler, lg Logger) (*http.Server, error) {
	panic(wire.Build(provideServer))
}
