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
	allRegions         bool
	dryRun             bool
	deleteBackingImage bool
	snapshotId         string
)

// Cmd represents the snapshots command
var Cmd = &cobra.Command{
	Use:     "snapshots",
	Aliases: []string{"snapshot"},
	Short:   "Delete EBS snapshots",
	Long: `Delete EBS snapshots for all or a specific region

aws-resource delete snapshots --region <region name>`,
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
		reporter.Errorf("Unable to build AWS client")
		return err
	}

	if !dryRun {
		reporter.Warnf("Dry run %t will delete resources", dryRun)
	}

	var snapshots []*ec2.Snapshot
	if allRegions {
		reporter.Infof("Deleting ebs snapshots in all regions")
		regions, err := awsClient.DescribeRegions(&ec2.DescribeRegionsInput{})
		if err != nil {
			reporter.Errorf("Failed to describe regions")
			return err
		}

		for _, region := range regions.Regions {
			awsClient, err := aws.NewClient().
				Logger(logging).
				Profile(arguments.Profile).
				RoleArn(arguments.RoleArn).
				Region(*region.RegionName).
				Build()

			if err != nil {
				reporter.Errorf("Unable to build AWS client for region: %s: %s", *region.RegionName, err)
			}

			ss, err := describeSnapshots(awsClient, reporter)
			if err != nil {
				reporter.Errorf("Unable to describe snapshots for region %s: err", *region.RegionName)
				return err
			}
			snapshots = append(snapshots, ss...)
			for _, s := range ss {
				deleteSnapshot(awsClient, cmd, reporter, s, dryRun)
				if err != nil {
					fmt.Errorf("Unable to delete shapshot: %s", err)
					return err
				}
				reporter.Infof("Deleted snapshot %s", *s.SnapshotId)

			}
		}

	}

	if snapshotId != "" {
		snapshot, err := awsClient.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{&snapshotId},
		})

		if err != nil {
			reporter.Errorf("Unable to describe snapshot id: %s", snapshotId)
			return err
		}
		snapshots = append(snapshots, snapshot.Snapshots...)
		for _, s := range snapshot.Snapshots {
			deleteSnapshot(awsClient, cmd, reporter, s, dryRun)
			if err != nil {
				fmt.Errorf("Unable to delete shapshot: %s", err)
				return err
			}
			reporter.Infof("Deleted snapshot %s", *s.SnapshotId)
		}

	}

	// FIXME: Check if the default has been specified or remove the
	// default of us-east-1 else this will always match
	if (snapshotId == "" && arguments.Region != "") && !allRegions {
		reporter.Infof("Deleting ebs snapshots in %s", arguments.Region)
		snapshots, err = describeSnapshots(awsClient, reporter)
		if err != nil {
			reporter.Errorf("Unable to describe snapshots for region %s", arguments.Region)
			return err
		}
	}

	if len(snapshots) == 0 {
		msg := arguments.Region
		if allRegions {
			msg = "all regions"
		}
		reporter.Infof("No snapshots found in %s", msg)
		return nil
	}

	return nil

}

func init() {
	// Add global flags
	flags := Cmd.Flags()
	arguments.AddFlags(flags)

	Cmd.Flags().BoolVar(&allRegions, "all-regions", false, "Delete snapshots in all regions")
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

func deleteSnapshot(awsClient aws.Client, cmd *cobra.Command, reporter *rprtr.Object, snapshot *ec2.Snapshot, dryRun bool) error {
	input := &ec2.DeleteSnapshotInput{
		DryRun:     &dryRun,
		SnapshotId: snapshot.SnapshotId,
	}

	// Output from delete snapshot doesn't contain information we need
	_, err := awsClient.DeleteSnapshot(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case InvalidSnapshotInUse:
			reporter.Infof("Snapshot in use with backing AMI use --delete-backing-image to force delete: %s", aerr.Message())
			amiId, err := parseInUseSnapshotImageIdErr(aerr.Message())
			if err != nil {
				reporter.Errorf("Parsing of snapshot (%s) in use error: %v", *snapshot.SnapshotId, err)
				return err
			}
			if deleteBackingImage {
				flags := images.Cmd.Flags()
				// Setting image ID for deletion
				err = flags.Set("image-id", amiId)
				if err != nil {
					reporter.Errorf("Unable to set image flag: %s", err)
				}
				// Setting dry-run flag if specified
				err = flags.Set("dry-run", strconv.FormatBool(dryRun))
				if err != nil {
					reporter.Errorf("Unable to set dry run flag: %s", err)
				}
				err = images.Cmd.RunE(cmd, []string{})
				if err != nil {
					reporter.Errorf("Unable to list EC2 instances: %s", err)
					return err
				}
				_, err = awsClient.DeleteSnapshot(input)
				if err != nil {
					return reporter.Errorf("Unable to delete snapshot: %s", err)
				}
				reporter.Infof("Snapshot %s deleted", *snapshot.SnapshotId)
			}
		// Don't return the error here just report that the deletion would have been
		// successful without the dryRun flag set
		case "DryRunOperation":
			reporter.Infof("deletion of %s in %s", *snapshot.SnapshotId, aerr.Message())
			return nil
		default:
			reporter.Errorf("error %s", err)
			return err

		}
	} else {
		reporter.Errorf("Delete snapshot failed: %s", err)
		return err

	}
	return nil
}

func describeSnapshots(awsClient aws.Client, reporter *rprtr.Object) ([]*ec2.Snapshot, error) {
	owner := "self"

	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{&owner},
	}

	var snapshots []*ec2.Snapshot
	err := awsClient.DescribeSnapshotsPages(input, func(output *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		snapshots = append(snapshots, output.Snapshots...)
		return !lastPage
	})
	if err != nil {
		reporter.Errorf("Unable to describe snapshots %s", err)
		return nil, err
	}
	return snapshots, err
}
