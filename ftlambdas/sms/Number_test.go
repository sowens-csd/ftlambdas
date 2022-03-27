package sms

import (
	"testing"
)

const (
	internalPhone1 = "+15555555555"
	plivoPhone1    = "15555555555"
)

func TestPlusStrippedForPlivoUse(t *testing.T) {
	if toPlivoNumber(internalPhone1) != plivoPhone1 {
		t.Errorf("Expected %s, was %s", plivoPhone1, toPlivoNumber(internalPhone1))
	}
}

func TestNumberWithoutPlusUnaffected(t *testing.T) {
	if toPlivoNumber(plivoPhone1) != plivoPhone1 {
		t.Errorf("Expected %s, was %s", plivoPhone1, toPlivoNumber(plivoPhone1))
	}
}

func TestEmptyNumberUnaffected(t *testing.T) {
	if toPlivoNumber("") != "" {
		t.Errorf("Expected %s, was %s", "", toPlivoNumber(""))
	}
}

func TestAddsPlusIfNotThere(t *testing.T) {
	if plivoToInternalNumber(plivoPhone1) != internalPhone1 {
		t.Errorf("Expected %s, was %s", internalPhone1, plivoToInternalNumber(plivoPhone1))
	}
}

func TestUnchangedIfPlusAlreadyThere(t *testing.T) {
	if plivoToInternalNumber(internalPhone1) != internalPhone1 {
		t.Errorf("Expected %s, was %s", internalPhone1, plivoToInternalNumber(internalPhone1))
	}
}

func TestUnchangedIfEmpty(t *testing.T) {
	if plivoToInternalNumber("") != "" {
		t.Errorf("Expected %s, was %s", "", plivoToInternalNumber(""))
	}
}
