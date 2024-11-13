package httpclient

// make sure we conform to HTTPRunner
var _ HTTPRunner = &Client{}

type HTTPRunner interface{}

type Client struct{}

func NewHTTPRunner() HTTPRunner {
	return &Client{}
}
