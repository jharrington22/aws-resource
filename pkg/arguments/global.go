package arguments

import (
	"github.com/spf13/pflag"
)

var (
	Region  string
	Profile string
	RoleArn string
)

func AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&Region, "region", "r", "us-east-1", "AWS Region")
	fs.StringVarP(&Profile, "profile", "p", "", "AWS Profile")
	fs.StringVarP(&RoleArn, "role-arn", "a", "", "AWS IAM Role ARN")
}
