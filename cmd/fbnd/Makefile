linux-amd64:
	GOOS=linux GOARCH=amd64 go build .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build .

windows-amd64:
	GOOS=windows GOARCH=amd64 go build .

linux-amd64-release:
	GOOS=linux GOARCH=amd64 go build \
		 -ldflags "-w -s -X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"'" \
		 .

darwin-amd64-release:
	GOOS=darwin GOARCH=amd64 go build \
		 -ldflags "-w -s -X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"'" \
		 .

windows-amd64-release:
	GOOS=windows GOARCH=amd64 go build \
		 -ldflags "-w -s -X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"'" \
		 .
