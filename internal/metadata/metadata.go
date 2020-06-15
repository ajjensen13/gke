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
package metadata

import (
	"cloud.google.com/go/compute/metadata"
	"sync"
)

type MetadataType struct {
	OnGCE              bool
	ProjectID          string
	InstanceID         string
	InstanceName       string
	ExternalIP         string
	InternalIP         string
	Hostname           string
	Zone               string
	InstanceTags       []string
	InstanceAttributes map[string]string
	ProjectAttributes  map[string]string
}

var (
	pkgMetadata     MetadataType
	pkgMetadataOnce sync.Once
)

func initMetadata() {
	pkgMetadata.OnGCE = metadata.OnGCE()
	if !pkgMetadata.OnGCE {
		return
	}

	pid, err := metadata.ProjectID()
	if err != nil {
		panic(err)
	}
	pkgMetadata.ProjectID = pid

	iid, err := metadata.InstanceID()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceID = iid

	instName, err := metadata.InstanceName()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceName = instName

	exIP, err := metadata.ExternalIP()
	if err != nil {
		panic(err)
	}
	pkgMetadata.ExternalIP = exIP

	inIP, err := metadata.InternalIP()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InternalIP = inIP

	host, err := metadata.Hostname()
	if err != nil {
		panic(err)
	}
	pkgMetadata.Hostname = host

	vals, err := metadata.InstanceAttributes()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceAttributes = make(map[string]string, len(vals))

	for _, val := range vals {
		v, err := metadata.InstanceAttributeValue(val)
		if err != nil {
			panic(err)
		}
		pkgMetadata.InstanceAttributes[val] = v
	}

	instTags, err := metadata.InstanceTags()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceTags = instTags

	vals, err = metadata.ProjectAttributes()
	if err != nil {
		panic(err)
	}
	pkgMetadata.ProjectAttributes = make(map[string]string, len(vals))
	for _, val := range vals {
		v, err := metadata.ProjectAttributeValue(val)
		if err != nil {
			panic(err)
		}
		pkgMetadata.ProjectAttributes[val] = v
	}

	zone, err := metadata.Zone()
	if err != nil {
		panic(err)
	}
	pkgMetadata.Zone = zone
}

func Metadata() *MetadataType {
	pkgMetadataOnce.Do(initMetadata)
	return &pkgMetadata
}
