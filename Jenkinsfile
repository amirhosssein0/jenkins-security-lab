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

  stages {
    stage('placeholder') {
      steps { echo 'agent works' }
    }

    stage('Install SBOM/Sign tools') {
      steps {
        container('tools') {
          sh '''
            apk add --no-cache curl ca-certificates
            curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
            curl -O -L https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64
            chmod +x cosign-linux-amd64 && mv cosign-linux-amd64 /usr/local/bin/cosign
          '''
        }
      }
    }
  }
}