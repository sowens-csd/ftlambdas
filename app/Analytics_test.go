package app

import (
	"fmt"
	"testing"
)

const (
	validSample = `[{"appId":"app1"},{"appId":"app2"}]`
)

func TestMakesExpectedRecords(t *testing.T) {
	records, err := makeRecords(validSample)
	if nil != err {
		t.Errorf("Error making records %s", err.Error())
	}
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
	for _, record := range records {
		recordData := string(record.Data)
		fmt.Printf("Record: %s\n", recordData)
	}
}
