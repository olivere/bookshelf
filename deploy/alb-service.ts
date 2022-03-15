import ecs = require("aws-cdk-lib/aws-ecs");
import ecr = require("aws-cdk-lib/aws-ecr");
import servicediscovery = require("aws-cdk-lib/aws-servicediscovery");
import ecs_patterns = require("aws-cdk-lib/aws-ecs-patterns");
import { Vpc } from "aws-cdk-lib/aws-ec2";
import { Construct } from "constructs";

/**
 * Properties to pass into ALBService class for creating a Fargate service
 * that is backed by an Application Load Balancer (ALB).
 *
 * See ALBService for more details.
 */
export interface ALBServiceProps {
  vpc: Vpc;
  cluster: ecs.Cluster;
  namespace: servicediscovery.PrivateDnsNamespace;
  name: string;
  tag?: string;
  environment?: { [key: string]: string };
}

/**
 * ALBService creates an ECS task and service from an image stored in an
 * ECR repository that is backed by an Application Load Balancer (ALB).
 *
 * The ALBService class and its ALBServiceProps make a number of conventions.
 *
 * 1. The ECR repository must be accessible by name `bookshelf/${serviceName}`.
 * 2. The container image to be used has tag `latest` by default.
 * 3. The desired count of the new service is 1.
 * 4. All services use Fargate.
 * 5. All services are accessible at port 8080 by all other services in the
 *    cluster.
 * 6. All services in the ECS cluster are accessible by its name via a
 *    private DNS service registered under the ".local" domain. E.g. a service
 *    by the name "identity" is accessible at "http://identity.local:8080".
 */
export class ALBService extends Construct {
  /**
   * ECS Service Definition, once the service is successfully created.
   */
  public readonly service: ecs_patterns.ApplicationLoadBalancedFargateService;

  /**
   *
   * @param scope Construct
   * @param id ID of the new service
   * @param props properties to configure the new service
   */
  constructor(scope: Construct, id: string, props: ALBServiceProps) {
    super(scope, id);

    // Access the ECR repository at `bookshelf/${serviceName}`
    const repository = ecr.Repository.fromRepositoryName(
      this,
      "Repository",
      `bookshelf/${props.name}`
    );

    // Create the service
    const service = new ecs_patterns.ApplicationLoadBalancedFargateService(
      this,
      "ALBService",
      {
        cluster: props.cluster,
        taskImageOptions: {
          image: ecs.ContainerImage.fromEcrRepository(
            repository,
            props.tag || "latest"
          ),
          containerPort: 8080,
          environment: props.environment,
        },
        cloudMapOptions: {
          name: props.name,
          cloudMapNamespace: props.namespace,
        },
      }
    );
    this.service = service;
  }
}
