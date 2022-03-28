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
	"github.com/jharrington22/aws-resource/cmd/whoami"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

var (
	instanceNames bool
	profile       string
	roleArn       string
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
		Profile(profile).
		RoleArn(roleArn).
		Region(arguments.Region).
		Build()

	if err != nil {
		reporter.Errorf("Unable to build AWS client")
		return err
	}
	if profile != "" || roleArn != "" {
		err := whoami.WhoAmICmd.RunE(cmd, args)
		if err != nil {
			reporter.Errorf("Unable to verify AWS account %s", err)
		}
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
			Profile(profile).
			RoleArn(roleArn).
			Region(regionName).
			Build()

		if err != nil {
			reporter.Errorf("Unable to build AWS client in %s", regionName)
			return err
		}

		input := &ec2.DescribeInstancesInput{}

		result, err := awsClient.DescribeInstances(input)

		var instanceNamesList []string
		var runningInstanceList []*ec2.Instance
		for _, r := range result.Reservations {
			for _, i := range r.Instances {
				if *i.State.Name == "running" {
					runningInstanceList = append(runningInstanceList, i)
					if instanceNames {
						instanceNamesList = append(instanceNamesList, getInstanceName(i.Tags))
					}
				}
			}
		}
		if len(runningInstanceList) > 0 {
			instancesFound = true
			reporter.Infof("Found %d running instances in %s", len(runningInstanceList), regionName)
		}

		if len(instanceNamesList) != 0 {
			reporter.Infof("Instance names:\n%s", strings.Join(instanceNamesList, "\n"))
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

	Cmd.Flags().BoolVar(&instanceNames, "instance-names", false, "Print instance names")
	Cmd.Flags().StringVarP(&roleArn, "role-arn", "a", "", "AWS Role to assume")
	Cmd.Flags().StringVarP(&profile, "profile", "p", "", "AWS Profile to use")
}
