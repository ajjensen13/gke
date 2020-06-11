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
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"github.com/google/wire"
	"log"
	"net"
	"net/http"
	"time"
)

func NewServer(lg *logging.Logger, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Second * 60,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		ErrorLog:          lg.StandardLogger(logging.Error),
		BaseContext: func(_ net.Listener) (ctx context.Context) {
			ctx, _ = Alive()
			return
		},
	}
}

// NewLogClient returns a GCP logging client
func NewLogClient(ctx context.Context, opts ...logging.LoggerOption) (*LogClient, error) {
	panic(wire.Build(provideLoggingClient, provideLogClient, wire.Value(&pkgConfig)))
}

func provideLoggingClient(ctx context.Context, config *Config) (*logging.Client, error) {
	result, err := logging.NewClient(ctx, config.ProjectId)
	if err != nil {
		return nil, err
	}

	result.OnError = func(err error) {
		log.Printf("%v", err)
	}

	err = result.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func provideLogClient(client *logging.Client, config *Config, opts []logging.LoggerOption) (*LogClient, error) {
	var defaultOpts []logging.LoggerOption
	if config.CommonLogLabels != nil && len(config.CommonLogLabels) > 0 {
		defaultOpts = append(defaultOpts, logging.CommonLabels(config.CommonLogLabels))
	}
	if len(opts) > 0 {
		defaultOpts = append(defaultOpts, opts...)
	}

	return &LogClient{
		Client: client,
		opts:   defaultOpts,
	}, nil
}

type LogClient struct {
	opts []logging.LoggerOption
	*logging.Client
}

func (l *LogClient) Logger(logID string, opts ...logging.LoggerOption) *Logger {
	os := append(l.opts, opts...)
	return &Logger{
		logId:  logID,
		client: l,
		opts:   os,
		Logger: l.Client.Logger(
			logID,
			os...,
		),
	}
}

type Logger struct {
	logId  string
	client *LogClient
	opts   []logging.LoggerOption
	*logging.Logger
}

type fmtPayload struct {
	Message string
	Args    []interface{}
}

type errPayload struct {
	Message string
	Err     error
}

func (l *Logger) Logf(severity logging.Severity, format string, args ...interface{}) string {
	message := fmt.Sprintf(format, args...)
	l.Log(logging.Entry{Severity: severity, Payload: fmtPayload{
		Message: message,
		Args:    args,
	}})
	return message
}

func (l *Logger) Infof(format string, args ...interface{}) string {
	return l.Logf(logging.Info, format, args...)
}

func (l *Logger) Noticef(format string, args ...interface{}) string {
	return l.Logf(logging.Notice, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) string {
	return l.Logf(logging.Warning, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) string {
	return l.Logf(logging.Error, format, args...)
}

func (l *Logger) InfoErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Info, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *Logger) NoticeErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Notice, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *Logger) WarnErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Warning, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *Logger) ErrorErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Error, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *Logger) Child(suffix string, opts ...logging.LoggerOption) *Logger {
	return l.client.Logger(
		l.logId+"-"+suffix,
		append(l.opts, opts...)...,
	)
}
