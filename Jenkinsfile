pipeline {
  agent any
  stages {
    stage('build') {
      steps {
        sh '''go build .
ls -al'''
        archiveArtifacts(artifacts: './backend', caseSensitive: true, onlyIfSuccessful: true)
      }
    }

  }
}