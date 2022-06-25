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
package elb

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "elb",
	Short: "Delete ELB instances",
	Long: `Delete ELB instances for all or a specific region

aws-resource delete elb`,
	RunE: run,
}

var (
	allRegions bool
	dryRun     bool
)

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Deleting elb instances")

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

	var runningLoadBalancerDescriptionsList []*elb.LoadBalancerDescription
	var deletedLoadBalancerDescriptionsList []*elb.LoadBalancerDescription
	var loadBalancerDescriptions []*elb.LoadBalancerDescription
	if allRegions {
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

			input := &elb.DescribeLoadBalancersInput{}

			err = awsClient.DescribeLoadBalancersPages(input, func(page *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
				loadBalancerDescriptions = append(loadBalancerDescriptions, page.LoadBalancerDescriptions...)
				return !lastPage
			})
		}

		if arguments.Region != "" && !allRegions {

			awsClient, err := aws.NewClient().
				Logger(logging).
				Profile(arguments.Profile).
				RoleArn(arguments.RoleArn).
				Region(arguments.Region).
				Build()

			if err != nil {
				reporter.Errorf("Unable to build AWS client in %s", arguments.Region)
				return err
			}

			input := &elb.DescribeLoadBalancersInput{}

			var loadBalancerDescriptions []*elb.LoadBalancerDescription
			err = awsClient.DescribeLoadBalancersPages(input, func(page *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
				loadBalancerDescriptions = append(loadBalancerDescriptions, page.LoadBalancerDescriptions...)
				return !lastPage
			})

		}

		if len(loadBalancerDescriptions) > 0 {
			reporter.Infof("Found %d running load balancers in %s", len(loadBalancerDescriptions), regionName)
			awsClient.DeleteLoadBalancer()
		}
	}

	if len(runningLoadBalancerDescriptionsList) == 0 {
		reporter.Infof("No running load balancers found")
	}
	return

}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)

	Cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate if delete would be successful")
	Cmd.Flags().BoolVar(&allRegions, "all-regions", false, "Delete snapshots in all regions")
}

func deleteLoadBalancer(awsClient aws.Client, reporter *rptr.Object, loadBalancerDescriptionList []*elb.LoadBalancerDescription, dryRun bool) {

}
