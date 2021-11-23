package aws

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func tableAwsLambdaAlias(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_lambda_alias",
		Description: "AWS Lambda Alias",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.AllColumns([]string{"name", "function_name", "region"}),
			ShouldIgnoreError: isNotFoundError([]string{"InvalidParameter", "ResourceNotFoundException"}),
			Hydrate:           getLambdaAlias,
		},
		List: &plugin.ListConfig{
			ParentHydrate: listAwsLambdaFunctions,
			Hydrate:       listLambdaAliases,
		},
		GetMatrixItem: BuildRegionList,
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the alias.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Alias.Name"),
			},
			{
				Name:        "function_name",
				Description: "The name of the function.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "alias_arn",
				Description: "The Amazon Resource Name (ARN) of the alias.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Alias.AliasArn"),
			},
			{
				Name:        "function_version",
				Description: "The function version that the alias invokes.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Alias.FunctionVersion"),
			},
			{
				Name:        "revision_id",
				Description: "A unique identifier that changes when you update the alias.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Alias.RevisionId"),
			},
			{
				Name:        "description",
				Description: "A description of the alias.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Alias.Description"),
			},
			{
				Name:        "policy",
				Description: "Contains the resource-based policy.",
				Hydrate:     getLambdaAliasPolicy,
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "policy_std",
				Description: "Contains the contents of the resource-based policy in a canonical form for easier searching.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getLambdaAliasPolicy,
				Transform:   transform.FromField("Policy").Transform(unescape).Transform(policyToCanonical),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Alias.Name"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Alias.AliasArn").Transform(arnToAkas),
			},
		}),
	}
}

type aliasRowData = struct {
	Alias        *lambda.AliasConfiguration
	FunctionName *string
}

//// LIST FUNCTION

func listLambdaAliases(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("listLambdaAliases")

	svc, err := LambdaService(ctx, d)
	if err != nil {
		return nil, err
	}

	function := h.Item.(*lambda.FunctionConfiguration)

	err = svc.ListAliasesPages(
		&lambda.ListAliasesInput{FunctionName: function.FunctionName},
		func(page *lambda.ListAliasesOutput, lastPage bool) bool {
			for _, alias := range page.Aliases {
				d.StreamLeafListItem(ctx, &aliasRowData{alias, function.FunctionName})
			}
			return !lastPage
		},
	)

	return nil, err
}

//// HYDRATE FUNCTIONS

func getLambdaAlias(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	matrixRegion := d.KeyColumnQualString(matrixKeyRegion)
	plugin.Logger(ctx).Trace("getLambdaAlias")

	name := d.KeyColumnQuals["name"].GetStringValue()
	functionName := d.KeyColumnQuals["function_name"].GetStringValue()
	region := d.KeyColumnQuals["region"].GetStringValue()

	// Empty check
	if name == "" || functionName == "" || region != matrixRegion {
		return nil, nil
	}

	// Create Session
	svc, err := LambdaService(ctx, d)
	if err != nil {
		return nil, err
	}

	// Build params
	params := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(name),
	}

	rowData, err := svc.GetAlias(params)
	if err != nil {
		plugin.Logger(ctx).Debug("getLambdaAlias__", "ERROR", err)
		return nil, err
	}

	return &aliasRowData{rowData, aws.String(functionName)}, nil
}

func getLambdaAliasPolicy(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getLambdaAliasPolicy")

	alias := h.Item.(*aliasRowData)

	// Create Session
	svc, err := LambdaService(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("getLambdaAliasPolicy", "error_LambdaService", err)
		return nil, err
	}

	input := &lambda.GetPolicyInput{
		FunctionName: aws.String(*alias.FunctionName),
		Qualifier:    aws.String(*alias.Alias.Name),
	}

	op, err := svc.GetPolicy(input)
	if err != nil {
		plugin.Logger(ctx).Error("getLambdaAliasPolicy", "error_GetPolicy", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return lambda.GetPolicyOutput{}, nil
			}
		}
		return nil, err
	}

	return op, nil
}
