package notification

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

type callDestination struct {
	Email     string `json:"email"`
	UserIDStr string `json:"userIdStr"`
	Nickname  string `json:"nickname"`
}

type callInfo struct {
	SessionID       string            `json:"sessionId"`
	CallChannel     string            `json:"callChannel,omitempty"`
	DeviceID        string            `json:"deviceId,omitempty"`
	CallingUserName string            `json:"callingUserName,omitempty"`
	CallType        string            `json:"callType,omitempty"`
	AddMember       string            `json:"addMember,omitempty"`
	CallMembers     []callDestination `json:"callMembers,omitempty"`
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
	CommandChannel   string         `json:"commandChannel,omitempty"`
	StoryInfo        *storyInfo     `json:"storyInfo,omitempty"`
	CallInfo         *callInfo      `json:"callInfo,omitempty"`
	Result           *commandResult `json:"result,omitempty"`
}
