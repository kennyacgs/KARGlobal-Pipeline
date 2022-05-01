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
