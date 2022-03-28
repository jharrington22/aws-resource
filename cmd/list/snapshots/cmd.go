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
	"os"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// Cmd represents the snapshots command
var Cmd = &cobra.Command{
	Use:   "snapshots",
	Short: "List EBS snapshots",
	Long: `List EBS snapshots for all or a specific region

aws-resource list snapshots`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Listing ebs snapshots")

	awsClient, err := aws.NewClient().
		Logger(logging).
		Profile(arguments.Profile).
		RoleArn(arguments.RoleArn).
		Region(arguments.Region).
		Build()

	if err != nil {
		reporter.Errorf("Unable to build AWS client")
		return err
	}

	regions, err := awsClient.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		reporter.Errorf("Failed to describe regions")
		return err
	}

	var availableSnapshots []*ec2.Snapshot
	for _, region := range regions.Regions {

		regionName := *region.RegionName

		awsClient, err := aws.NewClient().
			Logger(logging).
			Profile(arguments.Profile).
			RoleArn(arguments.RoleArn).
			Region(regionName).
			Build()

		if err != nil {
			reporter.Errorf("Unable to build AWS client in %s", regionName)
			os.Exit(1)
		}

		owner := "self"

		input := &ec2.DescribeSnapshotsInput{
			OwnerIds: []*string{&owner},
		}

		result, err := awsClient.DescribeSnapshots(input)
		if err != nil {
			reporter.Errorf("Unable to describe snapshots %s", err)
			return err
		}

		var snapshots []*ec2.Snapshot
		for _, snapshot := range result.Snapshots {
			snapshots = append(snapshots, snapshot)
			availableSnapshots = append(availableSnapshots, snapshot)
		}

		if len(snapshots) > 0 {
			reporter.Infof("Found %d snapshots in %s", len(snapshots), regionName)
		}
	}

	if len(availableSnapshots) == 0 {
		reporter.Infof("No snapshots found")
	}

	return

}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)
}
