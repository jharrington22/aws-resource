package route53

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jharrington22/aws-resource/pkg/arguments"
	"github.com/jharrington22/aws-resource/pkg/aws"
	logging "github.com/jharrington22/aws-resource/pkg/logging"
	rprtr "github.com/jharrington22/aws-resource/pkg/reporter"
	"github.com/spf13/cobra"
)

var (
	zoneId string
)

var Cmd = &cobra.Command{
	Use:   "route53",
	Short: "Delete route53 resources",
	Long: `Delete route53 resources"
aws-resource delete route53.`,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) (err error) {

	reporter := rprtr.CreateReporterOrExit()
	logging := logging.CreateLoggerOrExit(reporter)

	awsClient, err := aws.NewClient().
		Logger(logging).
		Profile(arguments.Profile).
		RoleArn(arguments.RoleArn).
		Build()

	if err != nil {
		reporter.Errorf("Unable to build AWS client")
		return err
	}
	if err != nil {
		reporter.Errorf("Failed to describe regions")
		return err
	}

	if zoneId != "" {
		err := deleteHostedZoneId(awsClient, zoneId, reporter)
		if err != nil {
			reporter.Errorf("Unable to delete hosted zone: %s", err)
		}
	} else {
		reporter.Infof("No zone id specified")
		err := deleteAllHostedZones(reporter, awsClient)
		if err != nil {
			reporter.Errorf("Unable to delete hosted zones: %s", err)
		}
	}
	return
}

func init() {
	flags := Cmd.Flags()
	arguments.AddFlags(flags)
	Cmd.Flags().StringVarP(&zoneId, "zone-id", "i", "", "Delete specific zone id")
}

func deleteHostedZoneId(client aws.Client, zoneId string, reporter *rprtr.Object) error {

	input := &route53.DeleteHostedZoneInput{
		Id: &zoneId,
	}

	_, err := client.DeleteHostedZonesByName(input)
	if err != nil {
		return fmt.Errorf("unable to delete hosted zone: %s", err)
	}
	reporter.Infof("Hosted zone %s deleted", zoneId)
	return err
}

func deleteAllHostedZones(
	reporter *rprtr.Object, awsClient aws.Client) error {

	input := &route53.ListHostedZonesByNameInput{}

	output, err := awsClient.ListHostedZonesByName(input)
	if err != nil {
		reporter.Errorf("Unable to describe hosted zones %s", err)
		return err
	}

	for _, zone := range output.HostedZones {
		err = deleteHostedZoneId(awsClient, *zone.Id, reporter)
		if err != nil {
			reporter.Errorf("Unable to delete hosted zone %s: %s", *zone.Id, err)
		}
	}
	return nil
}
