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
package whoami

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var WhoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "Get AWS account information",
	Long:  `Get AWS account information by querying STS`,
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
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
