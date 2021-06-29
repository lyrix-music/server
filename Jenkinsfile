pipeline {
  agent any
  stages {
    stage('build') {
      steps {
        sh '''go build -ldflags="-X \'github.com/lyrix-music/server/meta.BuildTime=$(date +%s)\' -X \'github.com/lyrix-music/server/meta.BuildVersion=$(git describe --always)\' -s -w" .
ls -al'''
        sh '''mv server "$TARGET_DIR/jenkins/bin/."
sudo systemctl restart lyrix-backend'''
      }
    }

  }
  environment {
    TARGET_DIR = '/home/bots'
  }
}