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

package log

import (
	"cloud.google.com/go/logging"
	"log"
)

// Client is used to provision new loggers and close underlying connections during shutdown.
type Client interface {
	// Logger returns a logger with a provided logID.
	Logger(logID string) Logger
	// Close waits for all opened loggers to be flushed and closes the client.
	Close() error
}

// Logger logs entries to a single log.
type Logger interface {
	// StandardLogger returns a *log.Logger for a given severity.
	StandardLogger(severity logging.Severity) *log.Logger
	// Log queues a single log entry.
	Log(entry logging.Entry)
	// Flush flushes the queued log entries.
	Flush() error
}
