package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/sirupsen/logrus"
)

var (
	defaultAWSRegion = "us-east-1"
)

type Client interface {
	DeregisterImage(input *ec2.DeregisterImageInput) (*ec2.DeregisterImageOutput, error)
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
	DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeV2LoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
	DescribeSnapshots(input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error)
	DescribeSnapshotsPages(input *ec2.DescribeSnapshotsInput, fn func(*ec2.DescribeSnapshotsOutput, bool) bool) error
	DeleteSnapshot(input *ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error)
	DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
	GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
	ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error)
}

type ClientBuilder struct {
	logger      *logrus.Logger
	region      *string
	profile     *string
	roleArn     *string
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

func (b *ClientBuilder) RoleArn(value string) *ClientBuilder {
	b.roleArn = aws.String(value)
	return b
}

// Create AWS session with a specific set of credentials
func (b *ClientBuilder) BuildSessionWithOptionsCredentials(value *credentials.Value) (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        b.region,
			Credentials:                   credentials.NewStaticCredentials(value.AccessKeyID, value.SecretAccessKey, ""),
			Retryer: client.DefaultRetryer{
				MinRetryDelay: 1 * time.Second,
			},
		},
	},
	)
}

func (b *ClientBuilder) BuildSessionWithOptions() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        b.region,
		},
	})
}

func (b *ClientBuilder) BuildSessionWithProfileOptions() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           *b.profile,
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
	}
	if b.profile != nil && *b.profile != "" {
		sess, err = b.BuildSessionWithProfileOptions()
	} else {
		sess, err = b.BuildSessionWithOptions()
	}
	if err != nil {
		return nil, err
	}

	if b.roleArn != nil {
		if *b.roleArn != "" {
			assumeRoleCreds := stscreds.NewCredentials(sess, *b.roleArn)
			return &awsClient{
				logger:        b.logger,
				ec2Client:     ec2.New(sess, &aws.Config{Credentials: assumeRoleCreds}),
				elbClient:     elb.New(sess, &aws.Config{Credentials: assumeRoleCreds}),
				elbV2Client:   elbv2.New(sess, &aws.Config{Credentials: assumeRoleCreds}),
				iamClient:     iam.New(sess, &aws.Config{Credentials: assumeRoleCreds}),
				route53Client: route53.New(sess, &aws.Config{Credentials: assumeRoleCreds}),
				stsClient:     sts.New(sess, &aws.Config{Credentials: assumeRoleCreds}),
			}, nil
		}
	}

	return &awsClient{
		logger:        b.logger,
		ec2Client:     ec2.New(sess),
		elbClient:     elb.New(sess),
		elbV2Client:   elbv2.New(sess),
		iamClient:     iam.New(sess),
		route53Client: route53.New(sess),
		stsClient:     sts.New(sess),
	}, nil
}

type awsClient struct {
	logger        *logrus.Logger
	ec2Client     ec2iface.EC2API
	elbClient     elbiface.ELBAPI
	elbV2Client   elbv2iface.ELBV2API
	iamClient     iamiface.IAMAPI
	route53Client route53iface.Route53API
	stsClient     stsiface.STSAPI
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
		}
		return nil, fmt.Errorf("describe snpshots failed, %s", err)
	}

	return result, nil
}

func (c *awsClient) DescribeSnapshotsPages(input *ec2.DescribeSnapshotsInput, fn func(*ec2.DescribeSnapshotsOutput, bool) bool) error {
	err := c.ec2Client.DescribeSnapshotsPages(input, fn)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return aerr
			}
		}
		return fmt.Errorf("describe snapshots failed, %s", err)
	}
	return nil
}

func (c *awsClient) DeleteSnapshot(input *ec2.DeleteSnapshotInput) (output *ec2.DeleteSnapshotOutput, err error) {
	output, err = c.ec2Client.DeleteSnapshot(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			return nil, fmt.Errorf("describe snapshots failed, %s", err)
		}
	}
	return output, nil
}

func (c *awsClient) DescribeImages(input *ec2.DescribeImagesInput) (output *ec2.DescribeImagesOutput, err error) {
	output, err = c.ec2Client.DescribeImages(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			return nil, fmt.Errorf("describe images failed, %s", err)
		}
	}
	return output, nil
}

func (c *awsClient) DeregisterImage(input *ec2.DeregisterImageInput) (output *ec2.DeregisterImageOutput, err error) {
	output, err = c.ec2Client.DeregisterImage(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return nil, aerr
			}
		} else {
			return nil, fmt.Errorf("deregister images failed, %s", err)
		}
	}
	return output, nil
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

func (c *awsClient) ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {

	result, err := c.route53Client.ListHostedZonesByName(input)
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
