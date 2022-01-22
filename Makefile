release:
	goreleaser release --skip-validate --rm-dist
build:
	time goreleaser build --debug --snapshot --skip-publish --rm-dist --skip-validate --timeout 3m
