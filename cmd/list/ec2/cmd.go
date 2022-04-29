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
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

var (
	imageId       bool
	instanceNames bool
	instanceType  bool
	launchTime    bool
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "List EC2 instances",
	Long: `List EC2 instances for all or a specific region

aws-resource list ec2`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {
	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Listing ec2 instances")

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
		reporter.Errorf("Failed to describe regions; %s", err)
		return err
	}

	var instancesFound bool

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
			return err
		}

		input := &ec2.DescribeInstancesInput{}

		result, err := awsClient.DescribeInstances(input)

		var runningInstanceList []*ec2.Instance
		var detail []string
		for _, r := range result.Reservations {
			for _, i := range r.Instances {
				if *i.State.Name == "running" {
					runningInstanceList = append(runningInstanceList, i)
					if instanceNames {
						detail = append(detail, getInstanceName(i.Tags))
					}
					if launchTime {
						detail = append(detail, i.LaunchTime.String())
					}
					if instanceType {
						detail = append(detail, *i.InstanceType)
					}
					if imageId {
						detail = append(detail, *i.ImageId)
					}
				}
			}
		}
		if len(runningInstanceList) > 0 {
			instancesFound = true
			reporter.Infof("Found %d running instances in %s", len(runningInstanceList), regionName)
		}
		if (instanceNames || launchTime || instanceType) && len(detail) > 0 {
			reporter.Infof("%s", strings.Join(detail, ", "))
		}
	}
	if !instancesFound {
		reporter.Infof("No instances found in account")
	}

	return
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
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)

	Cmd.Flags().BoolVar(&imageId, "image-id", false, "Print image id")
	Cmd.Flags().BoolVar(&instanceNames, "instance-names", false, "Print instance names")
	Cmd.Flags().BoolVar(&instanceType, "instance-type", false, "Print instance type")
	Cmd.Flags().BoolVar(&launchTime, "launch-time", false, "Print instance launch time")
}
