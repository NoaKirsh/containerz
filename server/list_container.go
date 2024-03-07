// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"github.com/openconfig/containerz/containers"

	cpb "github.com/openconfig/gnoi/containerz"
)

// ListContainer returns all containers that match the spec defined in the request.
func (s *Server) ListContainer(request *cpb.ListContainerRequest, srv cpb.Containerz_ListContainerServer) error {
	filters := make(map[options.FilterKey][]string, len(request.GetFilter()))
	for _, f := range request.GetFilter() {
		vals, ok := filters[options.FilterKey(f.GetKey())]
		if !ok {
			vals = []string{}
		}
		filters[options.FilterKey(f.GetKey())] = append(vals, f.GetValue()...)
	}

	return s.mgr.ContainerList(srv.Context(), request.GetAll(), request.GetLimit(), srv, options.WithFilter(filters))
}
