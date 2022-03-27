package sms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/plivo/plivo-go"
	"github.com/sowens-csd/ftlambdas/awsproxy"
)

// AvailableNumber information about a single phone number
type AvailableNumber struct {
	AddressRequirements string `json:"address_requirements"`
	FriendlyName        string `json:"friendly_name"`
	ISOCountry          string `json:"iso_country"`
	PhoneNumber         string `json:"phone_number"`
	PostalCode          string `json:"postal_code"`
	Region              string `json:"region"`
}

// AvailableNumbers information about a set of phone numbers
type AvailableNumbers struct {
	AvailableNumbers []AvailableNumber `json:"available_phone_numbers"`
}

type provisionResponse struct {
	FriendlyName   string `json:"friendly_name"`
	PhoneNumber    string `json:"phone_number"`
	PhoneNumberSID string `json:"sid"`
	Status         string `json:"status"`
}

const twilioAPIEndpoint = "https://api.twilio.com/2010-04-01/Accounts/%s"
const testTwilioAPIEndpoint = "https://api.twilio.com/2010-04-01/Accounts/ACed007ac23d8fbd8416211b44285208fa"
const testTwilioSID = "ACed007ac23d8fbd8416211b44285208fa"
const testTwilioSecret = "dd65445d8fb0e261cf1fab3354d40158"

const getNumbersPath = "AvailablePhoneNumbers"
const standardOptions = "&SmsEnabled=true&ExcludeAllAddressRequired=true"
const mobileAction = "Mobile.json?PageSize=20" + standardOptions
const localAction = "Local.json?PageSize=20" + standardOptions
const assignNumberAction = "IncomingPhoneNumbers.json"

// GetAvailableNumbers returns phone numbers available for prorivisioning in the given country and area code
func GetAvailableNumbers(ftCtx awsproxy.FTContext, isoCountry string, areacode string, client *http.Client) (AvailableNumbers, error) {
	emptyResponse := AvailableNumbers{}
	if len(isoCountry) == 0 {
		return emptyResponse, fmt.Errorf("Country code required")
	}
	// return getAvailableTwilioNumbers(ftCtx, isoCountry, areacode, client)
	return getAvailablePlivoNumbers(ftCtx, isoCountry, areacode)

}

func getAvailableTwilioNumbers(ftCtx awsproxy.FTContext, isoCountry string, areacode string, client *http.Client) (AvailableNumbers, error) {
	emptyResponse := AvailableNumbers{}
	twilioAccount, twilioSID, twilioSecret, twilioAppSID := awsproxy.TwilioParameters()
	if len(twilioAccount) == 0 || len(twilioSID) == 0 || len(twilioSecret) == 0 || len(twilioAppSID) == 0 {
		return emptyResponse, fmt.Errorf("Account, SID, Secret, or App SID not configured in Parameter Store or SetupAccessParameters not called")
	}
	apiEndpoint := fmt.Sprintf(twilioAPIEndpoint, twilioAccount)
	requestURL := fmt.Sprintf("%s/%s/%s/%s", apiEndpoint, getNumbersPath, isoCountry, mobileAction)
	if isoCountry == "US" || isoCountry == "CA" {
		requestURL = fmt.Sprintf("%s/%s/%s/%s", apiEndpoint, getNumbersPath, isoCountry, localAction)
		if len(areacode) > 0 {
			requestURL += "&AreaCode=" + areacode
		}
	}

	req, err := http.NewRequest("GET", requestURL, nil)
	if nil != err {
		ftCtx.RequestLogger.Info().Str("url", requestURL).Err(err).Msg("NewRequest failed")
		return emptyResponse, err
	}
	req.SetBasicAuth(twilioSID, twilioSecret)
	resp, err := client.Do(req)
	if nil != err {
		return emptyResponse, err
	}
	ftCtx.RequestLogger.Debug().Str("url", requestURL).Msg(fmt.Sprintf("HTTP Status on available numbers call %d", resp.StatusCode))
	if http.StatusOK != resp.StatusCode {
		return emptyResponse, fmt.Errorf("Non 200 status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return emptyResponse, err
	}
	stringBody := string(respBody)
	inputJSON := []byte(string(stringBody))
	var numbers AvailableNumbers
	err = json.Unmarshal(inputJSON, &numbers)
	if nil != err {
		return emptyResponse, err
	}
	ftCtx.RequestLogger.Debug().Msg("Got a valid response.")
	return numbers, nil
}

/// Search Plivo for available phone numbers given the country and prefix provided.
func getAvailablePlivoNumbers(ftCtx awsproxy.FTContext, isoCountry string, areacode string) (AvailableNumbers, error) {
	emptyResponse := AvailableNumbers{}
	plivoClient, err := getPlivoClient(ftCtx)
	if err != nil {
		return emptyResponse, err
	}
	response, err := plivoClient.PhoneNumbers.List(
		plivo.PhoneNumberListParams{
			CountryISO: isoCountry,
			Pattern:    areacode,
		},
	)
	if err != nil {
		return emptyResponse, err
	}
	var phoneNumbers []AvailableNumber
	for _, phoneNumber := range response.Objects {
		ftCtx.RequestLogger.Info().Str("phone", phoneNumber.Number).Msg("Plivo #.")
		formatted := plivoToInternalNumber(phoneNumber.Number)
		phoneNumbers = append(phoneNumbers, AvailableNumber{
			PhoneNumber:  formatted,
			FriendlyName: formatted,
			ISOCountry:   phoneNumber.Country,
			Region:       phoneNumber.Region})

	}
	return AvailableNumbers{AvailableNumbers: phoneNumbers}, nil
}

// AssignNumber assigns the requested number to the online user account
func AssignNumber(ftCtx awsproxy.FTContext, phoneNumber string, isMock bool, client *http.Client) (string, string, error) {
	assignedPhoneNumber := ""
	phoneNumberSID := ""
	ftCtx.RequestLogger.Debug().Msg("Trying to assign phone number.")
	// phoneNumberSID, phoneNumber, err = provisionWithTwilio(ftCtx, request.PhoneNumber, request.IsMock, client)
	phoneNumberSID, assignedPhoneNumber, err := provisionWithPlivo(ftCtx, phoneNumber, isMock)
	if nil != err {
		return "", "", err
	}
	if phoneNumberSID == "" || assignedPhoneNumber == "" {
		ftCtx.RequestLogger.Info().Str("phoneNumber", assignedPhoneNumber).Str("phoneNumberSID", phoneNumberSID).Msg("Phone number or SID missing.")
		return "", "", fmt.Errorf("Failed to assign number")
	}
	return phoneNumberSID, assignedPhoneNumber, nil
}

func provisionWithTwilio(ftCtx awsproxy.FTContext, phoneNumber string, mockRequest bool, client *http.Client) (string, string, error) {
	twilioAccount, twilioSID, twilioSecret, twilioAppSID := awsproxy.TwilioParameters()
	if len(twilioAccount) == 0 || len(twilioSID) == 0 || len(twilioSecret) == 0 || len(twilioAppSID) == 0 {
		return "", "", fmt.Errorf("Account, SID, Secret, or App SID not configured in Parameter Store or SetupAccessParameters not called")
	}
	provisionRequest := url.Values{}
	provisionRequest.Set("PhoneNumber", phoneNumber)
	provisionRequest.Set("SmsApplicationSid", twilioAppSID)
	apiEndpoint := fmt.Sprintf(twilioAPIEndpoint, twilioAccount)
	if mockRequest {
		apiEndpoint = testTwilioAPIEndpoint
		twilioSID = testTwilioSID
		twilioSecret = testTwilioSecret
	}
	requestURL := fmt.Sprintf("%s/%s", apiEndpoint, assignNumberAction)
	ftCtx.RequestLogger.Debug().Str("requestURL", requestURL).Msg("making Twilio request.")
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(provisionRequest.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(provisionRequest.Encode())))
	req.SetBasicAuth(twilioSID, twilioSecret)
	resp, err := client.Do(req)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("Error posting to Twilio.")
		return "", "", err
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("HTTP Status on provision call %d", resp.StatusCode))
	if http.StatusOK != resp.StatusCode && http.StatusCreated != resp.StatusCode {
		ftCtx.RequestLogger.Info().Msg(fmt.Sprintf("Twilio status %d.", resp.StatusCode))
		return "", "", fmt.Errorf("Non 200/201 status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("Error reading body from Twilio.")
		return "", "", err
	}
	stringBody := string(respBody)
	fmt.Printf(stringBody)
	responseJSON := []byte(string(stringBody))
	var provisionResponse provisionResponse
	err = json.Unmarshal(responseJSON, &provisionResponse)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("Error unmarshaling Twilio response.")
		return "", "", err
	}
	return provisionResponse.PhoneNumberSID, provisionResponse.PhoneNumber, nil

}

/// Given a phone number to buy either successfully buy it or return an explanatory error
func provisionWithPlivo(ftCtx awsproxy.FTContext, phoneNumber string, mockRequest bool) (string, string, error) {
	plivoClient, err := getPlivoClient(ftCtx)
	if err != nil {
		return "", "", err
	}
	provisionNumber := toPlivoNumber(phoneNumber)
	appSID := awsproxy.PlivoAppSID(ftCtx.Context)
	if len(appSID) == 0 {
		ftCtx.RequestLogger.Info().Msg("Plivo AppSID not set")
		return "", "", fmt.Errorf("Plivo AppSID not set")
	}

	ftCtx.RequestLogger.Info().Str("phone", provisionNumber).Str("appSID", appSID).Msg("Plivo provision")
	var assignedNumber string
	if !mockRequest {
		response, err := plivoClient.PhoneNumbers.Create(
			provisionNumber,
			plivo.PhoneNumberCreateParams{AppID: appSID},
		)
		if err != nil {
			return "", "", err
		}
		if len(response.Numbers) == 0 {
			return "", "", fmt.Errorf("create phone number failed")
		}
		if len(response.Numbers) > 1 {
			return "", "", fmt.Errorf("create phone number returned too many numbers")
		}
		if strings.ToLower(response.Numbers[0].Status) != "success" {
			return "", "", fmt.Errorf("create phone number status %s", response.Numbers[0].Status)
		}
		assignedNumber = plivoToInternalNumber(response.Numbers[0].Number)
	} else {
		assignedNumber = provisionNumber
	}
	return assignedNumber, assignedNumber, nil
}

// DeleteNumber deprovisions the given number from Plivo / Twilio
func DeleteNumber(ftCtx awsproxy.FTContext, number string) error {
	plivoClient, err := getPlivoClient(ftCtx)
	if nil != err {
		return err
	}
	err = plivoClient.Numbers.Delete(toPlivoNumber(number))
	return err
}

func getPlivoClient(ftCtx awsproxy.FTContext) (*plivo.Client, error) {
	plivoAccount, plivoSID, plivoSecret := awsproxy.PlivoParameters(ftCtx.Context)
	if len(plivoAccount) == 0 || len(plivoSID) == 0 || len(plivoSecret) == 0 {
		return nil, fmt.Errorf("Account, SID, or Secret not configured in Parameter Store or SetupAccessParameters not called")
	}
	client, err := plivo.NewClient(plivoSID, plivoSecret, &plivo.ClientOptions{})
	if err != nil {
		return nil, err
	}
	return client, err
}

func toPlivoNumber(phoneNumber string) string {
	provisionNumber := phoneNumber
	if strings.HasPrefix(provisionNumber, "+") {
		runes := []rune(provisionNumber)
		provisionNumber = string(runes[1:])
	}
	return provisionNumber
}

func plivoToInternalNumber(plivoNumber string) string {
	if len(plivoNumber) == 0 || strings.HasPrefix(plivoNumber, "+") {
		return plivoNumber
	}
	return fmt.Sprintf("+%s", plivoNumber)
}
