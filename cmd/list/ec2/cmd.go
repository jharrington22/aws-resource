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
package ec2

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

var instanceNames bool

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "List EC2 instances",
	Long: `List EC2 instances for all or a specific region

aws-resource list ec2`,
	Run: func(cmd *cobra.Command, args []string) {

		reporter := rprtr.CreateReporterOrExit()
		logging := logging.CreateLoggerOrExit(reporter)

		awsClient, err := aws.NewClient().
			Logger(logging).
			Region(arguments.Region).
			Build()

		if err != nil {
			reporter.Errorf("Unable to build AWS client")
			os.Exit(1)
		}

		regions, err := awsClient.DescribeRegions(&ec2.DescribeRegionsInput{})
		if err != nil {
			reporter.Errorf("Failed to describe regions")
			os.Exit(1)
		}

		for _, region := range regions.Regions {

			regionName := *region.RegionName

			awsClient, err := aws.NewClient().
				Logger(logging).
				Region(regionName).
				Build()

			if err != nil {
				reporter.Errorf("Unable to build AWS client in %s", regionName)
				os.Exit(1)
			}

			input := &ec2.DescribeInstancesInput{}

			result, err := awsClient.DescribeInstances(input)

			instanceCount := 0
			var instanceNamesList []string
			var runningInstanceList []*ec2.Instance
			for _, r := range result.Reservations {
				for _, i := range r.Instances {
					if *i.State.Name == "running" {
						runningInstanceList = append(runningInstanceList, i)
						instanceCount++
						if instanceNames {
							instanceNamesList = append(instanceNamesList, getInstanceName(i.Tags))
						}
					}
				}
			}
			reporter.Infof("Found %d running instances in %s", instanceCount, regionName)

			if len(instanceNamesList) != 0 {
				reporter.Infof("%s", strings.Join(instanceNamesList, "\n"))
			}
		}
	},
}

func getInstanceName(tags []*ec2.Tag) string {
	if len(tags) == 0 {
		return "Instance has no tags"
	}
	for _, t := range tags {
		if *t.Key == "Name" {
			return *t.Value
		}
	}
	return "Instance has no tag \"Name\""
}

func init() {

	Cmd.Flags().BoolVar(&instanceNames, "instance-names", false, "Print instance names")
}
