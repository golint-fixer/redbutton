cd $GOPATH/src/redbutton

unformattedFiles=`gofmt -l *`;


if [ "$unformattedFiles" != '' ]
then
    echo "Check formatting for files:"
    echo "$unformattedFiles";
    exit 1;
fi

checkImports=`goimports -l *`;

if [ "$checkImports" != '' ]
then
    echo "Check imports sections for files:"
    echo "$checkImports";
    exit 1;
fi
