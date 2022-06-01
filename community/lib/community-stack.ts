import { Stack, StackProps, CfnOutput } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigw from '@aws-cdk/aws-apigatewayv2-alpha';
import * as apigwintg from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { HttpLambdaAuthorizer, HttpLambdaResponseType } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';
import * as path from 'path';

export class CommunityStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // The code that defines your stack goes here

    const jwtAuthorizer = this.buildAndInstallGOLambda('communityAuthorizerFn', path.join(__dirname, '../jwtAuthorizer'), 'main');
    const authorizer = new HttpLambdaAuthorizer('communityAuthorizer', jwtAuthorizer, {
      responseTypes: [HttpLambdaResponseType.SIMPLE], // Define if returns simple and/or iam response
    });
    jwtAuthorizer.addToRolePolicy(new iam.PolicyStatement({
      actions: ['ssm:GetParameter'],
      resources: ['arn:aws:ssm:ca-central-1:788541814854:parameter/*'],
    }));
    const apiFunction = this.buildAndInstallGOLambda('communityApiFn', path.join(__dirname, '../api'), 'main');

    // defines an API Gateway REST API resource backed by our "hello" function.
    const httpApi = new apigw.HttpApi(this, 'CommunityHttpApi', {
      apiName: 'CommunityHttpApi',
      defaultAuthorizer: authorizer,
      defaultIntegration: new apigwintg.HttpLambdaIntegration(
        'communityApiLambdaIntg',
        apiFunction,
      ),
    });

    new CfnOutput(this, 'lambda-url', { value: httpApi.url! });
  }

  /**
   * buildAndInstallGOLambda build the code and create the lambda
   * @param id - CDK id for this lambda
   * @param lambdaPath - Location of the code
   * @param handler - name of the handler to call for this lambda
   */
  buildAndInstallGOLambda(id: string, lambdaPath: string, handler: string): lambda.Function {
    const environment = {
      CGO_ENABLED: '0',
      GOOS: 'linux',
      GOARCH: 'amd64',
    };
    return new lambda.Function(this, id, {
      code: lambda.Code.fromAsset(lambdaPath, {
        bundling: {
          image: lambda.Runtime.GO_1_X.bundlingImage,
          user: "root",
          environment,
          command: [
            'bash', '-c', [
              'GOOS=linux go build -o /asset-output/main',
            ].join(' && ')
          ]
        },
      }),
      handler,
      runtime: lambda.Runtime.GO_1_X,
    });
  }
}
