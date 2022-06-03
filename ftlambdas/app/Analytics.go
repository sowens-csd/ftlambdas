package app

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/sowens-csd/ftlambdas/ftlambdas/awsproxy"
)

// AnalyticEvent a single event from the app that describes a behaviour
type AnalyticEvent struct {
	AppID          string `json:"appId" dynamodbav:"appId"`
	SessionID      string `json:"sessionId" dynamodbav:"sessionId"`
	OSVersion      string `json:"osVersion" dynamodbav:"osVersion"`
	DeviceModel    string `json:"deviceModel" dynamodbav:"deviceModel"`
	Locale         string `json:"locale" dynamodbav:"locale"`
	AppVersion     string `json:"appVersion" dynamodbav:"appVersion"`
	Timezone       string `json:"timezone,omitempty" dynamodbav:"timezone,omitempty"`
	EventType      int    `json:"eventType" dynamodbav:"eventType"`
	Target         string `json:"target" dynamodbav:"target"`
	EventAt        int    `json:"eventAt" dynamodbav:"eventAt"`
	Elapsed        int    `json:"elapsed,omitempty" dynamodbav:"elapsed,omitempty"`
	ItemCount      int    `json:"itemCount,omitempty" dynamodbav:"itemCount,omitempty"`
	EventSource    string `json:"eventSource,omitempty" dynamodbav:"eventSource,omitempty"`
	EventCompleted bool   `json:"eventCompleted,omitempty" dynamodbav:"eventCompleted,omitempty"`
	usedSpeech     bool   `json:"usedSpeech,omitempty" dynamodbav:"usedSpeech,omitempty"`
}

var analyticsStream *firehose.Client

// SaveAnalyticEvent update the data store with a new analytic event
func SaveAnalyticEvent(ftCtx awsproxy.FTContext, data string) error {
	getAnalyticsStream(ftCtx)
	streamName := os.Getenv("analyticsStream")
	recordData := []byte(data)
	record := types.Record{Data: recordData}
	recordInput := firehose.PutRecordInput{
		DeliveryStreamName: &streamName,
		Record:             &record,
	}
	fhResp, err := analyticsStream.PutRecord(ftCtx.Context, &recordInput)
	ftCtx.RequestLogger.Debug().Str("fh_record_id", *fhResp.RecordId).Msg("Put analytics")
	return err
}

// SaveAnalyticEvents update the data store with a set of new analytic events
func SaveAnalyticEvents(ftCtx awsproxy.FTContext, data string) error {
	getAnalyticsStream(ftCtx)
	streamName := os.Getenv("analyticsStream")
	ftCtx.RequestLogger.Debug().Str("stream_name", streamName).Msg("Stream setup")
	records, err := makeRecords(data)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("make analytics records error")
		return err
	}
	recordInput := firehose.PutRecordBatchInput{
		DeliveryStreamName: &streamName,
		Records:            records,
	}
	_, err = analyticsStream.PutRecordBatch(ftCtx.Context, &recordInput)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("put analytics error")
	}
	return err
}

func getAnalyticsStream(ftCtx awsproxy.FTContext) *firehose.Client {
	if nil == analyticsStream {
		analyticsStream = firehose.NewFromConfig(ftCtx.Config.(aws.Config))
		if nil == analyticsStream {
			panic("could not create analytic stream")
		}
	}
	return analyticsStream
}

func makeRecords(data string) ([]types.Record, error) {
	var events []AnalyticEvent
	eventsJSON := []byte(data)
	err := json.Unmarshal(eventsJSON, &events)
	var records []types.Record
	if err != nil {
		return records, err
	}
	for _, event := range events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return records, err
		}
		recordData := []byte(eventJSON)
		record := types.Record{Data: recordData}
		records = append(records, record)
	}
	return records, nil
}
