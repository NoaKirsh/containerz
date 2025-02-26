// Copyright 2025 Google LLC
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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/prototext"
)

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List plugins and optionally list a specific plugin by instance name",
	RunE: func(command *cobra.Command, args []string) error {
		resp, err := containerzClient.ListPlugin(command.Context(), instance)
		if err != nil {
			return err
		}

		for _, res := range resp {
			fmt.Println(prototext.Format(res))
		}

		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
}
