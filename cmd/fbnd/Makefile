linux-amd64:
	GOOS=linux GOARCH=amd64 go build \
		 -ldflags "-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"'" \
		 .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build \
		 -ldflags "-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"'" \

windows-amd64:
	GOOS=windows GOARCH=amd64 go build \
		 -ldflags "-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"'" \
		 .

linux-amd64-release:
	GOOS=linux GOARCH=amd64 go build \
		 -o fbnd-linux-amd64-release \
		 -ldflags "-w -s \
		 	-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"' \
			-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.version="$$(git describe --tags)"'" \
		 .

darwin-amd64-release:
	GOOS=darwin GOARCH=amd64 go build \
		 -o fbnd-darwin-amd64-release \
		 -ldflags "-w -s \
		 	-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"' \
			-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.version="$$(git describe --tags)"'" \
		 .

windows-amd64-release:
	GOOS=windows GOARCH=amd64 go build \
		 -o fbnd-windows-amd64-release \
		 -ldflags "-w -s \
		 	-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.commitID="$$(git rev-parse HEAD)"' \
			-X 'github.com/n9v9/fbnd/cmd/fbnd/cmd.version="$$(git describe --tags)"'" \
		 .

zipped-releases: linux-amd64-release darwin-amd64-release windows-amd64-release
	# Linux
	mv fbnd-linux-amd64-release fbnd
	tar -czvf "fbnd-$$(git describe --tags)-linux-amd-64.tar.gz" fbnd
	rm fbnd
	# Darwin
	mv fbnd-darwin-amd64-release fbnd
	tar -czvf "fbnd-$$(git describe --tags)-darwin-amd-64.tar.gz" fbnd
	rm fbnd
	# Windows
	mv fbnd-windows-amd64-release "fbnd.exe"
	zip -r "fbnd-$$(git describe --tags)-windows-amd-64.zip" "fbnd.exe"
	rm "fbnd.exe"
