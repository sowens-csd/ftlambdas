package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/plivo/plivo-go"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sharing"
)

type googleCloudResult struct {
	MessageId string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

type googleCloudResponse struct {
	MulticastID int64               `json:"multicast_id"`
	Success     int                 `json:"success"`
	Failure     int                 `json:"failure"`
	Results     []googleCloudResult `json:"results"`
}

type googleCloudMessage struct {
	GCM string
}

type smsMedia struct {
	MimeType string `json:"mimeType"`
	MediaURL string `json:"mediaUrl"`
}

type smsDetails struct {
	NotificationType string `json:"notificationType"`
	MessageID        string `json:"messageId"`
	SentFrom         string `json:"sentFrom"`
	SentTo           string `json:"sentTo"`
	ReceivedAt       string `json:"receivedAt"`
	MsgContent       string `json:"msgContent"`
	AttachedImages   string `json:"attachedImages,omitempty"`
}

type smsNotification struct {
	Data smsDetails `json:"data"`
}

type storyInfo struct {
	StoryID         string `json:"storyId"`
	SlideNumber     int    `json:"slideNumber"`
	CallingUserName string `json:"callingUserName,omitempty"`
	CallType        string `json:"callType,omitempty"`
}

type callInfo struct {
	SessionID       string `json:"sessionId"`
	CallChannel     string `json:"callChannel,omitempty"`
	DeviceID        string `json:"deviceId,omitempty"`
	CallingUserName string `json:"callingUserName,omitempty"`
	CallType        string `json:"callType,omitempty"`
}

type commandResult struct {
	Status    string `json:"status"`
	ErrorCode string `json:"errorCode,omitempty"`
	Message   string `json:"message,omitempty"`
}

// CommandType must be one of:
// - 'showStory'
// - 'changeSlide'
type commandDetails struct {
	NotificationType string         `json:"notificationType"`
	SessionID        string         `json:"sessionId"`
	ToUser           string         `json:"toUser"`
	FromUser         string         `json:"fromUser"`
	CommandType      string         `json:"commandType"`
	CommandID        string         `json:"commandId"`
	RequestID        string         `json:"requestId,omitempty"`
	StoryInfo        *storyInfo     `json:"storyInfo,omitempty"`
	CallInfo         *callInfo      `json:"callInfo,omitempty"`
	Result           *commandResult `json:"result,omitempty"`
}

type pushNotification struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Badge string `json:"badge,omitempty"`
}

type commandNotification struct {
	Data             commandDetails    `json:"data"`
	PushNotification *pushNotification `json:"notification,omitempty"`
}

type fcmCommandNotification struct {
	To               string            `json:"to"`
	ContentAvailable *bool             `json:"content_available,omitempty"`
	APNSPriority     *int              `json:"apns-priority,omitempty"`
	Data             *interface{}      `json:"data,omitempty"`
	PushNotification *pushNotification `json:"notification,omitempty"`
}

type plivoMediaInfo struct {
	MediaURL    string `json:"media_url"`
	ContentType string `json:"content_type"`
}

// SendFromCommand creates a new push notification that represents a remote command
func SendFromCommand(ftCtx awsproxy.FTContext, command string) error {
	commandJSON := []byte(command)
	var notification commandNotification
	err := json.Unmarshal(commandJSON, &notification)
	if nil != err {
		return err
	}
	details := notification.Data
	areInSame, err := sharing.AreInSameGroup(ftCtx, details.ToUser)
	if nil != err {
		return err
	}
	if !areInSame {
		return fmt.Errorf("Current user is not an active member in a group with DestinationUserID")
	}
	destinationUser, err := sharing.LoadOnlineUser(ftCtx, details.ToUser)
	if nil != err {
		return err
	}

	err = SendFCM(ftCtx, notification.Data, notification.PushNotification, destinationUser, &http.Client{Timeout: 30 * time.Second})

	return err
}

// SendAuthVerifyCommand creates a new push notification that represents a remote command
func SendAuthVerifyCommand(ftCtx awsproxy.FTContext, requestID string, destinationUser *sharing.OnlineUser, client *http.Client) error {
	var notification = commandNotification{
		Data: commandDetails{
			NotificationType: "remoteCommand",
			CommandType:      "verifyAuth",
			ToUser:           destinationUser.ID,
			FromUser:         destinationUser.ID,
			RequestID:        requestID,
		},
		PushNotification: &pushNotification{
			Title: "Add Device",
			Body:  fmt.Sprintf("Code: %s", requestID),
		},
	}

	return SendFCM(ftCtx, notification.Data, notification.PushNotification, destinationUser, client)
}

// SendAlert creates a new push notification that represents a notification to display
func SendAlert(ftCtx awsproxy.FTContext, title, alert string, destinationUser *sharing.OnlineUser, client *http.Client) error {
	var notification = commandNotification{
		PushNotification: &pushNotification{
			Title: title,
			Body:  alert,
		},
	}
	if isOldUser(destinationUser) {
		structuredContent, _ := json.Marshal(&notification)
		gcm := googleCloudMessage{GCM: string(structuredContent)}
		msgContent, _ := json.Marshal(&gcm)
		strContent := string(msgContent)
		Send(ftCtx, strContent, destinationUser)
		return nil
	}
	return SendFCM(ftCtx, nil, notification.PushNotification, destinationUser, client)
}

// SendStoryChangeCommand creates a new push notification that represents a remote command
func SendStoryChangeCommand(ftCtx awsproxy.FTContext, storyID string, destinationUser *sharing.OnlineUser, client *http.Client) error {
	var notification = commandNotification{
		Data: commandDetails{
			NotificationType: "remoteCommand",
			CommandType:      "storyChange",
			ToUser:           destinationUser.ID,
			FromUser:         destinationUser.ID,
			StoryInfo:        &storyInfo{StoryID: storyID},
		},
		PushNotification: &pushNotification{
			Title: "Story Update",
			Body:  "There's new story content in Folktells.",
		},
	}

	ftCtx.RequestLogger.Debug().Str("storyID", storyID).Msg("Building notification")
	return SendFCM(ftCtx, notification.Data, notification.PushNotification, destinationUser, client)
}

// SendFromSMS creates a new push notification based on the information in the provided SMS
// message.
func SendFromSMS(ftCtx awsproxy.FTContext, smsData string, client *http.Client) error {
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("SMS body: %s", smsData))
	smsDetails, err := buildSmsDetails(ftCtx, smsData)
	if nil != err {
		return err
	}
	destinationUser, err := sharing.LoadOnlineUserByPhone(ftCtx, smsDetails.SentTo)
	if nil != err {
		return err
	}

	if isOldUser(destinationUser) {
		msgContent, err := buildMessageContent(*smsDetails)
		if nil != err {
			return err
		}
		Send(ftCtx, msgContent, destinationUser)
		return nil
	}
	return SendFCM(ftCtx, smsDetails, nil, destinationUser, client)
}

// Send a notification to the endpoints for this user
func Send(ftCtx awsproxy.FTContext, message string, onlineUser *sharing.OnlineUser) {
	ftCtx.RequestLogger.Debug().Msg("Send message")
	snsClient := getSNSClient(ftCtx)
	for _, deviceToken := range onlineUser.DeviceTokens {
		ftCtx.RequestLogger.Debug().Str("gcmMessage", message).Str("appID", deviceToken.AppInstallID).Str("SNSEndpoint", deviceToken.SNSEndpoint).Msg("Publish to")
		input := &sns.PublishInput{
			MessageStructure: aws.String("json"),
			Message:          aws.String(message),
			TargetArn:        aws.String(deviceToken.SNSEndpoint),
		}
		_, err := snsClient.Publish(ftCtx.Context, input)
		if err != nil {
			ftCtx.RequestLogger.Debug().Str("appID", deviceToken.AppInstallID).Str("SNSEndpoint", deviceToken.SNSEndpoint).Err(err).Msg("Publish failed ")
		}
	}
}

func SendFCM(ftCtx awsproxy.FTContext, data interface{}, pushNotification *pushNotification, onlineUser *sharing.OnlineUser, client *http.Client) error {
	if isOldUser(onlineUser) {
		var notification = commandNotification{
			Data:             data.(commandDetails),
			PushNotification: pushNotification,
		}
		structuredContent, _ := json.Marshal(&notification)
		gcm := googleCloudMessage{GCM: string(structuredContent)}
		msgContent, _ := json.Marshal(&gcm)
		strContent := string(msgContent)
		Send(ftCtx, strContent, onlineUser)
		return nil
	}
	fcm := fcmCommandNotification{}
	if nil != data {
		fcm.Data = &data
	}
	if nil == pushNotification {
		normal := 5
		contentAvailable := true
		fcm.APNSPriority = &normal
		fcm.ContentAvailable = &contentAvailable
	} else {
		fcm.PushNotification = pushNotification
	}
	// After the send this array will hold all of the tokens that successfully delivered
	// a message to a device. If it doesn't match the user's current token list then
	// the current token list will be udpated to match the good token list
	goodTokens := make([]sharing.DeviceNotificationToken, 0, len(onlineUser.DeviceTokens))
	for _, device := range onlineUser.DeviceTokens {
		fcm.To = device.NotificationToken
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(fcm)
		if nil != err {
			return err
		}
		ftCtx.RequestLogger.Debug().Str("payload", buf.String()).Msg("posting to fcm")
		req, err := http.NewRequest(http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(buf.Bytes()))
		if nil != err {
			return err
		}
		fcmKey := awsproxy.FCMParameters(ftCtx.Context)
		if len(fcmKey) == 0 {
			return fmt.Errorf("fcmKey environment variable not present")
		}
		req.Header.Add("Authorization", fmt.Sprintf("key=%s", fcmKey))
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if nil != err {
			return err
		}
		defer resp.Body.Close()
		if http.StatusOK == resp.StatusCode {
			respBody, err := ioutil.ReadAll(resp.Body)
			if nil == err {
				stringBody := string(respBody)
				inputJSON := []byte(string(stringBody))
				var fcmResponse googleCloudResponse
				err = json.Unmarshal(inputJSON, &fcmResponse)
				if nil == err && fcmResponse.Failure > 0 && len(fcmResponse.Results) > 0 {
					ftCtx.RequestLogger.Debug().Int("success", fcmResponse.Success).Int("failure", fcmResponse.Failure).Str("token", device.NotificationToken).Str("error", fcmResponse.Results[0].Error).Msg("got valid response")
					switch fcmResponse.Results[0].Error {
					case "MissingRegistration":
					case "InvalidRegistration":
					case "NotRegistered":
						// In any of these cases the token that was used was not correct and should be
						// removed from the array of good tokens
						ftCtx.RequestLogger.Debug().Str("error", fcmResponse.Results[0].Error).Str("token", device.NotificationToken).Msg("bad token")
						break
					default:
						ftCtx.RequestLogger.Debug().Str("error", fcmResponse.Results[0].Error).Str("token", device.NotificationToken).Msg("good token")
						goodTokens = append(goodTokens, device)
						break
					}
				} else {
					ftCtx.RequestLogger.Debug().Int("success", fcmResponse.Success).Int("failure", fcmResponse.Failure).Str("token", device.NotificationToken).Str("error", fcmResponse.Results[0].Error).Msg("good token, no failures or..")
					goodTokens = append(goodTokens, device)
				}
			} else {
				ftCtx.RequestLogger.Error().Err(err).Str("token", device.NotificationToken).Msg("good token, error reading response")
				goodTokens = append(goodTokens, device)
			}
		} else {
			ftCtx.RequestLogger.Info().Int("httpStatus", resp.StatusCode).Str("token", device.NotificationToken).Msg("send failed")
			goodTokens = append(goodTokens, device)
		}
	}
	if len(goodTokens) != len(onlineUser.DeviceTokens) {
		ftCtx.RequestLogger.Debug().Int("good", len(goodTokens)).Int("current", len(onlineUser.DeviceTokens)).Msg("updating user tokens")
		onlineUser.DeviceTokens = goodTokens
		onlineUser.UpdateDeviceTokens(ftCtx)
	}
	return nil
}

func isOldUser(onlineUser *sharing.OnlineUser) bool {
	isOld := false
	for _, device := range onlineUser.DeviceTokens {
		isOld = isOld || device.AppVersion == ""
	}
	return isOld
}

func buildMessageContent(sms smsDetails) (string, error) {
	smsNotify := smsNotification{Data: sms}
	structuredContent, _ := json.Marshal(&smsNotify)
	gcm := googleCloudMessage{GCM: string(structuredContent)}
	msgContent, _ := json.Marshal(&gcm)
	return string(msgContent), nil
}

func buildSmsDetails(ftCtx awsproxy.FTContext, smsData string) (*smsDetails, error) {
	if len(smsData) == 0 {
		return nil, fmt.Errorf("expected form data in the body")
	}
	msgValues, err := url.ParseQuery(smsData)
	if nil != err {
		return nil, err
	}

	var attachedImages []smsMedia
	var messageID string
	var content string
	var toAddress string
	var fromAddress string

	messageID = getSingleValue("MessageUUID", msgValues)
	if len(messageID) > 0 { // Plivo
		toAddress = fmt.Sprintf("+%s", getSingleValue("To", msgValues))
		fromAddress = fmt.Sprintf("+%s", getSingleValue("From", msgValues))
		content = getSingleValue("Text", msgValues)
		msgType := getSingleValue("Type", msgValues)
		noImages := true
		if msgType == "mms" {
			content = getSingleValue("Body", msgValues)
			mediaCount := getSingleValue("MediaCount", msgValues)
			if mediaCount == "" {
				return nil, fmt.Errorf("'MediaCount' is required for MMS")
			}
			numImages, err := strconv.Atoi(mediaCount)
			ftCtx.RequestLogger.Debug().Int("numImages", numImages).Msg("Looking for media")
			if nil == err && numImages > 0 {
				noImages = false
				for image := 0; image < numImages; image++ {
					mediaField := fmt.Sprintf("Media%d", image)
					mediaURL := getSingleValue(mediaField, msgValues)
					ftCtx.RequestLogger.Debug().Str("mediaField", mediaField).Str("mediaURL", mediaURL).Msg("Got media URL param")
					if mediaURL != "" {
						ftCtx.RequestLogger.Debug().Str("mediaURL", mediaURL).Msg("Looking up media 0")
						mediaResourceURL, contentType, err := plivoMediaLookup(ftCtx, mediaURL)
						if nil == err {
							attachedImages = append(attachedImages, smsMedia{MediaURL: mediaResourceURL, MimeType: contentType})
						} else {
							ftCtx.RequestLogger.Error().Err(err).Msg("Media lookup failed")
						}
					}
				}
			}
		}
		if (content == "" && noImages) || toAddress == "" || fromAddress == "" {
			return nil, fmt.Errorf("'To', 'From', 'Body'|'Text'|'Images' are required")
		}
	} else { // Twilio
		messageID = getSingleValue("MessageSid", msgValues)
		toAddress = getSingleValue("To", msgValues)
		fromAddress = getSingleValue("From", msgValues)
		content = getSingleValue("Body", msgValues)
		numMedia := getSingleValue("NumMedia", msgValues)
		if (content == "" && (numMedia == "" || numMedia == "0")) || toAddress == "" || fromAddress == "" {
			return nil, fmt.Errorf("'To', 'From', 'Body' and 'NumMedia' are required")
		}
		numImages, err := strconv.Atoi(numMedia)
		if nil == err && numImages > 0 {
			for image := 0; image < numImages; image++ {
				urlField := fmt.Sprintf("MediaUrl%d", image)
				mimeField := fmt.Sprintf("MediaContentType%d", image)
				mediaURL := getSingleValue(urlField, msgValues)
				contentType := getSingleValue(mimeField, msgValues)
				if mediaURL != "" && contentType != "" {
					attachedImages = append(attachedImages, smsMedia{MediaURL: mediaURL, MimeType: contentType})
				}
			}
		}
	}
	now := time.Now()
	receivedAt := strconv.Itoa(int(now.UTC().Unix() * 1000))
	sms := smsDetails{
		MessageID:        messageID,
		NotificationType: "sms",
		SentFrom:         fromAddress,
		SentTo:           toAddress,
		MsgContent:       content,
		ReceivedAt:       receivedAt,
	}
	if len(attachedImages) > 0 {
		imagesJSON, _ := json.Marshal(attachedImages)
		sms.AttachedImages = string(imagesJSON)
	}
	return &sms, nil
}

func plivoMediaLookup(ftCtx awsproxy.FTContext, mediaURL string) (string, string, error) {
	ftCtx.RequestLogger.Debug().Str("mediaURL", mediaURL).Msg("Looking up media 1")
	plivoAccount, plivoSID, plivoSecret := awsproxy.PlivoParameters()
	if len(plivoAccount) == 0 || len(plivoSID) == 0 || len(plivoSecret) == 0 {
		return "", "", fmt.Errorf("Account: %s, SID: %s, or Secret: %s not configured in Parameter Store or SetupAccessParameters not called", plivoAccount, plivoSID, plivoSecret)
	}
	client, err := plivo.NewClient(plivoSID, plivoSecret, &plivo.ClientOptions{})
	if err != nil {
		return "", "", err
	}
	ftCtx.RequestLogger.Debug().Str("mediaURL", mediaURL).Msg("Parsing URL")
	parsedURL, err := url.Parse(mediaURL)
	if err != nil {
		return "", "", err
	}
	mediaID := path.Base(parsedURL.Path)
	ftCtx.RequestLogger.Debug().Str("mediaID", mediaID).Msg("Got base")
	resp, err := client.Media.Get(mediaID)
	if err != nil {
		return mediaURL, "image/jpeg", nil
	}
	return mediaURL, resp.ContentType, nil
}

func getSingleValue(fieldName string, msgValues url.Values) string {
	if nil == msgValues[fieldName] || len(msgValues[fieldName]) != 1 {
		return ""
	}
	return msgValues[fieldName][0]
}
