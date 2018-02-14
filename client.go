package speedtest

import (
	"context"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context/ctxhttp"
)

type Client http.Client

type response http.Response

func (client *Client) get(ctx context.Context, url string) (resp *response, err error) {
	htResp, err := ctxhttp.Get(ctx, (*http.Client)(client), url)
	return (*response)(htResp), err
}

func (client *Client) post(ctx context.Context, url string, bodyType string, body io.Reader) (resp *response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", bodyType)
	htResp, err := ctxhttp.Do(ctx, (*http.Client)(client), req)

	return (*response)(htResp), err
}

func (resp *response) ReadContent() ([]byte, error) {
	content, err := ioutil.ReadAll(resp.Body)
	cerr := resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return content, cerr
	}
	return content, nil
}

func (resp *response) ReadXML(out interface{}) error {
	content, err := resp.ReadContent()
	if err != nil {
		return err
	}
	return xml.Unmarshal(content, out)
}