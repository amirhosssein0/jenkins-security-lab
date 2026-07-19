pipeline {
  agent {
    kubernetes {
      yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: kaniko
      image: gcr.io/kaniko-project/executor:debug
      command: ["sleep"]
      args: ["99d"]
      volumeMounts:
        - name: docker-config
          mountPath: /kaniko/.docker
    - name: checkov
      image: bridgecrew/checkov:latest
      command: ["sleep"]
      args: ["99d"]
    - name: trivy
      image: aquasec/trivy:latest
      command: ["sleep"]
      args: ["99d"]
    - name: tools     
      image: alpine:3.22
      command: ["sleep"]
      args: ["99d"]

  volumes:
    - name: docker-config
      secret:
        secretName: regcred
        items:
          - key: .dockerconfigjson
            path: config.json
"""
    }
  }

  environment {
  REGISTRY   = "docker.io/asdfghjkl0"
  IMAGE_NAME = "jenkins-security-lab-app"
  IMAGE_TAG  = "${env.BUILD_NUMBER}"
  FULL_IMAGE = "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('IaC Scan - Checkov') {
      steps {
        container('checkov') {
          sh 'checkov -d k8s/ --compact || true'
        }
      }
    }

    stage('Install SBOM/Sign tools') {
      steps {
        container('tools') {
          sh '''
            apk add --no-cache curl ca-certificates

            SYFT_VERSION=1.48.0
            curl -sSfL -o syft.tar.gz "https://github.com/anchore/syft/releases/download/v${SYFT_VERSION}/syft_${SYFT_VERSION}_linux_amd64.tar.gz"
            tar -xzf syft.tar.gz -C /usr/local/bin syft
            rm syft.tar.gz
            chmod +x /usr/local/bin/syft

            curl -sSfL -o /usr/local/bin/cosign "https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64"
            chmod +x /usr/local/bin/cosign
          '''
        }
      }
    }

    stage('Build & Push - Kaniko') {
      steps {
        container('kaniko') {
          sh """
            /kaniko/executor \
              --context=`pwd`/app \
              --dockerfile=`pwd`/app/Dockerfile \
              --destination=${FULL_IMAGE} \
              --cache=true
          """
        }
      }
    }
  }
}