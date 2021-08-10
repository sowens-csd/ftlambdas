package webrtc

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	signer "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	kvs "github.com/aws/aws-sdk-go-v2/service/kinesisvideo"
	"github.com/aws/aws-sdk-go-v2/service/kinesisvideo/types"
	"github.com/aws/aws-sdk-go/aws/session"
	kvsc "github.com/aws/aws-sdk-go/service/kinesisvideosignalingchannels"
	"github.com/sowens-csd/ftlambdas/awsproxy"
)

type WebRTCIceServer struct {
	Password string    `json:"password"`
	Username string    `json:"username"`
	Ttl      int64     `json:"ttl"`
	Uris     []*string `json:"uris"`
}

type WebRTCService struct {
	SignedURI  string
	IceServers []WebRTCIceServer
}

func CreateChannel(ftCtx awsproxy.FTContext, deviceId string) (string, error) {
	videoService := getVideoService(ftCtx)
	createChannelInput := kvs.CreateSignalingChannelInput{ChannelName: &deviceId}
	channelInfo, err := videoService.CreateSignalingChannel(ftCtx.Context, &createChannelInput)
	if nil != err {
		ftCtx.RequestLogger.Debug().Str("device", deviceId).Msg("failed to create channel")
		var ce *types.ResourceInUseException
		if errors.As(err, &ce) {
			ftCtx.RequestLogger.Debug().Str("device", deviceId).Msg("resource in use, listing")
			nc := types.ChannelNameCondition{ComparisonOperator: types.ComparisonOperatorBeginsWith, ComparisonValue: &deviceId}
			lci := kvs.ListSignalingChannelsInput{ChannelNameCondition: &nc}
			lco, err := videoService.ListSignalingChannels(ftCtx.Context, &lci)
			if nil == err && len(lco.ChannelInfoList) == 1 {
				return *lco.ChannelInfoList[0].ChannelARN, nil
			} else if nil != err {
				ftCtx.RequestLogger.Info().Str("device", deviceId).Err(err).Msg("failed to list channels")
			} else {
				ftCtx.RequestLogger.Info().Str("device", deviceId).Int("channels", len(lco.ChannelInfoList)).Err(err).Msg("unexpected channel count")
			}
		}
		return "", err
	}
	return *channelInfo.ChannelARN, nil
}

func GetServices(ftCtx awsproxy.FTContext, channelARN, deviceId string, viewer bool) (WebRTCService, error) {
	services := WebRTCService{}
	videoService := getVideoService(ftCtx)
	var smcCfg types.SingleMasterChannelEndpointConfiguration
	channelProtocol := types.ChannelProtocolHttps
	channelRole := types.ChannelRoleMaster
	if viewer {
		channelRole = types.ChannelRoleViewer
	}
	smcCfg = types.SingleMasterChannelEndpointConfiguration{Protocols: []types.ChannelProtocol{channelProtocol}, Role: channelRole}
	endpointInput := kvs.GetSignalingChannelEndpointInput{ChannelARN: &channelARN, SingleMasterChannelEndpointConfiguration: &smcCfg}
	channelEndpoint, err := videoService.GetSignalingChannelEndpoint(ftCtx.Context, &endpointInput)
	if nil != err {
		return services, err
	}

	iceServers, err := findIceServers(ftCtx, channelARN, deviceId)
	if nil != err {
		return services, err
	}
	channelProtocol = types.ChannelProtocolWss
	smcCfg = types.SingleMasterChannelEndpointConfiguration{Protocols: []types.ChannelProtocol{channelProtocol}, Role: channelRole}
	endpointInput = kvs.GetSignalingChannelEndpointInput{ChannelARN: &channelARN, SingleMasterChannelEndpointConfiguration: &smcCfg}
	channelEndpoint, err = videoService.GetSignalingChannelEndpoint(ftCtx.Context, &endpointInput)
	if nil != err {
		return services, err
	}
	if len(channelEndpoint.ResourceEndpointList) == 0 {
		return services, fmt.Errorf("No endpoints found.")
	}
	signedURI, err := buildSignedRequest(ftCtx, channelARN, deviceId, *channelEndpoint.ResourceEndpointList[0].ResourceEndpoint, viewer)
	return WebRTCService{SignedURI: signedURI, IceServers: iceServers}, err
}

func findIceServers(ftCtx awsproxy.FTContext, channelARN, deviceID string) ([]WebRTCIceServer, error) {
	kvsSession := session.Must(session.NewSession())
	service := kvsc.New(kvsSession)
	st := kvsc.ServiceTurn
	ftCtx.RequestLogger.Debug().Msg("got ice server service")
	isci := kvsc.GetIceServerConfigInput{
		ChannelARN: &channelARN,
		ClientId:   &deviceID,
		Service:    &st}
	iceConfig, err := service.GetIceServerConfig(&isci)
	if err != nil {
		ftCtx.RequestLogger.Error().Err(err).Msg("cfg returned error")
		return make([]WebRTCIceServer, 0), nil
	}
	ftCtx.RequestLogger.Debug().Int("servers", len(iceConfig.IceServerList)).Msg("cfg returned without error")
	iceServerList := make([]WebRTCIceServer, 0)
	for _, server := range iceConfig.IceServerList {
		iceServer := WebRTCIceServer{
			Password: *server.Password,
			Username: *server.Username,
			Ttl:      *server.Ttl,
			Uris:     server.Uris,
		}
		ftCtx.RequestLogger.Debug().Str("username", *server.Username).Int("uris", len(server.Uris)).Msg("found a server")
		iceServerList = append(iceServerList, iceServer)
	}
	ftCtx.RequestLogger.Debug().Int("total servers", len(iceServerList)).Msg("servers result")
	return iceServerList, nil
}

func buildSignedRequest(ftCtx awsproxy.FTContext, channelARN, deviceID, socketURL string, viewer bool) (string, error) {
	encodedARN := url.QueryEscape(channelARN)
	encodedDevice := url.QueryEscape(deviceID)
	var socketQuery string
	if viewer {
		socketQuery = fmt.Sprintf("X-Amz-ChannelARN=%s&X-Amz-ClientId=%s", encodedARN, encodedDevice)
	} else {
		socketQuery = fmt.Sprintf("X-Amz-ChannelARN=%s", encodedARN)
	}
	ftCtx.RequestLogger.Debug().Bool("viewer", viewer).Str("arn", encodedARN).Str("device", encodedDevice).Str("socketQuery", socketQuery).Msg("built URL")
	signer := signer.NewSigner()
	accessToken, secretKey := awsproxy.WebRTCParameters()
	// ftCtx.RequestLogger .Str("t": accessToken, "s": secretKey}).Debug().Msg("creds")
	appCreds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessToken, secretKey, ""))
	creds, err := appCreds.Retrieve(ftCtx.Context)
	if err != nil {
		return "", err
	}
	runes := []rune(socketURL)
	sepIndex := strings.Index(socketURL, ":")
	if sepIndex >= 0 {
		path := string(runes[sepIndex+1:]) + "/"

		ftCtx.RequestLogger.Debug().Str("query", socketQuery).Str("socketUrl", socketURL).Str("path", path).Msg("path and query built")
		request, err := http.NewRequest("GET", socketURL, nil)
		if nil != err {
			return "", err
		}
		request.URL.Opaque = path
		request.URL.RawQuery = socketQuery
		ftCtx.RequestLogger.Debug().Str("query", socketQuery).Str("socketUrl", socketURL).Str("path", path).Msg("request built")
		signedUri, _, err := signer.PresignHTTP(ftCtx.Context, creds, request, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "kinesisvideo", "ca-central-1", time.Now())
		ftCtx.RequestLogger.Debug().Str("signed", signedUri).Msg("signed URL")
		return signedUri, err
	}
	return "", fmt.Errorf("separator not found")
}

func getVideoService(ftCtx awsproxy.FTContext) *kvs.Client {
	cfg, err := config.LoadDefaultConfig(ftCtx.Context)
	if err != nil {
		// handle error
	}
	return kvs.NewFromConfig(cfg)
}
