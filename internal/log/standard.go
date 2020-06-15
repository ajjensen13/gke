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
)

// StandardClient wraps a standard logger. See *log.Logger
type StandardClient struct {
	writer io.Writer
}

// NewStandardClient returns a new StandardClient that writes to writer.
func NewStandardClient(writer io.Writer) Client {
	return StandardClient{writer}
}

// Logger returns a logger with a provided logID.
func (s StandardClient) Logger(logID string) Logger {
	return newStdLogger(s.writer, logID)
}

// Close implements log.Client.Close().
// For StandardClient, it is a no-op.
func (s StandardClient) Close() error {
	return nil // no-op
}

type standardLogger struct {
	logId      string
	bySeverity map[logging.Severity]*log.Logger
}

func newStdLogger(writer io.Writer, logId string) *standardLogger {
	result := standardLogger{
		logId:      logId,
		bySeverity: make(map[logging.Severity]*log.Logger, 8),
	}

	result.setLoggerForSeverity(writer, logging.Default, logId)
	result.setLoggerForSeverity(writer, logging.Debug, logId)
	result.setLoggerForSeverity(writer, logging.Info, logId)
	result.setLoggerForSeverity(writer, logging.Notice, logId)
	result.setLoggerForSeverity(writer, logging.Warning, logId)
	result.setLoggerForSeverity(writer, logging.Error, logId)
	result.setLoggerForSeverity(writer, logging.Alert, logId)
	result.setLoggerForSeverity(writer, logging.Critical, logId)
	result.setLoggerForSeverity(writer, logging.Emergency, logId)

	return &result
}

func (s *standardLogger) setLoggerForSeverity(writer io.Writer, severity logging.Severity, logId string) {
	const flags = log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix
	s.bySeverity[severity] = log.New(writer, fmt.Sprintf("%s %s ", logId, severity.String()), flags)
}

// StandardLogger implements log.Logger.StandardLogger().
func (s *standardLogger) StandardLogger(severity logging.Severity) *log.Logger {
	return s.bySeverity[severity]
}

// StandardLogger implements log.Logger.Log().
func (s *standardLogger) Log(entry logging.Entry) {
	if l, ok := s.bySeverity[entry.Severity]; ok {
		_ = l.Output(6, fmt.Sprintf("%v", entry.Payload))
		return
	}

	panic(fmt.Errorf("unknown log severity: %v", entry.Severity))
}

type flusher interface {
	Flush() error
}

type syncer interface {
	Sync() error
}

// StandardLogger implements log.Logger.Flush().
func (s *standardLogger) Flush() error {
	for _, logger := range s.bySeverity {
		w := logger.Writer()

		if f, ok := w.(flusher); ok {
			_ = f.Flush()
		}

		if f, ok := w.(syncer); ok {
			_ = f.Sync()
		}
	}

	return nil
}
