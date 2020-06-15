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
)

// NewLogger is a convenience function for providing a default logger. It creates
// a new client, then creates a new logger with DefaultLogID.
// Note: ctx should usually be context.Background() to ensure that the logging
// events occur event after AliveContext() is canceled.
func NewLogger(ctx context.Context) (lg Logger, cleanup func(), err error) {
	panic(wire.Build(NewLogClient, provideDefaultLogger, DefaultLogID))
}

func provideDefaultLogger(client LogClient, logId string) Logger {
	return client.Logger(logId)
}
