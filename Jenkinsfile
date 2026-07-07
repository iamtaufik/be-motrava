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

        stage('Build Docker Image') {
            steps {
                withCredentials([
                    file(credentialsId: 'BE_MOTRAVA_ENVIRONMENT', variable: 'ENV_FILE')
                ]) {
                    sh '''
                    cp $ENV_FILE ./.env

                    docker compose build be-motrava
                    '''
                }
            }
        }

        stage('Remove old containers') {
            steps {
                sh '''
                docker compose stop be-motrava || true
                docker compose rm -f be-motrava || true
                '''
            }
        }

        stage('Deploy with Docker Compose') {
            steps {
                sh '''
                docker compose up -d be-motrava
                '''
            }
        }

        // stage('Cleanup') {
        //     steps {
        //         sh '''
        //         docker images "hk-backend" \
        //         --format "{{.Tag}}" \
        //         | tail -n +2 \
        //         | xargs -r -I {} docker rmi hk-backend:{} || true
        //         '''
        //     }
        // }
    }
}