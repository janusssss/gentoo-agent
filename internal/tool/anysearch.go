package tool

import (
	"context"
	"net/http"
	"time"

	gmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zendev-sh/goai/mcp"
)

type Anysearch struct {
	client
}

func (a *Anysearch) initialize() error {
	a.name = "anysearch"
	a.Description = "Unified real-time search engine skill for AI agents." +
		" Supports general web search, vertical domain search, parallel batch search, and full-page content extraction."
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	transport := mcp.NewHTTPTransport("https://api.anysearch.com/mcp")
	a.Client = mcp.NewClient(a.name, "", mcp.WithTransport(transport))
	if err := a.Client.Connect(ctx); err != nil {
		return err
	}

	return nil
}

func (a *Anysearch) Connect() (*gmcp.ClientSession, error) {
	ctx := context.Background()
	client := gmcp.NewClient(&gmcp.Implementation{Name: "anysearch", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, &gmcp.StreamableClientTransport{
		Endpoint: "https://api.anysearch.com/mcp",
		HTTPClient: &http.Client{
			Transport: withAuth("Bearer as_sk_2e1b15372dc97a0ce0b406fed796603a"),
		},
	}, nil)
	if err != nil {
		return nil, err
	}
	return session, err
}

type authRT struct{ token string }

func withAuth(token string) http.RoundTripper { return &authRT{token: token} }

func (a *authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", a.token)

	return http.DefaultTransport.RoundTrip(r)

}
