package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

// DoRequest
// Takes the url and data plus method and headers of a http request
// And Does the request using http client
// The input and output format of data, is a buffer
// Method returns the http status code plus in addition
func DoRequest(ctx context.Context, logger *logrus.Logger, method string, addr *url.URL,
	postData []byte, headers []http.Header) ([]byte, int, error) {

	var (
		err        error
		body       io.Reader    = nil // default value when http method doesn't require body (GET , OPTIONS,...)
		reader     *gzip.Reader = nil
		respBuffer []byte
	)

	// Client  for do the request
	client := http.DefaultClient

	// Creating io.Reader
	if postData != nil {
		body = bytes.NewBuffer(postData)
	}

	// Creating request
	request, err := http.NewRequest(method, addr.String(), body)
	if err != nil {
		logger.
			WithField("request_url", addr).
			WithField("err", err).
			Errorln("error creating http request")

		return nil, http.StatusInternalServerError, err
	}

	// Add context
	request = request.WithContext(ctx)

	// Add headers
	addHeaders(request, headers)
	start := time.Now()
	logger.WithField("request", addr.String()).Infoln("started")

	// Do request
	resp, err := client.Do(request)
	if err != nil {
		logger.
			WithField("request_url", addr).
			WithField("err", err).
			Errorln("error calling web service")

		return nil, http.StatusInternalServerError, err
	}
	/*if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, gerrors.WrapError(models.ErrConnectionError, models.ErrInternalServerError)
	}*/

	// Decompressing the response body , because of gZipped responses of web service :((
	// Discovered after hard working for a couple of hours
	// Anyway, all of responses aren't compressed , so it should be checked before performing
	if !resp.Uncompressed {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			logger.
				WithField("request_url", addr).
				WithField("err", err).
				Errorln("error decompressing response of web service")

			return nil, http.StatusInternalServerError, err
		}
	}

	defer resp.Body.Close()

	// Read response
	// if response body be compressed , ioutils.Reader will read the gzip reader
	// Else the original response body of http will be read
	if reader != nil {
		respBuffer, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	} else {
		respBuffer, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	// Log the response body in successful case
	defer func() {
		logger.
			WithField("Request", addr).
			WithField("Status", resp.StatusCode).
			WithField("Has error", err != nil).
			Infoln("took time(ms):", time.Since(start).Milliseconds())
	}()

	return respBuffer, resp.StatusCode, nil
}

// adds the provided headers to the request
func addHeaders(request *http.Request, headers []http.Header) {

	for _, header := range headers {
		for k, v := range header {
			for _, h := range v {
				request.Header.Set(k, h)
			}
		}
	}
}
