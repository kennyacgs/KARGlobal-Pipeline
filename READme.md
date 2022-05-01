#  **<span>Take Home Exercise KARGlobal</span>**

## Overview and tools.

+ Cloud platform: linode
   + NodeBalancer: cloud native load balancer
   + LKE: Linode Kubernetes Engine
+ Docker: 
+ Jenkins 
+ Terraform
+ Web Application: Go
+ GitHub


## Step1. 'Develop' go application
#### Since "there is no restriction to the resources/documentation you use...", I sourced the Go application base code from a blog post, researched and tweaked it for my needs. Ref:

``` sh
https://golangbyexample.com/all-permutations-string-golang/
```

## step 2.  Build image using docker
#### Did a multi stage build to reduce image size with alpine as my final image.
```sh
FROM golang:alpine3.15 as first
WORKDIR /app
COPY ./go/go.mod ./  
COPY ./go/go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /demo
EXPOSE 8082
CMD [ "/demonstration" ]

FROM alpine:3.15.4

WORKDIR /
RUN addgroup -S demogroup && adduser -S demouser -G demogroup
COPY --from=first /demo /demo
EXPOSE 8082
USER demouser
ENTRYPOINT ["/demo"]
```


## Step 3. Install Docker & run Jenkins as a container for my CI/CD tool. 
#### To have Jenkins run docker natively within the container, mounted volume inside jenkins container to point to the docker.sock on host machine.

```sh
docker run -p 8180:8080 -p 50000:50000 -d -v /volume1/docker/jenkins:/var/jenkins_home -v /var/run/docker.sock:/var/run/docker.sock -v $(which docker):/usr/bin/docker jenkins/jenkins:lts
```


## Step 4. Install terraform inside Jenkins container in order to have Jenkins run them locally 
#### Note: (530e07edeff5) is my container ID fro Jenkins.
```sh
docker exec -u 0 -it 530e07edeff5 bash #for root access

apt-get update && sudo apt-get install -y gnupg software-properties-common curl

curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -

apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"

sudo apt-get update && sudo apt-get install terraform
```


## Step 5. Install kubectl inside Jenkins contaianer in order to have Jenkins run them locally 
```sh
curl -LO https://dl.k8s.io/release/v1.23.0/bin/linux/amd64/kubectl
curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```


## Step 6. Install AWS Cli inside Jenkins container in order to have Jenkins connect to Linode Object storage
#### Linode object storage is s3 compatible so here, we trick the cli to think it's connecting to AWS by configuring the cli with linode key and secret ID just like we would for AWS. Sneaky but it works 😉
```sh
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
 ./aws/install
 
 aws configure
```


## Step 7. Create access token on linode for terraform. 


## Step 8. Write terraform configuration files to provision infrastructure on Linode with LKE
#### Found inside ./linode/ (main.tf, variables.tf & terraform.tfvars )
```sh
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
```

## Step 9. Write Kubernetes manifest to pull the built go image and deploy.
#### Service is created as a load balancer which thus limiting to single entry access into cluster. Files @: ./k8s_deployment/ (demo-app.yaml, namespace.yaml)
```sh
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
  namespace: kar-demo-ns
spec:
  selector:
    matchLabels:
      app: demo-app
  template:
    metadata:
      labels:
        app: demo-app
    spec:
      containers:
      - name: server
        image: panny0109/kar-demo:1.1
        ports:
        - containerPort: 8082
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 128Mi
---
apiVersion: v1
kind: Service
metadata:
  name: demo-app-service
  namespace: kar-demo-ns
spec:
  type: LoadBalancer
  selector:
    app: demo-app
  ports:
    - name: http
      port: 2323
      targetPort: 8082
```

## Step 10. Create pipeline jobs in Jenkins to build application, provision infrastructure and finally destroy. 
#### I am running Jenkins from my local environment and because it does not have a public IP, I am unable to setup webhook for GitHub on commits. At this time everything is run by manual builds however, I could setup a reverse proxy to provision something if needed. 
```sh
pipeline {
	agent any

	environment { 
		DOCKERHUB_CREDENTIALS= credentials('DockerHub_Token') 
	}        
	stages{
		stage('Git checkout'){
			steps {
				echo 'Cloning SCM repo'
				git credentialsId: 'GitHub_Token', url: 'https://github.com/panny0109/KARGlobal-Demo.git'	    
			}
		}
		stage('DockerHub Login') {         
      			steps{  
				sh "echo $DOCKERHUB_CREDENTIALS_PSW | docker login -u $DOCKERHUB_CREDENTIALS_USR --password-stdin"
				echo 'Login Completed'                
      			}           
   		 }      
		stage('Docker Build and push'){
      			steps{
				sh 'docker build -t panny0109/kar-demo:1.1 .'
				sh 'docker push panny0109/kar-demo:1.1'
			}
    		}	
		
		stage('Terraform init & plan'){
			steps{
				dir('linode'){
					sh 'terraform init -reconfigure'
					sh 'terraform plan'
				}
			}
		}		
		stage('TF-Plan Approval'){
			steps{
				script{
					echo 'User approval required'
					def userInput = input message: 'Approve to apply changes', ok: 'Yes', submitter: 'admin'
				}
			}
		}  
		stage('TF Apply') {
			steps {
				script{
					dir('linode'){
						sh "terraform apply --auto-approve"
						KUBE_VAR = sh( 
						    script: 'terraform output kubeconfig',
						    returnStdout: true 
						)
						def CLUSTER_URL = sh( 
						    script: 'terraform output api_endpoints',
						    returnStdout: true
						).trim()
						sh "export KUBECONFIG=$KUBE_VAR"
						//sh "export CLUSTER_URL=$CLUSTER_VAR"
						sh "echo $CLUSTER_URL"	
					}
				}
     			}
		}

	} //stages
	post{
              always {  
                sh 'docker logout'           
              }      
        }  
} //pipeline
```


## Step 11. Upload all files to GitHub for your perusal 😁 
