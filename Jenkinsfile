#!groovy

node() {

    def projectName
    def version
    def containerRegistry
    def imageName
    def detailedName
    def latestName

    def shortCommit

    stage('Checkout') {
        def scmVars = checkout scm
        shortCommit = scmVars.GIT_COMMIT.take(7)

        projectName = appInfo('name')
        version = appInfo('version')
        containerRegistry = appInfo('registry')
        
        imageName = appInfo('image')
        detailedName = "$imageName-$shortCommit"
        latestName = "$containerRegistry/$projectName:latest"
    }

    stage('Test') {
        def testImage = "$projectName-test:$version"
        sh "docker build -t $testImage -f Dockerfile.test ."
        sh "docker run --rm $testImage"
        sh "docker rmi $testImage"
    }

}


def appInfo(String command) {
    return sh(script: "appv $command", returnStdout: true).trim()
}