set -e
#go get code.google.com/p/go.tools/cmd/cover
go get github.com/t-yuki/gocover-cobertura
go get github.com/nordligulv/go-for-teamcity 
go get github.com/kisielk/errcheck

# Automatic checks
#test -z "$(gofmt -l -w .     | tee /dev/stderr)"
#test -z "$(goimports -l -w . | tee /dev/stderr)"
#test -z "$(golint .          | tee /dev/stderr)"

go vet ./... 2>&1 | tee /home/lazada/go/src/lazada_api/govet.out 
go-for-teamcity -vet=/home/lazada/go/src/lazada_api/govet.out > /home/lazada/go/src/lazada_api/govet.xml
errcheck ./... > /home/lazada/go/src/lazada_api/errcheck.out
go-for-teamcity -errcheck=/home/lazada/go/src/lazada_api/errcheck.out > /home/lazada/go/src/lazada_api/errcheck.xml

GORACE="log_path=/home/lazada/go-race-reports/report strip_path_prefix=/home/lazada/go/src/lazada_api history_size=2" go test -race ./...
 
go-for-teamcity -log_path=/home/lazada/go-race-reports/report > /home/lazada/go/src/lazada_api/datarace.html

# Run test coverage on each subdirectories and merge the coverage profile.
 
echo "mode: count" > profile.cov
 
# Standard go tooling behavior is to ignore dirs with leading underscors
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d);
do
if ls $dir/*.go &> /dev/null; then
    go test -covermode=count -coverprofile=$dir/profile.tmp $dir 
    if [ -f $dir/profile.tmp ]
    then
        cat $dir/profile.tmp | tail -n +2 >> profile.cov
        rm $dir/profile.tmp
    fi
fi
done
 
#go tool cover -func profile.cov
go tool cover -html=profile.cov -o coverage.html 
gocover-cobertura profile.cov coverage.xml
