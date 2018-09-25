package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Uploader is an interface for uploaders
type Uploader interface {
	Upload(ctx context.Context, path string) error
}

// B2Uploader uploads data to backblaze B2
type B2Uploader struct {
	ApplicationID       string
	ApplicationKey      string
	BucketID            string
	Object              string
	UploadRetryInterval int
	HTTPClient          *http.Client
}

// B2AuthorizeAccountResponse is the b2 authorize account response
type B2AuthorizeAccountResponse struct {
	AuthorizationToken string
	APIUrl             string
}

// B2GetUploadURLResponse is the b2 get upload response
type B2GetUploadURLResponse struct {
	UploadURL          string
	AuthorizationToken string
}

// Upload uploads data to b2
func (b B2Uploader) Upload(ctx context.Context, path string) error {
	return b.upload(ctx, path)
}

func (b B2Uploader) upload(ctx context.Context, path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	idAndKey := fmt.Sprintf("%s:%s", b.ApplicationID, b.ApplicationKey)
	idAndKeyBase64 := base64.StdEncoding.EncodeToString([]byte(idAndKey))
	basicAuthString := fmt.Sprintf("Basic %s", idAndKeyBase64)

	request, _ := http.NewRequest(http.MethodGet, "https://api.backblazeb2.com/b2api/v1/b2_authorize_account", nil)
	request.Header.Set("Authorization", basicAuthString)

	respAuthorize, err := b.HTTPClient.Do(request)
	if err != nil {
		return err
	}

	if respAuthorize.StatusCode != 200 {
		return fmt.Errorf("Authorize account status code: %d", respAuthorize.StatusCode)
	}

	defer respAuthorize.Body.Close()
	body, err := ioutil.ReadAll(respAuthorize.Body)
	if err != nil {
		return err
	}

	authroizeAccountResp := B2AuthorizeAccountResponse{}
	err = json.Unmarshal(body, &authroizeAccountResp)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s/b2api/v1/b2_get_upload_url", authroizeAccountResp.APIUrl)
	getUploadURLBody := map[string]string{"bucketId": b.BucketID}
	bodyBytes, _ := json.Marshal(getUploadURLBody)
	request, _ = http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
	request.Header.Set("Authorization", authroizeAccountResp.AuthorizationToken)

	respGetUpload, err := b.HTTPClient.Do(request)
	if err != nil {
		return err
	}

	if respGetUpload.StatusCode != 200 {
		return fmt.Errorf("Get upload URL status code: %d", respGetUpload.StatusCode)
	}

	defer respGetUpload.Body.Close()
	body, err = ioutil.ReadAll(respGetUpload.Body)
	if err != nil {
		return err
	}

	getUploadURLResponse := B2GetUploadURLResponse{}
	err = json.Unmarshal(body, &getUploadURLResponse)
	if err != nil {
		return err
	}
	sha1OfData := fmt.Sprintf("%x", sha1.Sum(data))

	request, _ = http.NewRequest(http.MethodPost, getUploadURLResponse.UploadURL, bytes.NewBuffer(data))
	request.Header.Set("Authorization", getUploadURLResponse.AuthorizationToken)
	request.Header.Set("X-Bz-File-Name", b.Object)
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("X-Bz-Content-Sha1", sha1OfData)

	respUpload, err := b.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer respUpload.Body.Close()
	body, _ = ioutil.ReadAll(respUpload.Body)

	if respUpload.StatusCode != 200 {
		return fmt.Errorf("Upload file status code: %d", respUpload.StatusCode)
	}

	return nil
}
