package arguments

import (
	"github.com/spf13/pflag"
)

var (
	Region string
)

func AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&Region, "region", "r", "us-east-1", "AWS Region")
}
