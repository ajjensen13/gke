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
	"fmt"
	"io"
	"log"
	"strings"
)

// MultiClient wraps multiple clients. Each operation on a MultiClient
// is executed on each of its underlying clients.
type MultiClient []Client

func (mc MultiClient) Close() error {
	var es errs
	for _, l := range mc {
		err := l.Close()
		if err != nil {
			es = append(es, err)
		}
	}

	if len(es) > 0 {
		return fmt.Errorf("1 or more errors while closing log client: %w", es)
	}

	return nil
}

func (mc MultiClient) Logger(logID string) Logger {
	result := MultiLogger{make([]Logger, 0, len(mc))}
	for _, c := range mc {
		result.ls = append(result.ls, c.Logger(logID))
	}
	return result
}

// MultiLogger wraps multiple loggers. Each operation on a MultiLogger
// is executed on each of its underlying loggers.
type MultiLogger struct {
	ls []Logger
}

func (m MultiLogger) StandardLogger(severity logging.Severity) *log.Logger {
	writers := make([]io.Writer, 0, len(m.ls))
	for _, l := range m.ls {
		writers = append(writers, l.StandardLogger(severity).Writer())
	}
	return log.New(io.MultiWriter(writers...), severity.String(), log.LstdFlags)
}

func (m MultiLogger) Log(entry logging.Entry) {
	for _, l := range m.ls {
		l.Log(entry)
	}
}

func (m MultiLogger) Flush() error {
	var es errs
	for _, l := range m.ls {
		err := l.Flush()
		if err != nil {
			es = append(es, err)
		}
	}

	if len(es) > 0 {
		return fmt.Errorf("1 or more errors while flushing logger: %w", es)
	}

	return nil
}

type errs []error

func (e errs) Error() string {
	var builder strings.Builder
	for i, err := range e {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(err.Error())
	}
	return builder.String()
}
