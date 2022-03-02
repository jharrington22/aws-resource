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
package assumeRole

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

const (
	defaultOrganizationRoleArn = "OrganizationAccountAccessRole"
	EnvAwsAccessKeyId          = "AWS_ACCESS_KEY_ID"
	EnvAwsSecretAccessKey      = "AWS_SECRET_ACCESS_KEY"
	EnvAwsSessionToken         = "AWS_SESSION_TOKEN"
)

var roleSessionDuration int64 = 3600
var roleSessionName string = "awsResourceCLI"

var args struct {
	// AWS IAM Role ARN to assume, overwrites the default
	roleArn string
}

// listCmd represents the list command
var Cmd = &cobra.Command{
	Use:   "assume-role",
	Short: "Assume AWS IAM role",
	Long: `Assume AWS IAM role
	This command will attempt to assume the default OrganizationAccessRole that is created for accounts
	under AWS Organizations

	The command takes a --role-arn flag if you want to overwrite the above behavior.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		reporter := rprtr.CreateReporterOrExit()
		logging := logging.CreateLoggerOrExit(reporter)

		region := "us-east-1"

		awsClient, err := aws.NewClient().
			Logger(logging).
			Region(region).
			Build()

		if err != nil {
			fmt.Errorf("Unable to build AWS client")
			os.Exit(1)
		}

		identity, err := awsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			fmt.Errorf("Error %s", err)
			os.Exit(1)
		}

		fmt.Printf("AWS Account: %s\n", *identity.Account)

		role := defaultOrganizationRoleArn

		// if args.roleArn != "" {
		// 	role = args.roleArn
		// }

		input := &sts.AssumeRoleInput{
			DurationSeconds: &roleSessionDuration,
			RoleArn:         &role,
			RoleSessionName: &roleSessionName,
		}

		assumeRoleOutput := &sts.AssumeRoleOutput{}

		defaultSleepDelay := 500 * time.Millisecond

		// Should use a waiter here
		for i := 0; i < 100; i++ {
			time.Sleep(defaultSleepDelay)
			assumeRoleOutput, err = awsClient.AssumeRole(input)
			if err == nil {
				break
			}
			if i == 99 {
				logging.Info(fmt.Sprintf("Timed out while assuming role %s", role))
			}
		}
		if err != nil {
			// Log AWS error
			if aerr, ok := err.(awserr.Error); ok {
				logging.Error(aerr,
					fmt.Sprintf(`New AWS Error while getting STS credentials,
                        AWS Error Code: %s,
                        AWS Error Message: %s`,
						aerr.Code(),
						aerr.Message()))
			}
			os.Exit(1)
		}
		if err != nil {
			fmt.Errorf("Unable to assume role: %s", err)
			os.Exit(1)
		}

		// Set environment varables to be used by CLI
		err = os.Setenv(EnvAwsAccessKeyId, *assumeRoleOutput.Credentials.AccessKeyId)
		if err != nil {
			logging.Errorf("Failed to set ENV var %s: %s", EnvAwsAccessKeyId, err)
		}

		err = os.Setenv(EnvAwsSecretAccessKey, *assumeRoleOutput.Credentials.SecretAccessKey)
		if err != nil {
			logging.Errorf("Failed to set ENV var %s: %s", EnvAwsSecretAccessKey, err)
		}

		err = os.Setenv(EnvAwsSessionToken, *assumeRoleOutput.Credentials.SessionToken)
		if err != nil {
			logging.Errorf("Failed to set ENV var %s: %s", EnvAwsSessionToken, err)
		}

		awsClient, err = aws.NewClient().
			Logger(logging).
			Region(region).
			Build()

		if err != nil {
			fmt.Errorf("Unable to build AWS client")
			os.Exit(1)
		}

		identity, err = awsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			fmt.Errorf("Error %s", err)
			os.Exit(1)
		}

		fmt.Printf("AWS Account: %s\n", *identity.Account)
	},
}

func init() {

	// AWS IAM Role ARN to assume, overwrites the default
	Cmd.Flags().StringVarP(
		&args.roleArn,
		"role-arn",
		"r",
		"",
		"AWS IAM role ARN to assume",
	)
}
