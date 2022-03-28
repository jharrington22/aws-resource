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
package all

import (
	ec2cmd "github.com/jharrington22/aws-resource/cmd/list/ec2"
	elbcmd "github.com/jharrington22/aws-resource/cmd/list/elb"
	elbv2cmd "github.com/jharrington22/aws-resource/cmd/list/elbv2"
	route53cmd "github.com/jharrington22/aws-resource/cmd/list/route53"
	snapshotscmd "github.com/jharrington22/aws-resource/cmd/list/snapshots"
	volumescmd "github.com/jharrington22/aws-resource/cmd/list/volumes"
	"github.com/jharrington22/aws-resource/cmd/whoami"
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
	Use:   "all",
	Short: "List all AWS resources",
	Long: `List all AWS resources supported by aws-resource

aws-resource list all`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {
	reporter := rprtr.CreateReporterOrExit()

	reporter.Infof("Listing all resources")

	if profile != "" || roleArn != "" {
		err := whoami.WhoAmICmd.RunE(cmd, args)
		if err != nil {
			reporter.Errorf("Unable to verify AWS account %s", err)
		}
	}

	err = ec2cmd.Cmd.RunE(cmd, args)
	if err != nil {
		reporter.Errorf("Unable to list EC2 instances: %s", err)
		return err
	}

	err = elbcmd.Cmd.RunE(cmd, args)
	if err != nil {
		reporter.Errorf("Unable to list ELB instances: %s", err)
		return err
	}

	err = elbv2cmd.Cmd.RunE(cmd, args)
	if err != nil {
		reporter.Errorf("Unable to list ELB V2 instances: %s", err)
		return err
	}

	err = route53cmd.Cmd.RunE(cmd, args)
	if err != nil {
		reporter.Errorf("Unable to list route53 hosted zones: %s", err)
		return err
	}

	err = snapshotscmd.Cmd.RunE(cmd, args)
	if err != nil {
		reporter.Errorf("Unable to list snapshots: %s", err)
		return err
	}

	err = volumescmd.Cmd.RunE(cmd, args)
	if err != nil {
		reporter.Errorf("Unable to list volumes: %s", err)
		return err
	}

	return
}

func init() {

	Cmd.Flags().BoolVar(&instanceNames, "instance-names", false, "Print instance names")
	Cmd.Flags().StringVarP(&roleArn, "role-arn", "a", "", "AWS Role to assume")
	Cmd.Flags().StringVarP(&profile, "profile", "p", "", "AWS Profile to use")
}
