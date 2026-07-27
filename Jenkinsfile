pipeline {
  agent {
    kubernetes {
      yaml """
apiVersion: v1
kind: Pod
spec:
  serviceAccountName: jenkins-agent
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
      volumeMounts:
      - name: docker-config
        mountPath: /home/jenkins/.docker

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

            HELM_VERSION=3.21.3
            curl -sSfL -o helm.tar.gz "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz"
            tar -xzf helm.tar.gz --strip-components=1 -C /usr/local/bin linux-amd64/helm
            rm helm.tar.gz
            chmod +x /usr/local/bin/helm
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
              --tar-path=`pwd`/image.tar \
              --ignore-path=/product_uuid \
              --digest-file=digest.txt
          """
        }
      }
    }

    stage('SBOM - Syft') {
      steps {
        container('tools') {
          sh "syft docker-archive:image.tar -o cyclonedx-json=sbom.json"
        }
      }
      post {
        always { archiveArtifacts artifacts: 'sbom.json', fingerprint: true }
      }
    }

    stage('Vuln Scan - Trivy') {
      steps {
        container('trivy') {
          sh "trivy image --input image.tar --severity CRITICAL,HIGH --exit-code 1"
        }
      }
    }

    stage('Sign - cosign') {
      steps {
        container('tools') {
          withCredentials([
            file(credentialsId: 'cosign-private-key', variable: 'COSIGN_KEY'),
            string(credentialsId: 'cosign-password', variable: 'COSIGN_PASSWORD')
          ]) {
            sh """
              export DOCKER_CONFIG=/home/jenkins/.docker
              cosign sign --key \$COSIGN_KEY --tlog-upload=false --use-signing-config=false --new-bundle-format=false -y ${FULL_IMAGE}
            """
          }
        }
      }
    }

    stage('Attest SBOM - cosign') {
      steps {
        container('tools') {
          withCredentials([
            file(credentialsId: 'cosign-private-key', variable: 'COSIGN_KEY'),
            string(credentialsId: 'cosign-password', variable: 'COSIGN_PASSWORD')
          ]) {
            sh """
              export DOCKER_CONFIG=/home/jenkins/.docker
              cosign attest --key \$COSIGN_KEY --tlog-upload=false --use-signing-config=false --new-bundle-format=false --predicate sbom.json --type cyclonedx -y ${FULL_IMAGE}
            """
          }
        }
      }
    }

    stage('Deploy - Helm') {
      steps {
        container('tools') {
          sh """
            DIGEST=\$(cat digest.txt)
            helm upgrade --install app ./k8s/app-chart \
              --namespace app \
              --set image.repository=${REGISTRY}/${IMAGE_NAME} \
              --set image.digest=\${DIGEST}
          """
        }
      }
    }
  }
}