CDK_DEPLOY_ACCOUNT?=$(AWS_ACCOUNT_ID)
CDK_DEPLOY_REGION?=$(AWS_REGION)

default: build

.PHONY: build
build:
	npm run build

.PHONY: deploy
deploy:
	CDK_DEPLOY_ACCOUNT=$(CDK_DEPLOY_ACCOUNT) CDK_DEPLOY_REGION=$(CDK_DEPLOY_REGION) npx cdk deploy

.PHONY: bootstrap
bootstrap:
	CDK_NEW_BOOTSTRAP=1 npx cdk bootstrap \
		--cloudformation-execution-policies "arn:aws:iam::aws:policy/AdministratorAccess" \
		$(CDK_DEPLOY_ACCOUNT)/$(CDK_DEPLOY_REGION)
