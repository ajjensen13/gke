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
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	"os"

	"github.com/ajjensen13/gke/internal/metadata"
)

func NewGkeClient(ctx context.Context, parent string) (GkeClient, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return GkeClient{}, err
	}
	return GkeClient{client}, nil
}

type GkeClient struct {
	client *logging.Client
}

func (g GkeClient) Logger(logID string) Logger {
	md := metadata.Metadata()
	cn, _ := os.LookupEnv("CLUSTER_NAME")
	ns, ok := os.LookupEnv("NAMESPACE_NAME")
	if !ok {
		ns = "default"
	}
	return g.client.Logger(logID, logging.CommonResource(&mrpb.MonitoredResource{
		Type: "k8s_container",
		Labels: map[string]string{
			"project_id":     md.ProjectID,
			"location":       md.Zone,
			"cluster_name":   cn,
			"namespace_name": ns,
		},
	}))
}

func (g GkeClient) Close() error {
	return g.client.Close()
}

func (g GkeClient) Ping(ctx context.Context) error {
	return g.client.Ping(ctx)
}

type GkeLogger struct {
	*logging.Logger
}
