cd $GOPATH/src/redbutton

unformattedFiles=`gofmt -l *`;

if [ "$unformattedFiles" != '' ]
then
    echo "Check formatting for files:"
    echo "$unformattedFiles";
    exit 1;
fi