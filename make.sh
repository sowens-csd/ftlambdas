export GOOS=linux
export GOARCH=amd64
set -e
go build -ldflags="-s -w" -o ./r2/bin/passwordlessAuthorizer ./r2/passwordlessAuthorizer
go build -ldflags="-s -w" -o ./r2/bin/jwtAuthorizer ./r2/jwtAuthorizer
go build -ldflags="-s -w" -o ./r2/bin/signup ./r2/signup
go build -ldflags="-s -w" -o ./r2/bin/signup_verify ./r2/signup_verify
go build -ldflags="-s -w" -o ./r2/bin/addDevice_verify ./r2/addDevice_verify
go build -ldflags="-s -w" -o ./r2/bin/user ./r2/user
go build -ldflags="-s -w" -o ./r2/bin/user_post ./r2/user_post
go build -ldflags="-s -w" -o ./r2/bin/user_by_email ./r2/user_by_email
go build -ldflags="-s -w" -o ./r2/bin/verify_receipt ./r2/verify_receipt
go build -ldflags="-s -w" -o ./r2/bin/new_story ./r2/new_story
go build -ldflags="-s -w" -o ./r2/bin/stories ./r2/stories
go build -ldflags="-s -w" -o ./r2/bin/sms_getnumber ./r2/sms_getnumber
go build -ldflags="-s -w" -o ./r2/bin/sms_assignnumber ./r2/sms_assignnumber
go build -ldflags="-s -w" -o ./r2/bin/sms_receive ./r2/sms_receive
go build -ldflags="-s -w" -o ./r2/bin/remote_command ./r2/remote_command
go build -ldflags="-s -w" -o ./r2/bin/new_group ./r2/new_group
go build -ldflags="-s -w" -o ./r2/bin/group_member ./r2/group_member
go build -ldflags="-s -w" -o ./r2/bin/group_members ./r2/group_members
go build -ldflags="-s -w" -o ./r2/bin/delete_group ./r2/delete_group
go build -ldflags="-s -w" -o ./r2/bin/device_token ./r2/device_token
go build -ldflags="-s -w" -o ./r2/bin/app_usage ./r2/app_usage
go build -ldflags="-s -w" -o ./r2/bin/p2p_lookup ./r2/app_usage
go build -ldflags="-s -w" -o ./r2/bin/webrtc_channel ./r2/webrtc_channel
go build -ldflags="-s -w" -o ./r2/bin/webrtc_service ./r2/webrtc_service
go build -ldflags="-s -w" -o ./r2/bin/connection_handler ./r2/connection_handler
go build -ldflags="-s -w" -o ./r2/bin/disconnection_handler ./r2/disconnection_handler
go build -ldflags="-s -w" -o ./r2/bin/websocketAuthorizer ./r2/websocketAuthorizer
