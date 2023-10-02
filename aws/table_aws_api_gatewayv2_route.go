package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"

	apigatewayv2v1 "github.com/aws/aws-sdk-go/service/apigatewayv2"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsAPIGatewayV2Route(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_api_gatewayv2_route",
		Description: "AWS API Gateway Version 2 Route",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"route_id", "api_id"}),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"NotFoundException"}),
			},
			Hydrate: getAPIGatewayV2Route,
			Tags:    map[string]string{"service": "apigateway", "action": "GetRoute"},
		},
		List: &plugin.ListConfig{
			ParentHydrate: listAPIGatewayV2API,
			Hydrate:       listAPIGatewayV2Routes,
			Tags:          map[string]string{"service": "apigateway", "action": "GetRoutes"},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(apigatewayv2v1.EndpointsID),
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "route_key",
				Description: "The route key for the route.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.RouteKey"),
			},
			{
				Name:        "api_id",
				Description: "Represents the identifier of an API.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "route_id",
				Description: "The route ID.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.RouteId"),
			},
			{
				Name:        "api_gateway_managed",
				Description: "Specifies whether a route is managed by API Gateway.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("GetRouteOutput.ApiGatewayManaged"),
			},
			{
				Name:        "api_key_required",
				Description: "Specifies whether an API key is required for this route. Supported only for WebSocket APIs.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("GetRouteOutput.ApiKeyRequired"),
			},
			{
				Name:        "authorization_type",
				Description: "The authorization type for the route. For WebSocket APIs, valid values are NONE for open access, AWS_IAM for using AWS IAM permissions, and CUSTOM for using a Lambda authorizer For HTTP APIs, valid values are NONE for open access, JWT for using JSON Web Tokens, AWS_IAM for using AWS IAM permissions, and CUSTOM for using a Lambda authorizer.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.AuthorizationType"),
			},
			{
				Name:        "authorizer_id",
				Description: "The identifier of the Authorizer resource to be associated with this route. The authorizer identifier is generated by API Gateway when you created the authorizer.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.AuthorizerId"),
			},
			{
				Name:        "model_selection_expression",
				Description: "The model selection expression for the route. Supported only for WebSocket APIs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.ModelSelectionExpression"),
			},
			{
				Name:        "operation_name",
				Description: "The operation name for the route.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.OperationName"),
			},
			{
				Name:        "route_response_selection_expression",
				Description: "The route response selection expression for the route. Supported only for WebSocket APIs.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.RouteResponseSelectionExpression"),
			},
			{
				Name:        "target",
				Description: "The target for the route.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.Target"),
			},
			{
				Name:        "authorization_scopes",
				Description: "A list of authorization scopes configured on a route. The scopes are used with a JWT authorizer to authorize the method invocation. The authorization works by matching the route scopes against the scopes parsed from the access token in the incoming request. The method invocation is authorized if any route scope matches a claimed scope in the access token. Otherwise, the invocation is not authorized. When the route scope is configured, the client must provide an access token instead of an identity token for authorization purposes.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("GetRouteOutput.AuthorizationScopes"),
			},
			{
				Name:        "request_models",
				Description: "The request models for the route. Supported only for WebSocket APIs.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("GetRouteOutput.RequestModels"),
			},
			{
				Name:        "request_parameters",
				Description: "The request parameters for the route. Supported only for WebSocket APIs.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("GetRouteOutput.RequestParameters"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("GetRouteOutput.RouteId"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAPIGatewayV2RouteARN,
				Transform:   transform.FromValue().Transform(transform.EnsureStringArray),
			},
		}),
	}
}

//// LIST FUNCTION

func listAPIGatewayV2Routes(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get API details
	api := h.Item.(types.Api)

	// Create Session
	svc, err := APIGatewayV2Client(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gatewayv2_route.listAPIGatewayV2Routes", "connection_error", err)
		return nil, err
	}
	if svc == nil {
		// Unsupported region, return no data
		return nil, nil
	}

	// Limiting the results
	maxLimit := int32(500)
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

	pagesLeft := true
	params := &apigatewayv2.GetRoutesInput{
		ApiId:      api.ApiId,
		MaxResults: aws.String(fmt.Sprint(maxLimit)),
	}

	for pagesLeft {
		// apply rate limiting
		d.WaitForListRateLimit(ctx)

		result, err := svc.GetRoutes(ctx, params)
		if err != nil {
			plugin.Logger(ctx).Error("aws_api_gatewayv2_route.listAPIGatewayV2Routes", "api_error", err)
			return nil, err
		}

		for _, route := range result.Items {

			routeOp := &apigatewayv2.GetRouteOutput{
				ApiGatewayManaged:                route.ApiGatewayManaged,
				ApiKeyRequired:                   route.ApiKeyRequired,
				AuthorizationScopes:              route.AuthorizationScopes,
				AuthorizationType:                route.AuthorizationType,
				AuthorizerId:                     route.AuthorizerId,
				ModelSelectionExpression:         route.ModelSelectionExpression,
				OperationName:                    route.OperationName,
				RequestModels:                    route.RequestModels,
				RequestParameters:                route.RequestParameters,
				RouteId:                          route.RouteId,
				RouteKey:                         route.RouteKey,
				RouteResponseSelectionExpression: route.RouteResponseSelectionExpression,
				Target:                           route.Target,
			}
			d.StreamLeafListItem(ctx, &RouteInfo{
				ApiId:          *api.ApiId,
				GetRouteOutput: routeOp,
			})

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if result.NextToken != nil {
			pagesLeft = true
			params.NextToken = result.NextToken
		} else {
			pagesLeft = false
		}
	}

	return nil, nil
}

type RouteInfo struct {
	ApiId string
	*apigatewayv2.GetRouteOutput
}

//// HYDRATE FUNCTIONS

func getAPIGatewayV2Route(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	// Create Session
	svc, err := APIGatewayV2Client(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gatewayv2_route.getAPIGatewayV2Route", "service_client_error", err)
		return nil, err
	}
	if svc == nil {
		// Unsupported region, return no data
		return nil, nil
	}

	api := d.EqualsQuals["api_id"].GetStringValue()
	routeId := d.EqualsQuals["route_id"].GetStringValue()
	params := &apigatewayv2.GetRouteInput{
		ApiId:   aws.String(api),
		RouteId: aws.String(routeId),
	}

	item, err := svc.GetRoute(ctx, params)

	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gatewayv2_route.getAPIGatewayV2Route", "api_error", err)
		return nil, err
	}

	if item != nil {
		return RouteInfo{api, item}, nil
	}

	return nil, nil
}

func getAPIGatewayV2RouteARN(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	data := h.Item.(*RouteInfo)
	region := d.EqualsQualString(matrixKeyRegion)
	commonData, err := getCommonColumns(ctx, d, h)
	if err != nil {
		return nil, err
	}

	commonColumnData := commonData.(*awsCommonColumnData)
	// arn:partition:apigateway:region::/apis/api-id/routes/id
	arn := "arn:" + commonColumnData.Partition + ":apigateway:" + region + "::/apis/" + data.ApiId + "/routes/" + *data.RouteId

	return arn, nil
}
