.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/passwordlessAuthorizer lambdas/passwordlessAuthorizer/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/jwtAuthorizer lambdas/jwtAuthorizer/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/signup lambdas/signup/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/signup_verify lambdas/signup_verify/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/addDevice_verify lambdas/addDevice_verify/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/user lambdas/user/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/user_post lambdas/user_post/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/user_by_email lambdas/user_by_email/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/verify_receipt lambdas/verify_receipt/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/new_story lambdas/new_story/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/stories lambdas/stories/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/sms_getnumber lambdas/sms_getnumber/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/sms_assignnumber lambdas/sms_assignnumber/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/sms_receive lambdas/sms_receive/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/remote_command lambdas/remote_command/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/new_group lambdas/new_group/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/group_member lambdas/group_member/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/group_members lambdas/group_members/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete_group lambdas/delete_group/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/device_token lambdas/device_token/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/app_usage lambdas/app_usage/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/p2p_signup lambdas/p2p_lookup/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/webrtc_channel lambdas/webrtc_channel/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/webrtc_service lambdas/webrtc_service/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
