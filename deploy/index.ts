import ec2 = require("aws-cdk-lib/aws-ec2");
import ecs = require("aws-cdk-lib/aws-ecs");
import servicediscovery = require("aws-cdk-lib/aws-servicediscovery");
import cdk = require("aws-cdk-lib");
import { ALBService } from "./alb-service";
import { Service } from "./service";

/**
 * BookshelfStackProps represents global settings with regards to the
 * Bookshelf stack.
 */
interface BookshelfStackProps extends cdk.StackProps {
  maxAzs: number | undefined;
}

/**
 * Sets up the Bookshelf application on Fargate.
 */
class BookshelfStack extends cdk.Stack {
  public readonly backend: Service;
  public readonly frontend: ALBService;

  /**
   *
   * @param scope CDK app
   * @param id ID of the stack
   * @param props properties to configure the stack.
   */
  constructor(scope: cdk.App, id: string, props?: BookshelfStackProps) {
    super(scope, id, props);

    // Create VPC, Fargate Cluster, and private DNS namespace
    const vpc = new ec2.Vpc(this, "BookshelfVpc", {
      maxAzs: props?.maxAzs || 2,
    });
    const cluster = new ecs.Cluster(this, "BookshelfCluster", { vpc });
    const namespace = new servicediscovery.PrivateDnsNamespace(
      this,
      "BookshelfNS",
      {
        name: "local",
        vpc,
      }
    );

    // TODO Create other resources like certificates, secrets etc.

    // Create Backend service
    const backend = new Service(this, "Backend", {
      vpc,
      cluster,
      namespace,
      name: "backend",
      environment: {
        PORT: "8080",
      },
    });
    this.backend = backend;

    // Create Frontend service
    const frontend = new ALBService(this, "Frontend", {
      vpc,
      cluster,
      namespace,
      name: "frontend",
      environment: {
        PORT: "8080",
        BACKEND_URL: "http://backend.local:8080",
      },
    });
    this.frontend = frontend;
  }
}

// Create a new app and deploy the stack.
const app = new cdk.App();

new BookshelfStack(app, "Bookshelf", {
  maxAzs: 2,
});

// Creates the CloudFormation template in cdk.out
app.synth();
