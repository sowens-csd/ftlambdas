import { Stack, StackProps, CfnOutput } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigw from '@aws-cdk/aws-apigatewayv2-alpha';
import * as apigwintg from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import * as lambdago from '@aws-cdk/aws-lambda-go-alpha';
import { HttpLambdaAuthorizer, HttpLambdaResponseType } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';
import * as path from 'path';
import { spawnSync } from 'child_process';
import { print } from 'util';
import { HttpMethod } from 'aws-cdk-lib/aws-events';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';

export class CommunityStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // The code that defines your stack goes here
    console.error(`Starting 1`)

    const jwtAuthorizer = this.buildAndInstallGOLambda(this, 'communityAuthorizerFn', path.join(__dirname, '../jwtAuthorizer'), 'main');
    const authorizer = new HttpLambdaAuthorizer('communityAuthorizer', jwtAuthorizer, {
      responseTypes: [HttpLambdaResponseType.SIMPLE], // Define if returns simple and/or iam response
    });
    jwtAuthorizer.addToRolePolicy(new iam.PolicyStatement({
      actions: ['ssm:GetParameter'],
      resources: ['arn:aws:ssm:ca-central-1:788541814854:parameter/*'],
    }));
    // const apiFunction = this.buildAndInstallGOLambda(this, 'communityApiHandler', path.join(__dirname, '../api'), 'main');
    const createFunction = this.buildAndInstallGOLambda(this, 'folkCreateHandler', path.join(__dirname, '../folkCreate'), 'main');
    this.grantDBPrivileges(createFunction);

    // defines an API Gateway REST API resource backed by our "hello" function.
    const httpApi = new apigw.HttpApi(this, 'CommunityHttpApi', {
      apiName: 'CommunityHttpApi',
    });
    httpApi.addRoutes({
      path: '/folk',
      methods: [HttpMethod.POST],
      authorizer: authorizer,
      integration: new HttpLambdaIntegration(
        'CommunityFolkCreateHandlerLambdaIntg',
        createFunction,
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
  buildAndInstallGOLambda(scope: Construct, id: string, lambdaPath: string, handler: string): lambda.Function {
    const environment = {
      CGO_ENABLED: '0',
      GOOS: 'linux',
      GOARCH: 'amd64',
    };
    return new lambdago.GoFunction(scope, id, {
      entry: lambdaPath,
      architecture: lambda.Architecture.X86_64,
      runtime: lambda.Runtime.GO_1_X,
      environment: {
        'storyTable': 'dev-story',
        'LOG_LEVEL': 'debug'
      },
      bundling: {
        goBuildFlags: ['-ldflags "-s -w"'],
        environment: environment,
      },
    });
  }

  grantDBPrivileges(lambdaFunction: lambda.Function) {
    const dbPrivileges = new iam.PolicyStatement({
      actions: [
        'dynamodb:GetItem',
        'dynamodb:PutItem',
        'dynamodb:Query',
        'dynamodb:DescribeTable',
        'dynamodb:DeleteItem',
        'dynamodb:UpdateItem',
        'dynamodb:Scan',
      ],
      resources: ['arn:aws:dynamodb:ca-central-1:788541814854:table/dev-story'],
    });
    lambdaFunction.addToRolePolicy(dbPrivileges);
  }
}
