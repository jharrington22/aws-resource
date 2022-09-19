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
	"fmt"
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
	dryRun  bool
)

// Cmd represents the images command
var Cmd = &cobra.Command{
	Use:   "images",
	Short: "Delete Images",
	Long: `Delete images for all or a specific region

aws-resource delete images`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

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

	if imageId != "" {
		_, err := deleteImageId(awsClient, imageId, dryRun)
		if err != nil {
			_ = reporter.Errorf("Unable to delete image: %s", err)
		}
		reporter.Infof("Image %s deregistered", imageId)
	}

	if imageId == "" {
		reporter.Infof("No image id specified")
		err := deleteAllImages(reporter, logging, regions, dryRun)
		if err != nil {
			_ = reporter.Errorf("Unable to delete image: %s", err)
		}
	}

	return
}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)
	Cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate if delete would be successful")
	Cmd.Flags().StringVarP(&imageId, "image-id", "i", "", "Delete specific image id")
}

func deleteImageId(client aws.Client, imageId string, dryRun bool) (*ec2.DeregisterImageOutput, error) {

	input := &ec2.DeregisterImageInput{
		DryRun:  &dryRun,
		ImageId: &imageId,
	}

	output, err := client.DeregisterImage(input)
	if err != nil {
		return nil, fmt.Errorf("Unable to deregister image: %s", err)
	}

	return output, err
}

func deleteAllImages(reporter *rprtr.Object, logging *logrus.Logger, regions *ec2.DescribeRegionsOutput, dryRun bool) error {

	for _, region := range regions.Regions {

		regionName := *region.RegionName

		awsClient, err := aws.NewClient().
			Logger(logging).
			Profile(arguments.Profile).
			RoleArn(arguments.RoleArn).
			Region(regionName).
			Build()

		if err != nil {
			_ = reporter.Errorf("Unable to build AWS client in %s", regionName)
			os.Exit(1)
		}

		owner := "self"

		input := &ec2.DescribeImagesInput{
			Owners: []*string{&owner},
		}

		var images []*ec2.Image
		output, err := awsClient.DescribeImages(input)
		if err != nil {
			return reporter.Errorf("Unable to describe images %s", err)
		}

		for _, image := range output.Images {
			images = append(images, image)
			input := &ec2.DeregisterImageInput{
				DryRun:  &dryRun,
				ImageId: image.ImageId,
			}

			_, err := awsClient.DeregisterImage(input)
			if err != nil {
				return fmt.Errorf("Unable to deregister image: %s", err)
			}
			// Here we now can delete the backing snapshot by calling delete
			// snapshot again

			reporter.Infof("Image %s deregistered", imageId)
		}

		if len(images) > 0 {
			reporter.Infof("Found %d images in %s", len(images), regionName)
		}

	}

	return nil
}
