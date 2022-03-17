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
package route53

import (
	"os"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var Cmd = &cobra.Command{
	Use:   "route53",
	Short: "List route53 resources",
	Long: `List route53 resources"

aws-resource list route53.`,
	Run: func(cmd *cobra.Command, args []string) {

		reporter := rprtr.CreateReporterOrExit()
		logging := logging.CreateLoggerOrExit(reporter)

		region := "us-east-1"

		awsClient, err := aws.NewClient().
			Logger(logging).
			Region(region).
			Build()

		if err != nil {
			reporter.Errorf("Unable to build AWS client")
			os.Exit(1)
		}

		input := &route53.ListHostedZonesByNameInput{}

		result, err := awsClient.ListHostedZonesByName(input)

		hostedZonesCount := 0
		for _, _ = range result.HostedZones {
			hostedZonesCount++
		}
		reporter.Infof("Found %d hosted zoned", hostedZonesCount)
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
