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
package list

import (
	"fmt"

	"github.com/jharrington22/aws-resource/cmd/list/all"
	"github.com/jharrington22/aws-resource/cmd/list/ec2"
	"github.com/jharrington22/aws-resource/cmd/list/elb"
	"github.com/jharrington22/aws-resource/cmd/list/elbv2"
	"github.com/jharrington22/aws-resource/cmd/list/images"
	"github.com/jharrington22/aws-resource/cmd/list/route53"
	"github.com/jharrington22/aws-resource/cmd/list/snapshots"
	"github.com/jharrington22/aws-resource/cmd/list/volumes"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS resources",
	Long: `List AWS resources
aws-resource list ec2
aws-resource list elb
aws-resource list elbv2
aws-resource list images
aws-resource list route53
aws-resource list snapshots
aws-resource list volumes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list called")
	},
}

func init() {

	ListCmd.AddCommand(all.Cmd)
	ListCmd.AddCommand(ec2.Cmd)
	ListCmd.AddCommand(elb.Cmd)
	ListCmd.AddCommand(elbv2.Cmd)
	ListCmd.AddCommand(images.Cmd)
	ListCmd.AddCommand(route53.Cmd)
	ListCmd.AddCommand(snapshots.Cmd)
	ListCmd.AddCommand(volumes.Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
