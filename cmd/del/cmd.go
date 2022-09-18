/*
Copyright © 2022 James Harrington <james@harrington.net.au>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package del

import (
	"fmt"

	"github.com/jharrington22/aws-resource/cmd/del/ec2"
	"github.com/jharrington22/aws-resource/cmd/del/images"
	"github.com/jharrington22/aws-resource/cmd/del/snapshots"
	"github.com/jharrington22/aws-resource/cmd/del/subnets"
	"github.com/spf13/cobra"
)

// DeleteCmd represents the list command
var DelCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete AWS resources",
	Long: `Delete AWS resources
aws-resource delete snapshots`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")
	},
}

func init() {

	DelCmd.AddCommand(ec2.Cmd)
	DelCmd.AddCommand(images.Cmd)
	DelCmd.AddCommand(snapshots.Cmd)
	DelCmd.AddCommand(subnets.Cmd)

}
