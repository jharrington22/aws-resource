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
	"os"

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

			input := &elbv2.DescribeLoadBalancersInput{}

			result, err := awsClient.DescribeV2LoadBalancers(input)

			var loadBalancersV2 []*elbv2.LoadBalancer
			for _, loadBalancerV2 := range result.LoadBalancers {
				loadBalancersV2 = append(loadBalancersV2, loadBalancerV2)
			}
			reporter.Infof("Found %d running v2 load balancers in %s", len(loadBalancersV2), regionName)
		}

	},
}

func init() {
}
