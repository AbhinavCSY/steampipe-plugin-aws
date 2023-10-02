package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	cloudwatchlogsv1 "github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsCloudtrailTrailEvent(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_cloudtrail_trail_event",
		Description: "CloudTrail events from cloudwatch service.",
		List: &plugin.ListConfig{
			Hydrate:    listCloudwatchLogTrailEvents,
			Tags:       map[string]string{"service": "logs", "action": "FilterLogEvents"},
			KeyColumns: tableAwsCloudtrailEventsListKeyColumns(),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"ResourceNotFoundException"}),
			},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(cloudwatchlogsv1.EndpointsID),
		Columns: awsRegionalColumns([]*plugin.Column{
			// Top columns
			{Name: "filter", Type: proto.ColumnType_STRING, Transform: transform.FromQual("filter"), Description: "The cloudwatch filter pattern for the search."},
			{Name: "log_group_name", Type: proto.ColumnType_STRING, Transform: transform.FromQual("log_group_name"), Description: "The name of the log group to which this event belongs."},
			{Name: "log_stream_name", Type: proto.ColumnType_STRING, Description: "The name of the log stream to which this event belongs."},
			{Name: "timestamp", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("Timestamp").Transform(transform.UnixMsToTimestamp), Description: "The time when the event occurred."},
			{Name: "timestamp_ms", Type: proto.ColumnType_INT, Transform: transform.FromField("Timestamp"), Description: "The time when the event occurred."},

			// CloudTrail event fields
			{Name: "access_key_id", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Transform: transform.FromField("UserIdentity.AccessKeyId"), Description: "The AWS access key ID that was used to sign the request. If the request was made with temporary security credentials, this is the access key ID of the temporary credentials."},
			{Name: "aws_region", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The AWS region that the request was made to, such as us-east-2."},
			{Name: "error_code", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The AWS service error if the request returns an error."},
			{Name: "error_message", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "If the request returns an error, the description of the error."},
			{Name: "event_category", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "Shows the event category that is used in LookupEvents calls."},
			{Name: "event_id", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The ID of the event."},
			{Name: "event_name", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The name of the event returned."},
			{Name: "event_source", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The AWS service that the request was made to."},
			{Name: "event_time", Type: proto.ColumnType_TIMESTAMP, Hydrate: getCloudtrailMessageField, Description: "The date and time the request was made, in coordinated universal time (UTC)."},
			{Name: "event_type", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "Identifies the type of event that generated the event record."},
			{Name: "event_version", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The version of the log event format."},
			{Name: "read_only", Type: proto.ColumnType_BOOL, Hydrate: getCloudtrailMessageField, Description: "Information about whether the event is a write event or a read event."},
			{Name: "recipient_account_id", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "Represents the account ID that received this event."},
			{Name: "request_id", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The value that identifies the request."},
			{Name: "shared_event_id", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "GUID generated by CloudTrail to uniquely identify CloudTrail events from the same AWS action that is sent to different AWS accounts."},
			{Name: "source_ip_address", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The IP address that the request was made from."},
			{Name: "user_agent", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "The agent through which the request was made, such as the AWS Management Console, an AWS service, the AWS SDKs or the AWS CLI."},
			{Name: "user_type", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Transform: transform.FromField("UserIdentity.Type"), Description: "The name of the event returned."},
			{Name: "username", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Transform: transform.FromField("UserIdentity.Username"), Description: "The user name of the user that made the api request."},
			{Name: "user_identifier", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Transform: transform.FromField("UserIdentity.Arn", "UserIdentity.SessionContext.sessionIssuer.arn", "UserIdentity.SessionContext.sessionIssuer.principalId"), Description: "The name/arn of user/role that made the api call."},
			{Name: "vpc_endpoint_id", Type: proto.ColumnType_STRING, Hydrate: getCloudtrailMessageField, Description: "Identifies the VPC endpoint in which requests were made from a VPC to another AWS service, such as Amazon S3."},

			// Json fields
			{Name: "additional_event_data", Type: proto.ColumnType_JSON, Hydrate: getCloudtrailMessageField, Description: "Additional data about the event that was not part of the request or response."},
			{Name: "cloudtrail_event", Type: proto.ColumnType_JSON, Transform: transform.FromField("Message").Transform(trim).Transform(transform.UnmarshalYAML), Description: "The CloudTrail event in the json format."},
			{Name: "request_parameters", Type: proto.ColumnType_JSON, Hydrate: getCloudtrailMessageField, Description: "The parameters, if any, that were sent with the request."},
			{Name: "response_elements", Type: proto.ColumnType_JSON, Hydrate: getCloudtrailMessageField, Description: "The response element for actions that make changes (create, update, or delete actions)."},
			{Name: "resources", Type: proto.ColumnType_JSON, Hydrate: getCloudtrailMessageField, Description: "A list of resources referenced by the event returned."},
			{Name: "tls_details", Type: proto.ColumnType_JSON, Hydrate: getCloudtrailMessageField, Description: "Shows information about the Transport Layer Security (TLS) version, cipher suites, and the FQDN of the client-provided host name of a service API call."},
			{Name: "user_identity", Type: proto.ColumnType_JSON, Hydrate: getCloudtrailMessageField, Description: "Information about the user that made the request."},
		}),
	}
}

// https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference.html
// https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-record-contents.html
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/logging_cw_api_calls_cwe.html#cwe_info_in_ct
type cloudtrailEvent struct {
	AccountId *string `json:"accountId" type:"string"`

	// The AWS access key ID that was used to sign the request. If the request was made with temporary security credentials, this is the access key ID of the temporary credentials.
	AccessKeyId *string `json:"accessKeyId" type:"string"`

	// // A JSON string that contains a representation of the event returned.
	// CloudTrailEvent *string `json:"message" type:"string"`

	// The CloudTrail ID of the event returned.
	// A list of resources referenced by the event returned.
	Resources []*interface{} `json:"resources" type:"list"`

	// A user name or role name of the requester that called the API in the event
	// returned.
	Username *string `json:"userName" type:"string"`
	// contains filtered or unexported fields

	// Additional data about the event that was not part of the request or response.
	AdditionalEventData *interface{} `json:"additionalEventData" type:"map"`

	// Identifies the API version associated with the AwsApiCall eventType value.
	ApiVersion *string `json:"apiVersion"`

	// The AWS region that the request was made to, such as us-east-2.
	AwsRegion *string `json:"awsRegion" type:"string"`

	// Shows the event category that is used in LookupEvents calls.
	EventCategory *string `json:"eventCategory" type:"string"`

	// GUID generated by CloudTrail to uniquely identify each event.
	EventId *string `json:"eventID" type:"string"`

	// The name of the event returned.
	EventName *string `json:"eventName" type:"string"`

	// The AWS service that the request was made to.
	EventSource *string `json:"eventSource" type:"string"`

	// The date and time the request was made, in coordinated universal time (UTC).
	EventTime time.Time `json:"eventTime"`

	// Identifies the type of event that generated the event record.
	EventType *string `json:"eventType" type:"string"`

	// The version of the log event format. The current version is 1.08.
	EventVersion *string `json:"eventVersion" type:"string"`

	// The AWS service error if the request returns an error.
	ErrorCode *string `json:"errorCode"`

	// If the request returns an error, the description of the error.
	ErrorMessage *string `json:"errorMessage"`

	// A Boolean value that identifies whether the event is a management event.
	ManagementEvent *bool `json:"managementEvent" type:"bool"`

	// Represents the account ID that received this event.
	RecipientAccountId *string `json:"recipientAccountId" type:"string"`

	// The response element for actions that make changes (create, update, or delete actions).
	ResponseElements *interface{} `json:"responseElements" type:"map"`

	// GUID generated by CloudTrail to uniquely identify CloudTrail
	// events from the same AWS action that is sent to different AWS accounts.
	SharedEventId *string `json:"sharedEventID" type:"string"`

	// The IP address that the request was made from.
	SourceIpAddress *string `json:"sourceIPAddress" type:"string"`

	// Identifies whether this operation is a read-only operation.
	ReadOnly *bool `json:"readOnly" type:"bool"`

	// The value that identifies the request.
	RequestId *string `json:"requestID" type:"string"`

	// The parameters, if any, that were sent with the request.
	RequestParameters *interface{} `json:"requestParameters" type:"map"`

	// The agent through which the request was made,
	// such as the AWS Management Console, an AWS service,
	// the AWS SDKs or the AWS CLI.
	UserAgent *string `json:"userAgent" type:"string"`

	// Information about the user that made a request. For
	// UserIdentity *interface{} `json:"userIdentity" type:"map"`
	UserIdentity userIdentity `json:"userIdentity"`

	// Identifies the VPC endpoint in which requests were
	// made from a VPC to another AWS service, such as Amazon S3.
	VpcEndpointId *string `json:"vpcEndpointId" type:"string"`

	// Identifies the service event, including what triggered the event and the result.
	ServiceEventDetails *interface{} `json:"serviceEventDetails"`

	// If an event delivery was delayed, or additional information
	// about an existing event becomes available after the event is
	// logged, an addendum field shows information about why the event was delayed.
	Addendum *interface{} `json:"addendum"`

	// Shows information about edge devices that are targets of a request. Currently, S3 Outposts device events include this field.
	EdgeDeviceDetails *interface{} `json:"edgeDeviceDetails"`

	// Shows whether or not an event originated from a AWS Management Console session.
	SessionCredentialFromConsole *interface{} `json:"sessionCredentialFromConsole"`

	// Shows information about the Transport Layer Security (TLS) version, cipher suites, and the FQDN of the client-provided host name of a service API call.
	TlsDetails *interface{} `json:"tlsDetails"`
}

// https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-user-identity.html
type userIdentity struct {
	Type                string       `json:"type"`
	Username            string       `json:"userName"`
	PrincipalId         string       `json:"principalId"`
	Arn                 string       `json:"arn"`
	AccountId           string       `json:"accountId"`
	AccessKeyId         string       `json:"accessKeyId"`
	IdentityProvider    string       `json:"identityProvider"`
	InvokedBy           string       `json:"invokedBy"`
	SessionContext      *interface{} `json:"sessionContext"`
	WebIdFederationData *interface{} `json:"webIdFederationData"`
}

// https://docs.aws.amazon.com/AWSJavaScriptSDK/latest/AWS/CloudTrail.html#lookupEvents-property
func tableAwsCloudtrailEventsListKeyColumns() []*plugin.KeyColumn {
	return []*plugin.KeyColumn{
		// CloudWatch fields
		{Name: "log_group_name"},
		{Name: "log_stream_name", Require: plugin.Optional},
		{Name: "filter", Require: plugin.Optional, CacheMatch: "exact"},
		{Name: "region", Require: plugin.Optional},
		{Name: "timestamp", Operators: []string{">", ">=", "=", "<", "<="}, Require: plugin.Optional},

		// event fields
		{Name: "event_category", Require: plugin.Optional},
		{Name: "event_id", Require: plugin.Optional},
		{Name: "aws_region", Require: plugin.Optional},
		{Name: "source_ip_address", Require: plugin.Optional},
		{Name: "error_code", Require: plugin.Optional},
		{Name: "event_name", Require: plugin.Optional},
		{Name: "read_only", Require: plugin.Optional},
		{Name: "username", Require: plugin.Optional},
		{Name: "user_type", Require: plugin.Optional},
		{Name: "event_source", Require: plugin.Optional},
		{Name: "access_key_id", Require: plugin.Optional},
	}
}

func listCloudwatchLogTrailEvents(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	// Create session
	svc, err := CloudWatchLogsClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_cloudtrail_trail_event.listCloudwatchLogTrailEvents", "connection_error", err)
		return nil, err
	}

	equalQuals := d.EqualsQuals
	// quals := d.Quals

	// Limiting the results
	maxLimit := int32(10000)
	if d.QueryContext.Limit != nil {
		limit := int32(*d.QueryContext.Limit)
		if limit < maxLimit {
			if limit < 1 {
				maxLimit = 1
			} else {
				maxLimit = limit
			}
		}
	}

	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(equalQuals["log_group_name"].GetStringValue()),
		// Default to the maximum allowed
		Limit: aws.Int32(maxLimit),
	}

	if equalQuals["log_stream_name"] != nil {
		input.LogStreamNames = []string{equalQuals["log_stream_name"].GetStringValue()}
	}

	queryFilter := ""
	filter := buildQueryFilter(equalQuals)

	if equalQuals["filter"] != nil {
		queryFilter = equalQuals["filter"].GetStringValue()
	}

	if queryFilter != "" {
		input.FilterPattern = aws.String(queryFilter)
	} else if len(filter) > 0 {
		input.FilterPattern = aws.String(fmt.Sprintf("{ %s }", strings.Join(filter, " && ")))
	}

	quals := d.Quals

	if quals["timestamp"] != nil {
		for _, q := range quals["timestamp"].Quals {
			tsSecs := q.Value.GetTimestampValue().GetSeconds()
			tsMs := tsSecs * 1000
			switch q.Operator {
			case "=":
				input.StartTime = aws.Int64(tsMs)
				input.EndTime = aws.Int64(tsMs)
			case ">=", ">":
				input.StartTime = aws.Int64(tsMs)
			case "<", "<=":
				input.EndTime = aws.Int64(tsMs)
			}
		}
	}

	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(svc, input, func(o *cloudwatchlogs.FilterLogEventsPaginatorOptions) {
		o.Limit = maxLimit
		o.StopOnDuplicateToken = true
	})
	// List call
	for paginator.HasMorePages() {

		// apply rate limiting
		d.WaitForListRateLimit(ctx)

		output, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("aws_cloudtrail_trail_event.listCloudwatchLogTrailEvents", "api_error", err)
			return nil, err
		}

		for _, items := range output.Events {
			d.StreamListItem(ctx, items)

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, err
}

func getCloudtrailMessageField(ctx context.Context, _ *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	e := h.Item.(types.FilteredLogEvent)
	cte := cloudtrailEvent{}
	err := json.Unmarshal([]byte(*e.Message), &cte)
	if err != nil {
		return nil, err
	}
	return cte, nil
}

func buildQueryFilter(equalQuals plugin.KeyColumnEqualsQualMap) []string {
	filters := []string{}

	filterQuals := map[string]string{
		"access_key_id":     "userIdentity.accessKeyId",
		"aws_region":        "awsRegion",
		"error_code":        "errorCode",
		"event_category":    "eventCategory",
		"event_id":          "eventID",
		"event_name":        "eventName",
		"event_source":      "eventSource",
		"read_only":         "readOnly",
		"source_ip_address": "sourceIPAddress",
		"username":          "userIdentity.userName",
		"user_type":         "userIdentity.type",
	}

	for qual, filterKey := range filterQuals {
		if equalQuals[qual] != nil {
			filters = append(filters, fmt.Sprintf("( $.%s = \"%s\" )", filterKey, equalQuals[qual].GetStringValue()))
		}
	}

	return filters
}
