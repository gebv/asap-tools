release:
	goreleaser release --skip-validate --rm-dist
build:
	time goreleaser build --debug --snapshot --skip-publish --rm-dist --skip-validate --timeout 3m

# Extract the ngrok url for the current session
# Searches for a line for eg.
# t=2022-01-15T17:19:48+0000 lvl=info msg="started tunnel" obj=tunnels name="redirect-webhooks (http)" addr=http://debug-webhooks:8080 url=http://5b14-109-252-134-60.eu.ngrok.io
PARSENGROKURL ?= `docker logs asap-tools_ngrock_1 | grep -e "started tunnel.*ngrok.io" | head -1 | sed -n -e "s/.*url=http:\/\/\(.*\).ngrok.io$$/\1/p"`

debug-webhooks:
	docker-compose up -d ngrock
	sleep 2 # starting and configuring the ngrok agent
	@echo ">> Your ngrok url:"
	@echo "\thttps://$(PARSENGROKURL).ngrok.io"

	docker-compose up -d debug-webhooks

	# add firebase emulator
