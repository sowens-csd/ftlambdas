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
import * as s3 from 'aws-cdk-lib/aws-s3';
import { ArnPrincipal } from 'aws-cdk-lib/aws-iam';

export class CommunityStack extends Stack {
  constructor(scope: Construct, id: string, folktellsMediaBucket: s3.Bucket, props?: StackProps) {
    super(scope, id, props);

    // The code that defines your stack goes here
    // console.error(`Starting 1`)

    const jwtAuthorizer = this.buildAndInstallGOLambda(this, 'communityAuthorizerFn', path.join(__dirname, '../jwtAuthorizer'), 'main');
    const authorizer = new HttpLambdaAuthorizer('communityAuthorizer', jwtAuthorizer, {
      responseTypes: [HttpLambdaResponseType.SIMPLE], // Define if returns simple and/or iam response
    });
    this.grantSSMPrivileges(jwtAuthorizer);

    const orgFunction = this.buildAndInstallGOLambda(this, 'orgHandler', path.join(__dirname, '../org'), 'main');
    this.grantDBPrivileges(orgFunction);

    const folkFunction = this.buildAndInstallGOLambda(this, 'folkHandler', path.join(__dirname, '../folk'), 'main');
    this.grantDBPrivileges(folkFunction);

    const signupFunction = this.buildAndInstallGOLambda(this, 'signupHandler', path.join(__dirname, '../signup'), 'main');
    this.grantDBPrivileges(signupFunction);
    this.grantEmailPrivileges(signupFunction);

    const signupVerifyFunction = this.buildAndInstallGOLambda(this, 'signupVerifyHandler', path.join(__dirname, '../signup_verify'), 'main');
    this.grantDBPrivileges(signupVerifyFunction);

    const passwordlessAuthorizerFunction = this.buildAndInstallGOLambda(this, 'passwordlessAuthorizer', path.join(__dirname, '../passwordlessAuthorizer'), 'main');
    this.grantDBPrivileges(passwordlessAuthorizerFunction);
    this.grantSSMPrivileges(passwordlessAuthorizerFunction);

    const mediaAccessFunction = this.buildAndInstallGOLambda(this, 'mediaAccess', path.join(__dirname, '../mediaAccess'), 'main');
    this.grantSSMPrivileges(mediaAccessFunction);
    mediaAccessFunction.addEnvironment('s3Bucket', folktellsMediaBucket.bucketName);
    folktellsMediaBucket.grantReadWrite(mediaAccessFunction);

    // defines an API Gateway REST API resource 
    const httpApi = new apigw.HttpApi(this, 'CommunityHttpApi', {
      apiName: 'CommunityHttpApi',
    });
    httpApi.addRoutes({
      path: '/mgr/org',
      methods: [HttpMethod.GET],
      authorizer: authorizer,
      integration: new HttpLambdaIntegration(
        'CommunityOrgHandlerLambdaIntg',
        orgFunction,
      ),
    });
    httpApi.addRoutes({
      path: '/mgr/org/{orgID}/{subtype}',
      methods: [HttpMethod.GET],
      authorizer: authorizer,
      integration: new HttpLambdaIntegration(
        'CommunityOrgHandlerLambdaIntg',
        orgFunction,
      ),
    });
    httpApi.addRoutes({
      path: '/mgr/folk',
      methods: [HttpMethod.POST],
      authorizer: authorizer,
      integration: new HttpLambdaIntegration(
        'CommunityFolkHandlerLambdaIntg',
        folkFunction,
      ),
    });
    httpApi.addRoutes({
      path: '/mgr/media/{mediaFile}/{contentType}',
      methods: [HttpMethod.GET],
      authorizer: authorizer,
      integration: new HttpLambdaIntegration(
        'CommunityMediaAccessHandlerLambdaIntg',
        mediaAccessFunction,
      ),
    });
    httpApi.addRoutes({
      path: '/mgr/auth/signup',
      methods: [HttpMethod.POST],
      integration: new HttpLambdaIntegration(
        'CommunitySignupHandlerLambdaIntg',
        signupFunction,
      ),
    });
    httpApi.addRoutes({
      path: '/mgr/auth/signup/{requestID}',
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration(
        'CommunitySignupVerifyHandlerLambdaIntg',
        signupVerifyFunction,
      ),
    });
    httpApi.addRoutes({
      path: '/mgr/auth/token',
      methods: [HttpMethod.GET],
      integration: new HttpLambdaIntegration(
        'CommunityPasswordlessAuthorizerHandlerLambdaIntg',
        passwordlessAuthorizerFunction,
      ),
    });

    // const plan = httpApi.addUsagePlan('DevPlan', {
    //   name: 'Easy',
    //   throttle: {
    //     rateLimit: 10,
    //     burstLimit: 2
    //   }
    // });
    // const key = httpApi.addApiKey('ApiKey');
    // plan.addApiKey(key);
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
        'LOG_LEVEL': 'debug',
        'gatewayResources': 'arn:aws:execute-api:ca-central-1:788541814854:*'
      },
      bundling: {
        goBuildFlags: ['-ldflags "-s -w"'],
        environment: environment,
      },
    });
  }

  grantEmailPrivileges(lambdaFunction: lambda.Function) {
    const emailPrivileges = new iam.PolicyStatement({
      actions: [
        'ses:SendEmail',],
      resources: ['arn:aws:ses:ca-central-1:788541814854:identity/*'],
    });
    lambdaFunction.addToRolePolicy(emailPrivileges);
  }

  grantSSMPrivileges(lambdaFunction: lambda.Function) {
    const ssmPrivileges = new iam.PolicyStatement({
      actions: ['ssm:GetParameter'],
      resources: ['arn:aws:ssm:ca-central-1:788541814854:parameter/*'],
    });
    lambdaFunction.addToRolePolicy(ssmPrivileges);
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
      resources: ['arn:aws:dynamodb:ca-central-1:788541814854:table/dev-story',
        'arn:aws:dynamodb:ca-central-1:788541814854:table/dev-story/index/*'],
    });
    lambdaFunction.addToRolePolicy(dbPrivileges);
  }
}
