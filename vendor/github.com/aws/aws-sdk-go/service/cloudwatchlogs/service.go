// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package cloudwatchlogs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/private/protocol/jsonrpc"
	"github.com/aws/aws-sdk-go/private/signer/v4"
)

// This is the Amazon CloudWatch Logs API Reference. Amazon CloudWatch Logs
// enables you to monitor, store, and access your system, application, and custom
// log files. This guide provides detailed information about Amazon CloudWatch
// Logs actions, data types, parameters, and errors. For detailed information
// about Amazon CloudWatch Logs features and their associated API calls, go
// to the Amazon CloudWatch Developer Guide (http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide).
//
// Use the following links to get started using the Amazon CloudWatch Logs
// API Reference:
//
//  Actions (http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_Operations.html):
// An alphabetical list of all Amazon CloudWatch Logs actions. Data Types (http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_Types.html):
// An alphabetical list of all Amazon CloudWatch Logs data types. Common Parameters
// (http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/CommonParameters.html):
// Parameters that all Query actions can use. Common Errors (http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/CommonErrors.html):
// Client and server errors that all actions can return. Regions and Endpoints
// (http://docs.aws.amazon.com/general/latest/gr/index.html?rande.html): Itemized
// regions and endpoints for all AWS products.  In addition to using the Amazon
// CloudWatch Logs API, you can also use the following SDKs and third-party
// libraries to access Amazon CloudWatch Logs programmatically.
//
//  AWS SDK for Java Documentation (http://aws.amazon.com/documentation/sdkforjava/)
// AWS SDK for .NET Documentation (http://aws.amazon.com/documentation/sdkfornet/)
// AWS SDK for PHP Documentation (http://aws.amazon.com/documentation/sdkforphp/)
// AWS SDK for Ruby Documentation (http://aws.amazon.com/documentation/sdkforruby/)
//  Developers in the AWS developer community also provide their own libraries,
// which you can find at the following AWS developer centers:
//
//  AWS Java Developer Center (http://aws.amazon.com/java/) AWS PHP Developer
// Center (http://aws.amazon.com/php/) AWS Python Developer Center (http://aws.amazon.com/python/)
// AWS Ruby Developer Center (http://aws.amazon.com/ruby/) AWS Windows and .NET
// Developer Center (http://aws.amazon.com/net/)
//The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type CloudWatchLogs struct {
	*client.Client
}

// Used for custom client initialization logic
var initClient func(*client.Client)

// Used for custom request initialization logic
var initRequest func(*request.Request)

// A ServiceName is the name of the service the client will make API calls to.
const ServiceName = "logs"

// New creates a new instance of the CloudWatchLogs client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a CloudWatchLogs client from just a session.
//     svc := cloudwatchlogs.New(mySession)
//
//     // Create a CloudWatchLogs client with additional configuration
//     svc := cloudwatchlogs.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func New(p client.ConfigProvider, cfgs ...*aws.Config) *CloudWatchLogs {
	c := p.ClientConfig(ServiceName, cfgs...)
	return newClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion)
}

// newClient creates, initializes and returns a new service client instance.
func newClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion string) *CloudWatchLogs {
	svc := &CloudWatchLogs{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   ServiceName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "2014-03-28",
				JSONVersion:   "1.1",
				TargetPrefix:  "Logs_20140328",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBack(v4.Sign)
	svc.Handlers.Build.PushBackNamed(jsonrpc.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(jsonrpc.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(jsonrpc.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(jsonrpc.UnmarshalErrorHandler)

	// Run custom client initialization if present
	if initClient != nil {
		initClient(svc.Client)
	}

	return svc
}

// newRequest creates a new request for a CloudWatchLogs operation and runs any
// custom request initialization.
func (c *CloudWatchLogs) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	// Run custom request initialization if present
	if initRequest != nil {
		initRequest(req)
	}

	return req
}
