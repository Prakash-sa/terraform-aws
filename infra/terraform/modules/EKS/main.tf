module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "~> 21.0"  # Use the latest version
  name    = "my-eks-cluster"
  kubernetes_version = "1.33"
  endpoint_public_access = true
  
  upgrade_policy = {
    support_type = "STANDARD"
  }
  
  addons = {
    coredns = {
      most_recent = true
      resolve_conflicts_on_create  = "OVERWRITE"
      resolve_conflicts_on_update  = "OVERWRITE"
    }
    kube-proxy = {
      before_compute = true
      most_recent = true
    }
    vpc-cni = {
      before_compute = true
      most_recent = true
    }
  }

  vpc_id                   = module.vpc.vpc_id
  subnet_ids               = module.vpc.private_subnets
  control_plane_subnet_ids = module.vpc.intra_subnets
  enable_cluster_creator_admin_permissions = true

  eks_managed_node_groups = {
    my-node-group = {
      min_size     = 2
      max_size     = 2  # Keep the max size low for cost savings
      desired_size = 2

      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"  # Use SPOT instances for further savings
    }
  }

  tags = local.tags
}