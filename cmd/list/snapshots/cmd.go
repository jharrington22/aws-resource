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
package snapshots

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var Cmd = &cobra.Command{
	Use:   "snapshots",
	Short: "List EBS snapshots",
	Long: `List EBS snapshots for all or a specific region

aws-resource list snapshots.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("snapshot list called")

		reporter := rprtr.CreateReporterOrExit()
		logging := logging.CreateLoggerOrExit(reporter)

		awsClient, err := aws.NewClient().
			Logger(logging).
			Region(arguments.Region).
			Build()

		if err != nil {
			fmt.Errorf("Unable to build AWS client")
			os.Exit(1)
		}

		regions, err := awsClient.DescribeRegions(&ec2.DescribeRegionsInput{})
		if err != nil {
			fmt.Errorf("Failed to describe regions")
			os.Exit(1)
		}

		for _, region := range regions.Regions {

			regionName := *region.RegionName

			awsClient, err := aws.NewClient().
				Logger(logging).
				Region(regionName).
				Build()

			if err != nil {
				fmt.Errorf("Unable to build AWS client in %s", regionName)
				os.Exit(1)
			}

			owner := "self"

			input := &ec2.DescribeSnapshotsInput{
				OwnerIds: []*string{&owner},
			}

			result, err := awsClient.DescribeSnapshots(input)

			snapshotCount := 0
			for _, _ = range result.Snapshots {
				// fmt.Println(fmt.Sprintf("id: %s", *i.InstanceId))
				// fmt.Println(fmt.Sprintf("id: %v", i.Tags))
				// fmt.Println(fmt.Sprintf("state: %v", *i.State.Name))
				// if *i.ImageId != ami["initializationAMI"] {
				// 	fmt.Println("Instance doesn't have matching AMI ID, not terminating")
				// 	continue
				// }

				snapshotCount++

				//err := TerminateEC3Instance(svc, *i.InstanceId)
				//if err != nil {
				//	fmt.Println(err)
				//	continue
				//}
			}
			fmt.Printf("Found %d snapshots in %s\n", snapshotCount, regionName)
		}

	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
