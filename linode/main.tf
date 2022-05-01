terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "1.25.0" 
    }
  }
}

provider "linode" {
  token = var.token
}

terraform {
  backend "s3" {
    profile = "linode-s3"
    bucket = "ivan-storage-672200"
    key    =  "tf_state_backend"
    region = "us-east-1"
    endpoint = "us-east-1.linodeobjects.com"
    skip_credentials_validation = true                #This will help stop TF from checking with amazon if creds are valid amazon s3 creds. 
  }
}

resource "linode_lke_cluster" "demo-cluster" {
    k8s_version = var.k8s_version
    label = var.label
    region = var.region
    tags = var.tags

    dynamic "pool" {
        for_each = var.pools
        content {
            type  = pool.value["type"]
            count = pool.value["count"]
        }
    }
}

output "kubeconfig" {
   value = linode_lke_cluster.demo-cluster.kubeconfig
   sensitive = true
}

output "api_endpoints" {
   value = linode_lke_cluster.demo-cluster.api_endpoints[0]
}

output "status" {
   value = linode_lke_cluster.demo-cluster.status
}

output "id" {
   value = linode_lke_cluster.demo-cluster.id
}

output "pool" {
   value = linode_lke_cluster.demo-cluster.pool
} 
