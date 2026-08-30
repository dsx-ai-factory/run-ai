GitSHA=`git rev-parse HEAD`
Date=`date "+%Y-%m-%d %H:%M:%S"`

build: build.darwin build.amd64

build.darwin:
	go build -o=go-acc.darwin .

build.amd64:
	GOOS=linux GOARCH=amd64 go build -o=go-acc.amd64 .

clean:
	rm -rf go-acc.darwin go-acc.amd64
