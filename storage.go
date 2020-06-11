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
	"cloud.google.com/go/storage"
	"context"
)

func provideStorageClient(ctx context.Context) (StorageClient, func(), error) {
	result, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return result, func() { _ = result.Close() }, nil
}

type StorageClient interface {
	HMACKeyHandle(projectID, accessID string) *storage.HMACKeyHandle
	CreateHMACKey(ctx context.Context, projectID, serviceAccountEmail string, opts ...storage.HMACKeyOption) (*storage.HMACKey, error)
	ListHMACKeys(ctx context.Context, projectID string, opts ...storage.HMACKeyOption) *storage.HMACKeysIterator
	ServiceAccount(ctx context.Context, projectID string) (string, error)
	Bucket(name string) *storage.BucketHandle
	Buckets(ctx context.Context, projectID string) *storage.BucketIterator
}
