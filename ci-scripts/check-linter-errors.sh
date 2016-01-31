# build script - checks that linter spews no errors
# fail on most linter errors except the one about the comments.

linterErrors=`golint ./...| grep -v "should have comment or be unexported"`;

if [ "$linterErrors" != '' ]
then
    echo "Fix linter errors:"
    echo "$linterErrors";
    exit 1;
fi