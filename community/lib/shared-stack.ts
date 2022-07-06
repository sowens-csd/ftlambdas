import { Stack, StackProps, CfnOutput, Duration, RemovalPolicy } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as s3 from 'aws-cdk-lib/aws-s3';
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

export class SharedStack extends Stack {
    readonly folktellsMediaBucket: s3.Bucket;

    constructor(scope: Construct, id: string, props?: StackProps) {
        super(scope, id, props);

        // The code that defines your stack goes here
        // console.error(`Starting 1`)

        // 👇 create bucket
        this.folktellsMediaBucket = new s3.Bucket(this, 'folktells-media-bucket', {
            // bucketName: 'my-bucket',
            removalPolicy: RemovalPolicy.DESTROY,
            autoDeleteObjects: true,
            versioned: false,
            publicReadAccess: false,
            encryption: s3.BucketEncryption.S3_MANAGED,
            lifecycleRules: [
                {
                    abortIncompleteMultipartUploadAfter: Duration.days(2),
                    expiration: Duration.days(365),
                    transitions: [
                        {
                            storageClass: s3.StorageClass.INFREQUENT_ACCESS,
                            transitionAfter: Duration.days(30),
                        },
                    ],
                },
            ],
        });

        // 👇 Create User
        const mediaBucketUser = new iam.User(this, 'folktellsMediaBucket-user', {
            userName: 'folktellsMediaBucket-user',
        });
        const accessKey = new iam.AccessKey(this, 'AccessKey', { user: mediaBucketUser });
        this.folktellsMediaBucket.addToResourcePolicy(
            new iam.PolicyStatement({
                effect: iam.Effect.ALLOW,
                principals: [mediaBucketUser],
                actions: ['s3:PutObject', 's3:GetObject', 's3:DeleteObject', 's3:DeleteObjectTagging', 's3:PutObjectTagging', 's3:ListBucket'],
                resources: [`${this.folktellsMediaBucket.bucketArn}/*`],
            }),
        );

        // 👇 grant access to bucket
        this.folktellsMediaBucket.grantReadWrite(new iam.AccountRootPrincipal());
        new CfnOutput(this, 'media bucket name', { value: this.folktellsMediaBucket.bucketName });
        new CfnOutput(this, 'media bucket arn', { value: this.folktellsMediaBucket.bucketArn });
        const accessKeyVal = accessKey.secretAccessKey.unsafeUnwrap();
        new CfnOutput(this, 'user access key', { value: accessKey.accessKeyId });
        new CfnOutput(this, 'user access key Secret', { value: accessKeyVal });
    }
}
