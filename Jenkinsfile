pipeline {
  agent any
  stages {
    stage('build') {
      steps {
        sh '''go build .
ls -al'''
        sh '''mv backend "$TARGET_DIR/jenkins/bin/."
sudo sytemctl restart lyrix-backend'''
      }
    }

  }
  environment {
    TARGET_DIR = '/home/bots'
  }
}