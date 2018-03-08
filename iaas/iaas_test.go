package iaas

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/sacloud/open-service-broker-sacloud/version"
)

var testClient *client

func TestMain(m *testing.M) {

	//環境変数にトークン/シークレットがある場合のみテスト実施
	accessToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

	if accessToken == "" || accessTokenSecret == "" {
		log.Warn("Please Set ENV 'SAKURACLOUD_ACCESS_TOKEN' and 'SAKURACLOUD_ACCESS_TOKEN_SECRET'")
		os.Exit(0) // exit normal
	}
	log.SetOutput(ioutil.Discard)

	zone := os.Getenv("SAKURACLOUD_ZONE")
	if zone == "" {
		zone = "is1b"
	}

	acceptLanguage := os.Getenv("SAKURACLOUD_ACCEPT_LANGUAGE")

	retryMax := 0
	strRetryMax := os.Getenv("SAKURACLOUD_RETRY_MAX")
	if strRetryMax != "" {
		retryMax, _ = strconv.Atoi(strRetryMax)
	}

	retryInterval := int64(0)
	strInterval := os.Getenv("SAKURACLOUD_RETRY_INTERVAL")
	if strInterval != "" {
		retryInterval, _ = strconv.ParseInt(strInterval, 10, 64)
	}

	apiRootURL := os.Getenv("USACLOUD_API_ROOT_URL")

	traceMode := false
	if os.Getenv("SAKURACLOUD_TRACE_MODE") != "" {
		traceMode = true
	}

	testClient = NewClient(&ClientConfig{
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
		Zone:              zone,
		AcceptLanguage:    acceptLanguage,
		RetryMax:          retryMax,
		RetryIntervalSec:  retryInterval,
		APIRootURL:        apiRootURL,
		TraceMode:         traceMode,
	}).(*client)

	testClient.rawClient.UserAgent = fmt.Sprintf("oepn-service-broker-sacloud-test/v%s", version.Version)

	ret := m.Run()
	os.Exit(ret)
}
