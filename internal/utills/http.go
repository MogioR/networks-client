package utills

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

func Post(client *fasthttp.Client, url string, body []byte, headers map[string]string) ([]byte, error) {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.Header.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetBody(body)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	err := client.Do(req, resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("unexpected status code = %d on url = %s", resp.StatusCode(), url)
	}

	if len(resp.Body()) == 0 {
		return nil, fmt.Errorf("empty response body on url = %s", url)
	}

	bodyCopy := append([]byte(nil), resp.Body()...)
	return bodyCopy, nil
}
