import ec2 = require("aws-cdk-lib/aws-ec2");
import ecs = require("aws-cdk-lib/aws-ecs");
import ecr = require("aws-cdk-lib/aws-ecr");
import servicediscovery = require("aws-cdk-lib/aws-servicediscovery");
import { Vpc } from "aws-cdk-lib/aws-ec2";
import { Construct } from "constructs";

/**
 * Properties to pass into Service class for creating a microservice.
 *
 * See Service for more details.
 */
export interface ServiceProps {
  vpc: Vpc;
  cluster: ecs.Cluster;
  namespace: servicediscovery.PrivateDnsNamespace;
  name: string;
  tag?: string;
  environment?: { [key: string]: string };
}

/**
 * Service creates an ECS task and service from an image stored in an
 * ECR repository.
 *
 * The Service class and its ServiceProps make a number of conventions.
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
export class Service extends Construct {
  /**
   * ECS Task Definition, once the service is successfully created.
   */
  public readonly taskDefinition: ecs.FargateTaskDefinition;

  /**
   * ECS Service Definition, once the service is successfully created.
   */
  public readonly service: ecs.FargateService;

  /**
   *
   * @param scope Construct
   * @param id ID of the new service
   * @param props properties to configure the new service
   */
  constructor(scope: Construct, id: string, props: ServiceProps) {
    super(scope, id);

    // Access the ECR repository at `bookshelf/${serviceName}`
    const repository = ecr.Repository.fromRepositoryName(
      this,
      "Repository",
      `bookshelf/${props.name}`
    );

    // Create a task definition and map port 8080
    const taskDef = new ecs.FargateTaskDefinition(this, "TaskDef", {});
    taskDef.addContainer("Container", {
      image: ecs.ContainerImage.fromEcrRepository(
        repository,
        props.tag || "latest"
      ),
      portMappings: [{ containerPort: 8080 }],
      environment: props.environment,
    });
    this.taskDefinition = taskDef;

    // Create the service
    const service = new ecs.FargateService(this, "Service", {
      cluster: props.cluster,
      taskDefinition: taskDef,
      desiredCount: 1,
      cloudMapOptions: {
        name: props.name,
        cloudMapNamespace: props.namespace,
      },
    });
    service.connections.allowFrom(
      ec2.Peer.ipv4(props.vpc.vpcCidrBlock),
      ec2.Port.tcp(8080)
    );
    this.service = service;
  }
}
