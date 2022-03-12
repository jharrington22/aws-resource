package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/sirupsen/logrus"
)

var (
	defaultAWSRegion = "us-east-1"
)

type Client interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeV2LoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
	DescribeSnapshots(input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}

type ClientBuilder struct {
	logger      *logrus.Logger
	region      *string
	profile     *string
	credentials *credentials.Value
}

func NewClient() *ClientBuilder {
	return &ClientBuilder{}
}

func (b *ClientBuilder) Logger(value *logrus.Logger) *ClientBuilder {
	b.logger = value
	return b
}

func (b *ClientBuilder) Profile(value string) *ClientBuilder {
	b.profile = aws.String(value)
	return b
}

func (b *ClientBuilder) Region(value string) *ClientBuilder {
	b.region = aws.String(value)
	return b
}

// Create AWS session with a specific set of credentials
func (b *ClientBuilder) BuildSessionWithOptionsCredentials(value *credentials.Value) (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        b.region,
			Credentials:                   credentials.NewStaticCredentials(value.AccessKeyID, value.SecretAccessKey, ""),
			RequestRetryer: 					client.DefaultRetryer{
				MinRetryDelay: 1 * time.second,
			},
		},
	}
}

func (b *ClientBuilder) BuildSessionWithOptions() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		// Profile:           *b.profile,
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        b.region,
		},
	})
}

func (b *ClientBuilder) Build() (Client, error) {
	var err error

	if b.logger == nil {
		return nil, fmt.Errorf("Logger is required")
	}

	var sess *session.Session

	if b.region == nil || *b.region == "" {
		fmt.Printf("No region set using %s\n", defaultAWSRegion)
		b.region = aws.String(defaultAWSRegion)
	}

	// Create the AWS session:
	if b.credentials != nil {
		sess, err = b.BuildSessionWithOptionsCredentials(b.credentials)
	} else {
		sess, err = b.BuildSessionWithOptions()
	}
	if err != nil {
		return nil, err
	}

	c := &awsClient{
		logger:      b.logger,
		ec2Client:   ec2.New(sess),
		elbClient:   elb.New(sess),
		elbV2Client: elbv2.New(sess),
		iamClient:   iam.New(sess),
		stsClient:   sts.New(sess),
	}

	return c, err
}

type awsClient struct {
	logger      *logrus.Logger
	ec2Client   ec2iface.EC2API
	elbClient   elbiface.ELBAPI
	elbV2Client elbv2iface.ELBV2API
	iamClient   iamiface.IAMAPI
	stsClient   stsiface.STSAPI
}

func (c *awsClient) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

	result, err := c.ec2Client.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("describe instances failed, %s", err)
	}

	return result, nil

}

func (c *awsClient) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {

	result, err := c.elbClient.DescribeLoadBalancers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("describe load balancers failed, %s", err)
	}

	return result, nil

}

func (c *awsClient) DescribeV2LoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {

	result, err := c.elbV2Client.DescribeLoadBalancers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("describe v2 load balancers failed, %s", err)
	}

	return result, nil

}

func (c *awsClient) DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
	result, err := c.ec2Client.DescribeRegions(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("describe regions failed, %s", err)
	}

	return result, nil
}

func (c *awsClient) DescribeSnapshots(input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error) {
	result, err := c.ec2Client.DescribeSnapshots(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("describe snpshots failed, %s", err)
	}

	return result, nil
}

func (c *awsClient) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	result, err := c.ec2Client.DescribeVolumes(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("describe volumes failed, %s", err)
	}

	return result, nil
}

func (c *awsClient) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	result, err := c.stsClient.GetCallerIdentity(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return nil, err
		}
		return nil, fmt.Errorf("getting sts caller identity failed, %s", err)
	}

	return result, nil
}
