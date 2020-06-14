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
	"context"
)

func NewGkeClient(ctx context.Context, parent string) (Client, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return nil, err
	}
	return GkeClient{client}, nil
}

type GkeClient struct {
	client *logging.Client
}

func (g GkeClient) Logger(logID string, opts ...logging.LoggerOption) Logger {
	return g.client.Logger(logID, opts...)
}

func (g GkeClient) Close() error {
	panic("implement me")
}

type GkeLogger struct {
	*logging.Logger
}
