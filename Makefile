.PHONY: deploy build remove logs info test

install:
	npm install
	go mod tidy

build:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/bootstrap cmd/lambda/lambda-main.go

deploy:
	serverless deploy

deploy-prod:
	serverless deploy --stage prod

remove:
	serverless remove

logs:
	serverless logs -f api -t

info:
	serverless info

test-local:
	# Para testing local, puedes usar tools like `aws-lambda-go-local`
	go run main.go

test:
	go test -v -timeout 5m -coverprofile=coverage.out ./...

clean:
	rm -f bootstrap
	rm -f *.zip