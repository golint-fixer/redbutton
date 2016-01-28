'''
TODO: cleanup once ansible deployment is in place
'''
import os
import sys
import getopt

import fabric.api as fab

targetServer = 'never-crashed.com'
targetFolder = '~/redbutton.svc/'
goRoot = os.environ['GOROOT']
workspace = os.path.dirname(__file__)
serviceBinary = workspace+"/bin/redbutton-server"
foreverId = "redbutton-server"

def createRemoteFile(file,*contents):
    fab.run("rm -rf "+file)
    for i in contents:
        fab.run("echo '{}' >> {}".format(i,file))

def buildServer():
    with fab.shell_env(GOPATH=workspace+"/server",GOROOT=goRoot):
        with fab.lcd(workspace):
            with fab.lcd("server/src/redbutton"):
                fab.local(goRoot+"/bin/go get ./...")
            fab.local(goRoot+"/bin/go build -o "+serviceBinary+" redbutton/main")

def uploadStuff():
    fab.run("rm -rf "+targetFolder)
    fab.run("mkdir -p "+targetFolder)
    fab.put(serviceBinary,targetFolder+"/")
    with fab.cd(targetFolder):
        fab.run("chmod 700 "+os.path.basename(serviceBinary))

    fab.run("mkdir -p "+targetFolder+"ui/")
    with fab.lcd(workspace+"/ui"):
        fab.put("index.html",targetFolder+"ui/")
        fab.put("bower.json",targetFolder+"ui/")
        fab.put("style",targetFolder+"ui/")
        fab.put("pages",targetFolder+"ui/")
        fab.put("js",targetFolder+"ui/")

    with fab.cd(targetFolder+"ui/"):
        fab.run("bower install")

def createLaunchScripts():
    # a script to run a server
    launchScript = "run.sh"
    with fab.cd(targetFolder):
        createRemoteFile(launchScript, "cd {};PORT=8899 REDBUTTON_UIDIR=./ui ./{}".format(targetFolder,os.path.basename(serviceBinary)))
        fab.run("chmod 700 "+launchScript)


    # a shortcut for running application in background as Forever's item
    restartScript = "restart_service.sh"
    with fab.cd(targetFolder):
        createRemoteFile(restartScript,
                              'forever stop {};'.format(foreverId),
                              'forever start --uid {foreverId} -a -l {targetFolder}/redbutton.log -c /bin/bash {targetFolder}/{launchScript}'.format(foreverId=foreverId,targetFolder=targetFolder,launchScript=launchScript)
                              )
        fab.run("chmod 700 "+restartScript)
        fab.run("./"+restartScript)



fab.env.hosts = [targetServer]
fab.execute(buildServer)
fab.execute(uploadStuff)
fab.execute(createLaunchScripts)

fab.puts("done")
