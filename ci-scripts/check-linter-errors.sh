cd $GOPATH/src/redbutton

# fail on most linter errors except the one about the comments.

linterErrors=`golint ./...| grep -v "should have comment or be unexported"`;

if [ "$linterErrors" != '' ]
then
    echo "Fix linter errors:"
    echo "$linterErrors";
    exit 1;
fi