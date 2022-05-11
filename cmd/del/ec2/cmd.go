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
	dryRun bool
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "Delete EC2 instances",
	Long: `Delete EC2 instances for all or a specific region

aws-resource list ec2`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {
	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Deleting ec2 instances")

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

		instances := []*ec2.Instance{}
		err = awsClient.DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, r := range page.Reservations {
				instances = append(instances, r.Instances...)
			}
			return !lastPage
		})
		if err != nil {
			reporter.Errorf("Unable to describe instance pages %s", err)
		}

		var runningInstanceList []*ec2.Instance
		for _, i := range instances {
			if *i.State.Name == "running" {
				runningInstanceList = append(runningInstanceList, i)
			}
		}
		if len(runningInstanceList) > 0 {
			instancesFound = true
			reporter.Infof("Terminating %d running instances in %s", len(runningInstanceList), regionName)
			instanceIds := []*string{}
			for _, i := range runningInstanceList {
				instanceIds = append(instanceIds, i.InstanceId)
			}
			output, err := terminateInstances(awsClient, instanceIds, dryRun)
			if err != nil {
				reporter.Errorf("Could not terminate instances: %s", err)
			}
			deletedInstancesList := []string{}
			// The TerminatingInstances slice will be nil if dryRun is set
			if !dryRun {
				for _, di := range output.TerminatingInstances {
					deletedInstancesList = append(deletedInstancesList, *di.InstanceId)

				}
				reporter.Infof("Instance ID deleted %s", strings.Join(deletedInstancesList, ","))
			}
			if len(runningInstanceList) != len(deletedInstancesList) {
				reporter.Errorf("Number of deleted instance IDs %d does not match number of running instances %d in %s",
					len(deletedInstancesList), len(runningInstanceList), regionName)
			}
		}
	}
	if !instancesFound {
		reporter.Infof("No running instances found in account")
	}

	return
}

func terminateInstances(awsClient aws.Client, instanceIds []*string, dryRun bool) (*ec2.TerminateInstancesOutput, error) {
	input := ec2.TerminateInstancesInput{
		DryRun:      &dryRun,
		InstanceIds: instanceIds,
	}

	output, err := awsClient.TerminateInstances(&input)
	if err != nil {
		return nil, err
	}

	return output, err
}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)

	Cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate if delete would be successful")
}
