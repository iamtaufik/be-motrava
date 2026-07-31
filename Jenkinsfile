pipeline {
    agent any

    environment {
        COMPOSE_PROJECT_NAME = "be-motrava"
    }

    options {
        disableConcurrentBuilds()
    }

    stages {

        stage('Checkout') {
            steps {
                git branch: 'development',
                    url: 'https://github.com/iamtaufik/be-motrava.git',
                    credentialsId: '9d223e5d-6447-4ec3-83a5-70a0d718e0af'
            }
        }

        stage('Build Docker Images') {
            steps {
                withCredentials([
                    file(credentialsId: 'BE_MOTRAVA_ENVIRONMENT', variable: 'ENV_FILE')
                ]) {
                    sh '''
                    cp $ENV_FILE ./.env

                    docker compose build iam-service
                    docker compose build core-service
                    docker compose build caddy
                    '''
                }
            }
        }

        stage('Remove old containers') {
            steps {
                sh '''
                docker compose down --remove-orphans || true
                '''
            }
        }

        stage('Deploy with Docker Compose') {
            steps {
                sh '''
                docker compose up -d --remove-orphans iam-service core-service caddy
                '''
            }
        }
    }
}
