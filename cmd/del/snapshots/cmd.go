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
package snapshots

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jharrington22/aws-resource/cmd/del/images"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

const (
	InvalidSnapshotInUse = "InvalidSnapshot.InUse"
)

var (
	dryRun             bool
	deleteBackingImage bool
	snapshotId         string
)

// Cmd represents the snapshots command
var Cmd = &cobra.Command{
	Use:   "snapshots",
	Short: "Delete EBS snapshots",
	Long: `Delete EBS snapshots for all or a specific region

aws-resource delete snapshots --region <region name>`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	reporter.Infof("Deleting ebs snapshots")

	if arguments.Region == "" {
		reporter.Errorf("Region required")
		return err
	}

	reporter.Infof("Region: %s", arguments.Region)

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

	owner := "self"

	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{&owner},
	}

	var snapshots []*ec2.Snapshot
	err = awsClient.DescribeSnapshotsPages(input, func(output *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		snapshots = append(snapshots, output.Snapshots...)
		if output.NextToken == nil {
			return false
		}
		return true
	})
	if err != nil {
		reporter.Errorf("Unable to describe snapshots %s", err)
		return err
	}

	if len(snapshots) == 0 {
		reporter.Infof("No snapshots found in %s", arguments.Region)
		return nil
	}

	if !dryRun {
		reporter.Warnf("Dry run %t will delete resources", dryRun)
	}

	for _, snapshot := range snapshots {
		input := &ec2.DeleteSnapshotInput{
			DryRun:     &dryRun,
			SnapshotId: snapshot.SnapshotId,
		}

		fmt.Println("here")

		output, err := awsClient.DeleteSnapshot(input)
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println("in error")
			switch aerr.Code() {
			case InvalidSnapshotInUse:
				reporter.Infof("Snapshot in use with backing AMI use --delete-backing-image to force delete: %s", aerr.Message())
				amiId, err := parseInUseSnapshotImageIdErr(aerr.Message())
				if err != nil {
					reporter.Errorf("Parsing of snapshot (%s) in use error: %v", *snapshot.SnapshotId, err)
					continue
				}
				reporter.Infof("AMI: %v", amiId)
				if deleteBackingImage {
					flags := images.Cmd.Flags()
					reporter.Infof("Setting flags image id: %s dry-run: %t", amiId, dryRun)
					err = flags.Set("image-id", amiId)
					if err != nil {
						reporter.Errorf("Unable to set image flag: %s", err)
					}
					err = flags.Set("dry-run", strconv.FormatBool(dryRun))
					if err != nil {
						reporter.Errorf("Unable to set dry run flag: %s", err)
					}
					err = images.Cmd.RunE(cmd, args)
					if err != nil {
						reporter.Errorf("Unable to list EC2 instances: %s", err)
						return err
					}
					_, err = awsClient.DeleteSnapshot(input)
					if err != nil {
						return reporter.Errorf("Unable to delete snapshot: %s", err)
					}
				}
			default:
				fmt.Printf("error %s", err)

			}
		} else {
			reporter.Errorf("Delete snapshot failed: %s", err)
		}

		fmt.Println("Finished")

		if output != nil {
			reporter.Infof("Snapshot %s deleted", *snapshot.SnapshotId)
		}
	}

	return

}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)

	Cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate if delete would be successful")
	Cmd.Flags().BoolVar(&deleteBackingImage, "delete-backing-image", false, "Delete snapshots backing AMI")
}

func parseInUseSnapshotImageIdErr(errorMsg string) (string, error) {
	pattern := regexp.MustCompile(`.*(ami-[0-9a-z]{17}).*`)
	matches := pattern.FindStringSubmatch(errorMsg)
	if len(matches) != 2 {
		return "", fmt.Errorf("Unexpected number of in use Image ID's for snapshot: %v", matches)
	}
	return matches[1], nil
}
