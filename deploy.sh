#!/bin/bash

set -e

ROLE_NAME="lambda-basic-execution-role"
ZIP_FILE="main.zip"
LAMBDA_NAME="steller-sponsored-account"

build() {
  echo "Building..."
  GOARCH=amd64 GOOS=linux go build -tags lambda.norpc -o bootstrap *.go
}

package() {
  echo "Zipping..."
  zip -FS $ZIP_FILE bootstrap
}

create_lambda_basic_execution_role() {
  echo "Checking if role $ROLE_NAME exists..."
  if ! aws iam get-role --role-name $ROLE_NAME > /dev/null 2>&1; then
    echo "Creating lambda basic execution role..."
    aws iam create-role \
      --role-name $ROLE_NAME \
      --assume-role-policy-document '{
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Principal": {
              "Service": "lambda.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
          }
        ]
      }'
    aws iam attach-role-policy \
      --role-name $ROLE_NAME \
      --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
    aws iam attach-role-policy \
      --role-name $ROLE_NAME \
      --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole
  else
    echo "Role $ROLE_NAME already exists. Skipping..."
  fi
}

create_lambda_function() {
  echo "Creating lambda function..."
  ROLE_ARN=$(aws iam get-role --role-name $ROLE_NAME --query 'Role.Arn' --output text)
  ENV_VARS=$(cat .env | jq -R -s -c 'split("\n") | map(select(length > 0)) | map(split("=")) | map({(.[0]): .[1]}) | add')
  aws lambda create-function --function-name $LAMBDA_NAME \
    --zip-file fileb://$ZIP_FILE --handler bootstrap \
    --runtime provided.al2 --role $ROLE_ARN \
    --environment '{"Variables": '$ENV_VARS'}' \
    --timeout 15
}

update_lambda_function() {
  echo "Updating lambda function..."
  ENV_VARS=$(cat .env | jq -R -s -c 'split("\n") | map(select(length > 0)) | map(split("=")) | map({(.[0]): .[1]}) | add')
  aws lambda update-function-code --function-name $LAMBDA_NAME \
    --zip-file fileb://$ZIP_FILE
  aws lambda update-function-configuration --function-name $LAMBDA_NAME \
    --environment '{"Variables": '$ENV_VARS'}'
}

create_lambda_function_url() {
  echo "Creating Lambda function URL..."
  aws lambda create-function-url-config --function-name $LAMBDA_NAME --auth-type NONE \
    --cors '{
      "AllowOrigins": ["*"],
      "AllowMethods": ["POST"],
      "AllowHeaders": ["Content-Type"]
    }'
  FUNCTION_URL=$(aws lambda get-function-url-config --function-name $LAMBDA_NAME --query 'FunctionUrl' --output text)
  aws lambda add-permission --function-name $LAMBDA_NAME --statement-id function-url-permission --action lambda:InvokeFunctionUrl --principal '*' --function-url-auth-type NONE
  echo "Lambda Function URL: $FUNCTION_URL"
}

clean() {
  echo "Cleaning up..."
  rm -f $ZIP_FILE
  rm -f bootstrap
}

deploy() {
  build
  package
  create_lambda_basic_execution_role
  if aws lambda get-function --function-name $LAMBDA_NAME > /dev/null 2>&1; then
    update_lambda_function
  else
    create_lambda_function
    create_lambda_function_url
  fi
}

deploy
clean