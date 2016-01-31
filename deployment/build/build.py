import os
import subprocess

# obtain GOROOT (go installation root folder) from environment variable
goRoot = os.environ['GOROOT']

# assume gopath is in place
goPath = os.environ['GOPATH']
goBinary = goRoot+"/bin/go"

# find project roots relatively
projectRoot = os.path.abspath(os.path.dirname(__file__)+"../../..")
buildOutputRoot = projectRoot+"/bin"

# output binary for server
serviceBinary = buildOutputRoot+"/redbutton-server"


def goInvocation(*args):
    args = [goBinary]+list(args)
    result = subprocess.Popen(args, cwd=projectRoot)
    result.communicate()
    return_code = result.returncode
    if return_code != 0:
        print "go call failed"
        exit(return_code)

goInvocation("get", "./...")
goInvocation('build', '-o', serviceBinary, 'github.com/viktorasm/redbutton/server/main')