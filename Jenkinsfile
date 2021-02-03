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
    DOCKERHUB_URL = 'https://registry.hub.docker.com'
    DOCKERHUB_CREDENTIALS = 'dockerhub'
    ARTIFACTORY_URL = 'https://datoma.jfrog.io/artifactory'
    ARTIFACTORY_CREDENTIALS = 'ArtifactoryDockerhub'
    ANCHORE_URL = 'https://anchore.blackboards.de'
    ANCHORE_CREDENTIALS = 'anchore'

    TRIVY_VERSION = 'datoma/trivy-server:0.15.0'
    DOCKLE_VERSION = 'datoma/dockle:0.3.1'
    HADOLINT_VERSION = 'datoma/hadolint:1.21.0'
  }

  parameters{
    string(defaultValue: "latest", description: 'tag to build/push', name: 'DOCKER_IMAGE_TAG')
    booleanParam(defaultValue: false, description: 'create release branch', name: 'CREATE_RELEASE_BRANCH')
    booleanParam(defaultValue: false, description: 'deploy to dockerhub', name: 'PUSH_DOCKER')
    booleanParam(defaultValue: false, description: 'deploy to artifactory', name: 'PUSH_ARTIFACTORY')
  }

  options {
    buildDiscarder(logRotator(numToKeepStr: '', artifactNumToKeepStr: '15'))
    disableConcurrentBuilds()
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

    stage('Clone') {
      steps {
        git(url: "${GIT_URL}", branch: "${GIT_BRANCH}", credentialsId: "${GIT_CREDENTIALS}")

      }
    }
    
    stage('create branch') {
      when {
        expression { 
          params.CREATE_RELEASE_BRANCH == true
        }
      }
      steps {
        withCredentials([sshUserPrivateKey(credentialsId: 'Github_ssh', keyFileVariable: '', passphraseVariable: '', usernameVariable: '')]) {
          //sh("git push origin ${GIT_BRANCH}:release-${DOCKER_IMAGE_TAG}")
          sh("git checkout -b releases/${DOCKER_IMAGE_TAG}")
          sh("git push --set-upstream origin releases/${DOCKER_IMAGE_TAG}")
        }
       }
    }

    stage('Building the image') {
      steps {
        script {
          docker.withRegistry("${DOCKERHUB_URL}", "${DOCKERHUB_CREDENTIALS}") {
            dockerHubImageLatest = docker.build("${DOCKERHUB_IMAGE_NAME}:latest")
          }
        }
      }
    }
    
    stage('Tagging the image') {
      steps {
        script {
          docker.withRegistry("${DOCKERHUB_URL}", "${DOCKERHUB_CREDENTIALS}") {
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
                trivy_latest = sh(returnStdout: true, script: 'docker run --name trivy-client --rm -i -v /var/run/docker.sock:/var/run/docker.sock:ro ${TRIVY_VERSION} trivy client --remote https://trivy.blackboards.de ${DOCKERHUB_IMAGE_NAME}:latest')
                trivy_tag = sh(returnStdout: true, script: 'docker run --name trivy-client --rm -i -v /var/run/docker.sock:/var/run/docker.sock:ro ${TRIVY_VERSION} trivy client --remote https://trivy.blackboards.de --ignore-unfixed --severity CRITICAL,HIGH,MEDIUM ${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}')
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
                dockle_tag = sh(returnStdout: true, script: 'docker run --rm -v /var/run/docker.sock:/var/run/docker.sock ${DOCKLE_VERSION} ${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}')
              }
              echo "Dockle tag: ${dockle_tag}"
              writeFile(file: 'dockle_tag.txt', text: "${dockle_tag}")
          }
        }
        stage('anchore') {
          steps {
            script {
                sh 'echo "${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" > anchore_images'
                anchore engineCredentialsId: "${ANCHORE_CREDENTIALS}", engineRetries: '600', engineurl: "${ANCHORE_URL}", name: 'anchore_images'
              }
          }
        }
        stage('hadolint Tag') {
          steps {
            script {
              try {
                sh 'docker run --rm -i ${HADOLINT_VERSION} < Dockerfile | tee hadolint_tag.txt'
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
              docker.withRegistry("${DOCKERHUB_URL}", "${DOCKERHUB_CREDENTIALS}") {
                dockerHubImageLatest.push()
              }
            }
          }
        }
        stage('Deploy image with tag to Dockerhub') {
          steps {
            script {
              docker.withRegistry("${DOCKERHUB_URL}", "${DOCKERHUB_CREDENTIALS}") {
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
                docker.withRegistry("${ARTIFACTORY_URL}", "${ARTIFACTORY_CREDENTIALS}") {
                artifactoryImageLatest.push()
              }
            }
          }
        }
        stage('Push image with tag to Artifactory') {
          steps {
            script {
              docker.withRegistry("${ARTIFACTORY_URL}", "${ARTIFACTORY_CREDENTIALS}") {
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
      //sh "docker rmi ${DOCKERHUB_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
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