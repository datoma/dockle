pipeline {
  environment {
    DOCKERHUB_IMAGE_NAME = 'datoma/dockle'
    ARTIFACTORY_IMAGE_NAME = 'datoma.jfrog.io/docker-local/dockle'
    dockerHubImageLatest = ''
    dockerHubImagetag = ''
    artifactoryImageLatest = ''
    artifactoryImageTag = ''
    GIT_URL = 'git@github.com:datoma/dockle.git'
    GIT_BRANCH = 'master'
    GIT_CREDENTIALS = 'Github_ssh'
  }

  parameters{
    text(defaultValue: "latest", description: 'tag to build/push', name: 'DOCKER_IMAGE_TAG')
    booleanParam(defaultValue: false, description: 'deploy to dockerhub', name: 'PUSH_DOCKER')
    booleanParam(defaultValue: false, description: 'deploy to artifactory', name: 'PUSH_ARTIFACTORY')
  }

  agent any
  stages {
    stage("check params") {
      steps {
        script {
          params.each {
            if (DOCKER_IMAGE_TAG == null || DOCKER_IMAGE_TAG == "" || DOCKER_IMAGE_TAG == "latest")
              error "This pipeline stops here because no tag was set (var is ${DOCKER_IMAGE_TAG})"
            }
        }
      }
    }

    stage ("prepare") {
      steps {
        script {
            currentBuild.displayName = "#${BUILD_NUMBER} (DockerTag: ${DOCKER_IMAGE_TAG} - Branch: ${env.GIT_BRANCH})"
            sh 'printenv'
        }
      }
    }
    //stage ("Prompt for input") {
    //  steps {
    //    script {
    //      env.DOCKER_IMAGE_TAG = input message: 'Please enter the docker image tag', parameters: [string(defaultValue: '', description: '', name: 'DockerImageTag')]
    //      echo "docker tag: ${env.DOCKER_IMAGE_TAG}"
    //    }
    //  }
    //}

    stage('Clone') {
      steps {
        git(url: "${GIT_URL}", branch: "${GIT_BRANCH}", credentialsId: "${GIT_CREDENTIALS}")
      }
    }

    stage('Building the image') {
      steps {
        script {
          docker.withRegistry('https://registry.hub.docker.com', 'dockerhub') {
            dockerHubImageLatest = docker.build("${DOCKERHUB_IMAGE_NAME}:latest")
          }
        }
      }
    }
    
    stage('Tagging the image') {
      steps {
        script {
          docker.withRegistry('https://registry.hub.docker.com', 'dockerhub') {
            dockerHubImagetag = docker.build("${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}")
          }
        }
      }
    }

    stage('Test stages') {
      parallel {
        stage('Trivy Tag and latest') {
          steps {
              script {
                trivy_latest = sh(returnStdout: true, script: 'docker run --name trivy-client --rm -i -v /var/run/docker.sock:/var/run/docker.sock:ro datoma/trivy-server:latest trivy client --remote https://trivy.blackboards.de ${DOCKERHUB_IMAGE_NAME}:latest')
                trivy_tag = sh(returnStdout: true, script: 'docker run --name trivy-client --rm -i -v /var/run/docker.sock:/var/run/docker.sock:ro datoma/trivy-server:latest trivy client --remote https://trivy.blackboards.de --ignore-unfixed --severity CRITICAL,HIGH,MEDIUM ${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}')
              }
              echo "TRIVY latest: ${trivy_latest}"
              writeFile(file: 'trivy-latest.txt', text: "${trivy_latest}")
              echo "TRIVY Tag: ${trivy_tag}"
              writeFile(file: 'trivy-tag.txt', text: "${trivy_tag}")
          }
        }
        stage('dockle Tag') {
          steps {
            script {
                dockle_tag = sh(returnStdout: true, script: 'docker run --rm -v /var/run/docker.sock:/var/run/docker.sock datoma/dockle:latest ${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}')
              }
              echo "Dockle tag: ${dockle_tag}"
              writeFile(file: 'dockle_tag.txt', text: "${dockle_tag}")
          }
        }
        stage('hadolint Tag') {
          steps {
            script {
              try {
                sh 'docker run --rm -i hadolint/hadolint < Dockerfile | tee hadolint_tag.txt'
              } catch (err) {
                echo err.getMessage()
              }
            }
          }
        }

      }
    }

    stage('Deploy Images to Dockerhub') {
      when {
        expression { 
          params.PUSH_DOCKER == true
        }
      }
      parallel {
        stage('Deploy image with latest to Dockerhub') {
          steps {
            script {
              docker.withRegistry('https://registry.hub.docker.com', 'dockerhub') {
                dockerHubImageLatest.push()
              }
            }
          }
        }
        stage('Deploy image with tag to Dockerhub') {
          steps {
            script {
              docker.withRegistry('https://registry.hub.docker.com', 'dockerhub') {
                dockerHubImagetag.push()
              }
            }
          }
        }
      }
    }

    stage('Tagging Artifactory images') {
      when {
        expression { 
          params.PUSH_ARTIFACTORY == true
        }
      }
      steps {
        script {
          artifactoryImageLatest = docker.build("${ARTIFACTORY_IMAGE_NAME}:latest")
          artifactoryImageTag = docker.build("${ARTIFACTORY_IMAGE_NAME}:${DOCKER_IMAGE_TAG}")
        }
      }
    }

    stage('Push images to Artifactory') {
      when {
        expression { 
          params.PUSH_ARTIFACTORY == true
        }
      }
      parallel {
        stage('Push image with latest to Artifactory') {
          steps {
            script {
                docker.withRegistry('https://datoma.jfrog.io/artifactory', 'ArtifactoryDockerhub') {
                artifactoryImageLatest.push()
              }
            }
          }
        }
        stage('Push image with tag to Artifactory') {
          steps {
            script {
              docker.withRegistry('https://datoma.jfrog.io/artifactory', 'ArtifactoryDockerhub') {
                artifactoryImageTag.push()
              }
            }
          }
        }
      }
    }
  }

  post {
    always {
      archiveArtifacts artifacts: '*.txt', onlyIfSuccessful: true
      sh "docker rmi ${DOCKERHUB_IMAGE_NAME}:latest"
      sh "docker rmi ${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
    }
    success {
      script {
        if (params.PUSH_ARTIFACTORY == true) {
          sh "docker rmi ${ARTIFACTORY_IMAGE_NAME}:latest"
          sh "docker rmi ${ARTIFACTORY_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
        }
      }
    }
  }
}