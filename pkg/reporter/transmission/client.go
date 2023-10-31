package transmission

import (
	"fmt"
	"net/url"

	"github.com/hekmon/transmissionrpc/v3"

	"github.com/trim21/errgo"
)

func newClient(entryPoint string) (*transmissionrpc.Client, error) {
	u, err := url.Parse(entryPoint)
	if err != nil {
		return nil, errgo.Wrap(err, fmt.Sprintf("TRANSMISSION_API_ENTRYPOINT '%s' is not valid url", entryPoint))
	}

	client, err := transmissionrpc.New(u, nil)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to create transmission client")
	}

	return client, nil
}
