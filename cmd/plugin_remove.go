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
)

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a plugin by instance name",
	RunE: func(command *cobra.Command, args []string) error {
		if instance == "" {
			fmt.Println("--instance must be provided")
		}

		if err := containerzClient.RemovePlugin(command.Context(), instance); err != nil {
			return err
		}

		fmt.Printf("Successfully removed %s\n", instance)
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginRemoveCmd)
}
