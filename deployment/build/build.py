import os
import subprocess

# obtain GOROOT (go installation root folder) from environment variable
goRoot = os.environ['GOROOT']
goBinary = goRoot+"/bin/go"

# find project roots relatively
projectRoot = os.path.abspath(os.path.dirname(__file__)+"../../..")
goWorkspace = projectRoot+"/server"
buildOutputRoot = projectRoot+"/bin"

# output binary for server
serviceBinary = buildOutputRoot+"/redbutton-server"



def goInvocation(*args):
    environment = dict(os.environ)
    environment['GOPATH'] = goWorkspace
    args = [goBinary]+list(args)
    result = subprocess.Popen(args, env=environment, cwd=goWorkspace+"/src/redbutton")
    result.communicate()
    return_code = result.returncode
    if return_code != 0:
        print "go call failed"
        exit(return_code)

goInvocation("get", "./...")
goInvocation('build', '-o', serviceBinary, 'redbutton/main')