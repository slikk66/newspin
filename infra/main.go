package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecr"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		conf := config.New(ctx, "")
		envName := conf.Require("env")
		stateBucket := conf.Require("stateBucket")
		repoName := conf.Require("repoName")

		/// set infra prefix
		infraPrefix := "newspin-"
		infraEnvPrefix := "newspin-" + envName + "-"

		/// ECR for new image
		ecrName := infraPrefix + "newspin"

		ecrRepo, err := ecr.NewRepository(ctx, ecrName, &ecr.RepositoryArgs{
			Name:               pulumi.String(ecrName),
			ImageTagMutability: pulumi.String("MUTABLE"),
			ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
				ScanOnPush: pulumi.Bool(false),
			},
		})
		if err != nil {
			return err
		}

		/// user dynamo table
		userDynamoTableName := infraEnvPrefix + "users"

		usersTable, err := dynamodb.NewTable(ctx, userDynamoTableName, &dynamodb.TableArgs{
			Name:        pulumi.String(userDynamoTableName),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("username"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("username"),
					Type: pulumi.String("S"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(userDynamoTableName),
				"Environment": pulumi.String(envName),
			},
		})
		if err != nil {
			return err
		}

		/// pins dynamo table
		pinsDynamoTableName := infraEnvPrefix + "pins"

		pinsTable, err := dynamodb.NewTable(ctx, pinsDynamoTableName, &dynamodb.TableArgs{
			Name:        pulumi.String(pinsDynamoTableName),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("userId"),
			RangeKey:    pulumi.String("articleId"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("userId"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("articleId"),
					Type: pulumi.String("S"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(pinsDynamoTableName),
				"Environment": pulumi.String(envName),
			},
		})
		if err != nil {
			return err
		}

		/// iam role and associated for pods
		iamRoleName := infraEnvPrefix + "primary"
		iamPolicyName := infraEnvPrefix + "primary"

		// build the policy doc
		policyJson := pulumi.All(usersTable.Arn, pinsTable.Arn).ApplyT(
			func(args []any) (string, error) {
				usersArn := args[0].(string)
				pinsArn := args[1].(string)
				return fmt.Sprintf(`{
		              "Version": "2012-10-17",
		              "Statement": [
		                  {
		                      "Effect": "Allow",
		                      "Action": [
		                          "dynamodb:GetItem",
		                          "dynamodb:PutItem",
		                          "dynamodb:DeleteItem",
		                          "dynamodb:Query"
		                      ],
		                      "Resource": ["%s", "%s"]
		                  }
		              ]
		          }`, usersArn, pinsArn), nil
			},
		).(pulumi.StringOutput)
		if err != nil {
			return err
		}

		/// create the policy
		iamPolicy, err := iam.NewPolicy(ctx, iamPolicyName, &iam.PolicyArgs{
			Name:   pulumi.String(iamPolicyName),
			Policy: policyJson,
			Path:   pulumi.String("/"),
		})
		if err != nil {
			return err
		}

		/// create the assume role policy, locked to eks pods
		assumePolicyJson, err := json.Marshal(map[string]any{
			"Version": "2012-10-17",
			"Statement": []map[string]any{
				{
					"Action": []string{"sts:AssumeRole", "sts:TagSession"},
					"Effect": "Allow",
					"Sid":    "",
					"Principal": map[string]any{
						"Service": "pods.eks.amazonaws.com",
					},
				},
			},
		})
		if err != nil {
			return err
		}

		/// create the role
		iamRole, err := iam.NewRole(ctx, iamRoleName, &iam.RoleArgs{
			Name:             pulumi.String(iamRoleName),
			AssumeRolePolicy: pulumi.String(string(assumePolicyJson)),
			Tags: pulumi.StringMap{
				"Environment": pulumi.String(envName),
			},
		})
		if err != nil {
			return err
		}

		/// attach the policy
		_, err = iam.NewRolePolicyAttachment(ctx, iamPolicyName, &iam.RolePolicyAttachmentArgs{
			PolicyArn: iamPolicy.Arn,
			Role:      iamRole.Name,
		})
		if err != nil {
			return err
		}

		/// OIDC provider for github actions
		oidcProvider, err := iam.NewOpenIdConnectProvider(ctx, "github-oidc", &iam.OpenIdConnectProviderArgs{
			Url: pulumi.String("https://token.actions.githubusercontent.com"),
			ClientIdLists: pulumi.StringArray{
				pulumi.String("sts.amazonaws.com"),
			},
			ThumbprintLists: pulumi.StringArray{
				pulumi.String("ffffffffffffffffffffffffffffffffffffffff"),
			},
		})
		if err != nil {
			return err
		}

		/// iam role and associated
		iamOidcRoleName := infraEnvPrefix + "oidc"
		iamOidcPolicyName := infraEnvPrefix + "oidc"

		/// create the assume role policy, locked to eks pods
		oidcPolicyJson, err := json.Marshal(map[string]any{
			"Version": "2012-10-17",
			"Statement": []map[string]any{
				{
					"Effect":   "Allow",
					"Action":   []string{"s3:GetObject", "s3:PutObject"},
					"Resource": "arn:aws:s3:::" + stateBucket + "/*",
				},
				{
					"Effect": "Allow",
					"Action": []string{
						"ecr:GetAuthorizationToken",
						"ecr:BatchCheckLayerAvailability",
						"ecr:PutImage",
						"ecr:InitiateLayerUpload",
						"ecr:UploadLayerPart",
						"ecr:CompleteLayerUpload",
					},
					"Resource": "*",
				},
				{
					"Effect": "Allow",
					"Action": []string{
						"dynamodb:*",
						"iam:*",
						"ecr:*",
					},
					"Resource": "*",
				},
			},
		})
		if err != nil {
			return err
		}

		/// create the policy for the role
		iamOidcPolicy, err := iam.NewPolicy(ctx, iamOidcPolicyName, &iam.PolicyArgs{
			Name:   pulumi.String(iamOidcPolicyName),
			Policy: pulumi.String(string(oidcPolicyJson)),
			Path:   pulumi.String("/"),
		})
		if err != nil {
			return err
		}

		// build the assume policy doc
		assumePolicyJsonOidc := oidcProvider.Arn.ApplyT(func(arn string) (string, error) {
			return fmt.Sprintf(`{
          "Version": "2012-10-17",
          "Statement": [{
              "Effect": "Allow",
              "Action": "sts:AssumeRoleWithWebIdentity",
              "Principal": {"Federated": "%s"},
              "Condition": {
                  "StringEquals": {
                      "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
                  },
                  "StringLike": {
                      "token.actions.githubusercontent.com:sub": "repo:%s:*"
                  }
              }
          }]
      }`, arn, repoName), nil
		}).(pulumi.StringOutput)
		if err != nil {
			return err
		}

		/// create the oidc role, pass in the assume policy
		iamOidcRole, err := iam.NewRole(ctx, iamOidcRoleName, &iam.RoleArgs{
			Name:             pulumi.String(iamOidcRoleName),
			AssumeRolePolicy: assumePolicyJsonOidc,
			Tags: pulumi.StringMap{
				"Environment": pulumi.String(envName),
			},
		})
		if err != nil {
			return err
		}

		/// attach the policy
		_, err = iam.NewRolePolicyAttachment(ctx, iamOidcPolicyName, &iam.RolePolicyAttachmentArgs{
			PolicyArn: iamOidcPolicy.Arn,
			Role:      iamOidcRole.Name,
		})
		if err != nil {
			return err
		}

		/// outputs
		ctx.Export("usersTable", pulumi.Map{
			"name": usersTable.Name,
			"arn":  usersTable.Arn,
			"id":   usersTable.ID(),
		})
		ctx.Export("pinsTable", pulumi.Map{
			"name": pinsTable.Name,
			"arn":  pinsTable.Arn,
			"id":   pinsTable.ID(),
		})
		ctx.Export("ecrRepoUrl", ecrRepo.RepositoryUrl)
		ctx.Export("iamRoleArn", iamRole.Arn)
		ctx.Export("oidcProviderArn", oidcProvider.Arn)
		ctx.Export("oidcIamRoleArn", iamOidcRole.Arn)

		return nil
	})
}
