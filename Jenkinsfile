pipeline {
  agent any
  stages {
    stage('build') {
      steps {
        sh '''go build .
ls -al'''
        archiveArtifacts(artifacts: 'backend', caseSensitive: true, onlyIfSuccessful: true)
      }
    }

    stage('deploy') {
      steps {
        copyArtifacts(projectName: 'server', target: '/home/bots/jenkins/bin/')
      }
    }

  }
}