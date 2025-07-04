/**
 * Complete YAML configuration interface based on YAML_SPECIFICATION.md
 */
export interface YamlInfrastructureConfig {
  // Core Settings (Required)
  project: string;
  env: string;
  is_prod: boolean;
  region: string;
  state_bucket: string;
  state_file: string;
  
  // Optional ECR configuration
  ecr_account_id?: string;
  ecr_account_region?: string;
  
  // Workload Configuration
  workload?: {
    // Basic backend configuration
    backend_health_endpoint?: string;
    backend_external_docker_image?: string;
    backend_container_command?: string[];
    backend_image_port?: number;
    backend_remote_access?: boolean;
    
    // S3 bucket configuration
    bucket_postfix?: string;
    bucket_public?: boolean;
    
    // Environment configuration
    backend_env_variables?: Array<{
      name: string;
      value: string;
    }>;
    env_files_s3?: Array<{
      bucket: string;
      key: string;
    }>;
    
    // Monitoring and observability
    xray_enabled?: boolean;
    
    // Notifications
    setup_fcnsns?: boolean;
    slack_webhook?: string;
    
    // CI/CD configuration
    enable_github_oidc?: boolean;
    github_oidc_subjects?: string[];
    
    // Database admin tools
    install_pg_admin?: boolean;
    pg_admin_email?: string;
    
    // IAM policies
    policy?: Array<{
      actions: string[];
      resources: string[];
    }>;
    
    // EFS mounts
    efs?: Array<{
      name: string;
      mount_point: string;
    }>;
    
    // ALB configuration
    backend_alb_domain_name?: string;
  };
  
  // Domain Configuration
  domain?: {
    enabled: boolean;
    create_domain_zone?: boolean;
    domain_name?: string;
    api_domain_prefix?: string;
    add_domain_prefix?: boolean;
  };
  
  // Database Configuration
  postgres?: {
    enabled: boolean;
    dbname?: string;
    username?: string;
    public_access?: boolean;
    engine_version?: string;
  };
  
  // Authentication Configuration
  cognito?: {
    enabled: boolean;
    enable_web_client?: boolean;
    enable_dashboard_client?: boolean;
    dashboard_callback_urls?: string[];
    enable_user_pool_domain?: boolean;
    user_pool_domain_prefix?: string;
    backend_confirm_signup?: boolean;
    auto_verified_attributes?: string[];
  };
  
  // Email Service Configuration
  ses?: {
    enabled: boolean;
    domain_name?: string;
    test_emails?: string[];
  };
  
  // Message Queue Configuration
  sqs?: {
    enabled: boolean;
    name?: string;
  };
  
  // File Storage Configuration
  efs?: Array<{
    name: string;
    path: string;
  }>;
  
  // Additional S3 buckets
  buckets?: Array<{
    name: string;
    public?: boolean;
  }>;
  
  // Load Balancer Configuration
  alb?: {
    enabled: boolean;
  };
  
  // Scheduled Tasks
  scheduled_tasks?: Array<{
    name: string;
    schedule: string;
    docker_image?: string;
    container_command?: string;
    allow_public_access?: boolean;
  }>;
  
  // Event-driven Tasks
  event_processor_tasks?: Array<{
    name: string;
    rule_name: string;
    detail_types: string[];
    sources: string[];
    docker_image?: string;
    container_command?: string;
    allow_public_access?: boolean;
  }>;
  
  // GraphQL API Configuration
  pubsub_appsync?: {
    enabled: boolean;
    schema?: boolean;
    auth_lambda?: boolean;
    resolvers?: boolean;
  };
  
  // Additional Services
  services?: Array<{
    name: string;
    docker_image?: string;
    container_command?: string[];
    container_port?: number;
    host_port?: number;
    cpu?: number;
    memory?: number;
    desired_count?: number;
    remote_access?: boolean;
    xray_enabled?: boolean;
    essential?: boolean;
    env_vars?: Array<{
      name: string;
      value: string;
    }>;
    env_files_s3?: Array<{
      bucket: string;
      key: string;
    }>;
  }>;
}