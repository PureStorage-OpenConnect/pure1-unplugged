// Use this crazy thing instead of 'checkout scm' so we can get tags fetched...
def checkoutSCMWithTags() {
    checkout([
            $class: 'GitSCM',
            branches: scm.branches,
            doGenerateSubmoduleConfigurations: scm.doGenerateSubmoduleConfigurations,
            extensions: [
                    [$class: 'CloneOption', noTags: false, shallow: false, depth: 0, reference: ''],
                    [$class: 'CleanBeforeCheckout']
            ],
            userRemoteConfigs: scm.userRemoteConfigs
    ])
}

def runStageWithTimeout(String stageName, int timeoutMinutes, Closure stageToRun) {
    stage(stageName) {
        ansiColor('xterm') {
            timeout(timeoutMinutes) {
                stageToRun()
            }
        }
    }
}

props = [
            [
                $class: 'BuildDiscarderProperty',
                strategy: [
                    $class: 'LogRotator',
                    artifactDaysToKeepStr: '2',
                    artifactNumToKeepStr: '',
                    daysToKeepStr: '2',
                    numToKeepStr: ''
                ]
            ],
        ]

if (env.BRANCH_NAME == "master") {
    props.add(pipelineTriggers([cron('0 13 * * *')]))
}

properties(props)

node('pure1-unplugged') {
    try {
        // Always checkout the repo before running any stages
        checkoutSCMWithTags()
        parallel "Build + Test Server Binaries (Golang)": {
            runStageWithTimeout('Pull Go Builder Image', 10) {
                sh 'make pull-go-image'
            }

            runStageWithTimeout('Lint Golang', 10) {
                sh 'make go-style'
            }

            runStageWithTimeout('Build Golang bins', 10) {
                sh 'make go-bins'
            }

            runStageWithTimeout('Test Golang', 10) {
                sh 'make go-unit-tests'
            }

        }, "Build + Test Web Content (Angular)": {
            runStageWithTimeout('Pull Web Content Builder Image', 10) {
                sh 'make pull-web-image'
            }

            runStageWithTimeout('Setup Web Content', 10) {
                sh 'make web-setup'
            }

            runStageWithTimeout('Lint Web Content', 10) {
                sh 'make lint-web-content'
            }

            runStageWithTimeout('Build Web Content', 10) {
                sh 'make web-content'
            }

            runStageWithTimeout('Test Web Content', 10) {
                sh 'make test-web-content'
            }
        }

        parallel "pure1-unplugged docker image": {
            runStageWithTimeout('Build Docker Image', 20) {
                sh 'make pure1-unplugged-image'
                sh './scripts/push_pure1_unplugged_docker_image.sh'
            }
        }, "lorax-build docker image": {
            runStageWithTimeout('Build Lorax Docker Image', 20) {
                sh 'make lorax-image'
            }
        }

        runStageWithTimeout('Build Helm Chart', 10) {
            sh 'make helm-chart'
        }

        runStageWithTimeout('Build Install Bundle', 90) {
            sh 'make install-bundle'
        }

        runStageWithTimeout('Build Pure1-Unplugged RPM', 90) {
            sh 'make rpm'
        }

        runStageWithTimeout('Build Installer ISO', 90) {
            sh 'make iso'
        }

    } finally {
        archiveArtifacts allowEmptyArchive: true, artifacts: 'build/bin/**/*', defaultExcludes: false, fingerprint: false
        archiveArtifacts allowEmptyArchive: true, artifacts: 'build/gui/**/*', defaultExcludes: false, fingerprint: false
        archiveArtifacts allowEmptyArchive: true, artifacts: 'build/chart/*.tgz', defaultExcludes: false, fingerprint: false
        archiveArtifacts allowEmptyArchive: true, artifacts: 'build/iso/*.iso', defaultExcludes: false, fingerprint: true
        archiveArtifacts allowEmptyArchive: true, artifacts: 'build/bundle/*.tar.gz', defaultExcludes: false, fingerprint: true
        archiveArtifacts allowEmptyArchive: true, artifacts: 'build/bundle/*.tar.gz.sha1', defaultExcludes: false, fingerprint: true
    }
}
