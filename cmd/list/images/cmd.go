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
package images

import (
	"os"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	imageId string
)

// Cmd represents the images command
var Cmd = &cobra.Command{
	Use:   "images",
	Short: "List AMIs",
	Long: `List AMIs for all or a specific region

aws-resource list images`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Listing images")

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

	err = listAllImages(reporter, logging, regions)
	if err != nil {
		reporter.Errorf("Unable to delete image: %s", err)
	}

	return
}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)
	Cmd.Flags().StringVarP(&imageId, "image-id", "i", "", "Delete specific image id")
}

func listAllImages(reporter *rprtr.Object, logging *logrus.Logger, regions *ec2.DescribeRegionsOutput) error {

	var allImages []*ec2.Image
	var snapshotBackedImages []*ec2.Image
	var allSnapshotBackedImages []*ec2.Image
	var allSnapshots []*string
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

		input := &ec2.DescribeImagesInput{
			Owners: []*string{&owner},
		}

		var images []*ec2.Image
		output, err := awsClient.DescribeImages(input)
		if err != nil {
			reporter.Errorf("Unable to describe images %s", err)
			return err
		}

		for _, image := range output.Images {
			for _, bdm := range image.BlockDeviceMappings {
				if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
					allSnapshots = append(allSnapshots, bdm.Ebs.SnapshotId)
					snapshotBackedImages = append(snapshotBackedImages, image)
					allSnapshotBackedImages = append(allSnapshotBackedImages, image)
				}
			}
			allImages = append(allImages, image)
			images = append(images, image)
		}

		if len(images) > 0 {
			reporter.Infof("Found %d images in %s", len(images), regionName)
			reporter.Infof("Found %d snapshot backed images in %s", len(images), regionName)
			reporter.Infof("Snapshots:")
			for _, s := range allSnapshots {
				reporter.Infof("%s", *s)
			}
		}

	}

	return nil
}
