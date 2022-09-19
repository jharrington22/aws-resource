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
package elbv2

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "elbv2",
	Short: "List ELBv2 instances",
	Long: `List ELBv2 instances for all or a specific region

aws-resource list elbv2`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Listing elbv2 instances")

	awsClient, err := aws.NewClient().
		Logger(logging).
		Profile(arguments.Profile).
		RoleArn(arguments.RoleArn).
		Region(arguments.Region).
		Build()

	if err != nil {
		return reporter.Errorf("Unable to build AWS client")
	}

	regions, err := awsClient.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return reporter.Errorf("Failed to describe regions")
	}

	var runningLoadBalancersV2DescriptionsList []*elbv2.LoadBalancer
	for _, region := range regions.Regions {

		regionName := *region.RegionName

		awsClient, err := aws.NewClient().
			Logger(logging).
			Profile(arguments.Profile).
			RoleArn(arguments.RoleArn).
			Region(regionName).
			Build()

		if err != nil {
			return reporter.Errorf("Unable to build AWS client in %s", regionName)
		}

		input := &elbv2.DescribeLoadBalancersInput{}

		result, err := awsClient.DescribeV2LoadBalancers(input)
		if err != nil {
			return reporter.Errorf("Unable to describe load balancers v2: %s", err)
		}

		var loadBalancersV2 []*elbv2.LoadBalancer
		for _, loadBalancerV2 := range result.LoadBalancers {
			loadBalancersV2 = append(loadBalancersV2, loadBalancerV2)
			runningLoadBalancersV2DescriptionsList = append(runningLoadBalancersV2DescriptionsList, loadBalancerV2)
		}

		if len(loadBalancersV2) > 0 {
			reporter.Infof("Found %d running v2 load balancers in %s", len(loadBalancersV2), regionName)
		}
	}
	if len(runningLoadBalancersV2DescriptionsList) == 0 {
		reporter.Infof("No running v2 load balancers found")
	}
	return
}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)
}
