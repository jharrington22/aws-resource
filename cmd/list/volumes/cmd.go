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
package volumes

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// Cmd represents the volumes command
var Cmd = &cobra.Command{
	Use:   "volumes",
	Short: "List EBS volumes",
	Long: `List EBS volumes for all or a specific region

aws-resource list volumes`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {
	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Listing ebs volumes")

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

	var availableVolumes []*ec2.Volume
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

		input := &ec2.DescribeVolumesInput{}

		result, err := awsClient.DescribeVolumes(input)
		if err != nil {
			return reporter.Errorf("Unable to describe volumes %s", err)
		}

		var volumes []*ec2.Volume
		var attachedVolumes []*ec2.Volume
		var unattachedVolumes []*ec2.Volume
		for _, volume := range result.Volumes {
			volumes = append(volumes, volume)
			availableVolumes = append(availableVolumes, volume)
			if len(volume.Attachments) > 0 {
				attachedVolumes = append(attachedVolumes, volume)
			} else {
				unattachedVolumes = append(unattachedVolumes, volume)
			}
		}
		if len(volumes) > 0 {
			reporter.Infof("Found %d volumes in %s", len(volumes), regionName)
		}
		if len(volumes) > 0 {
			reporter.Infof("%d attached volumes in %s", len(attachedVolumes), regionName)
		}
		if len(volumes) > 0 {
			reporter.Infof("%d unattached volumes in %s", len(unattachedVolumes), regionName)
		}
	}
	if len(availableVolumes) == 0 {
		reporter.Infof("No volumes found")
	}
	return

}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)
}
