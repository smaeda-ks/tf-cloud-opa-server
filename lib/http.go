package lib

import (
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

func NewRetryHttpClient() *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.Logger = nil
	return retryClient.StandardClient()
}
