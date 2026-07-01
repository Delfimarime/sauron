package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

// ArtifactClient lists the artifacts of one registry kind and downloads their
// content archives.
type ArtifactClient interface {
	// List returns one page of artifact summaries.
	List(ctx context.Context, opts ...ListOption) (*ArtifactList, error)
	// Content downloads the named artifact's file-tree archive. version is the
	// Artifact-Version response header; it is empty when the registry declares
	// none.
	Content(ctx context.Context, name string) (archive []byte, version string, err error)
}

// artifactClient is the resty-backed ArtifactClient for a single kind.
type artifactClient struct {
	rest *resty.Client
	kind kind
}

// kind identifies an artifact collection and its registry path segment.
type kind string

const (
	kindSkills kind = "skills"
	kindAgents kind = "agents"
)

// List issues a GET against the kind's collection and decodes the page.
func (c *artifactClient) List(ctx context.Context, opts ...ListOption) (*ArtifactList, error) {
	options := ListOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	resp, err := c.rest.R().
		SetContext(ctx).
		SetQueryParams(queryFrom(options)).
		Get("/" + string(c.kind))
	if err != nil {
		return nil, fmt.Errorf("%w: list %s: %w", ErrTransport, c.kind, err)
	}

	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, apiErrorFrom(resp)
	}

	var list ArtifactList
	if err := json.Unmarshal(resp.Body(), &list); err != nil {
		return nil, fmt.Errorf("%w: decode %s listing: %w", ErrTransport, c.kind, err)
	}

	return &list, nil
}

// Content issues GET /{kind}/{name}/content, returning the raw archive bytes and
// the Artifact-Version header value. version is empty when the header is absent.
func (c *artifactClient) Content(ctx context.Context, name string) ([]byte, string, error) {
	resp, err := c.rest.R().
		SetContext(ctx).
		Get("/" + string(c.kind) + "/" + name + "/content")
	if err != nil {
		return nil, "", fmt.Errorf("%w: content %s/%s: %w", ErrTransport, c.kind, name, err)
	}

	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, "", apiErrorFrom(resp)
	}

	version := resp.Header().Get("Artifact-Version")
	return resp.Body(), version, nil
}

// queryFrom maps listing options to query parameters, mapping Search to the
// registry's free-text q parameter.
func queryFrom(options ListOptions) map[string]string {
	query := map[string]string{}
	if options.Search != nil {
		query["q"] = *options.Search
	}
	if options.Sort != nil {
		query["sort"] = *options.Sort
	}
	if options.Limit != nil {
		query["limit"] = strconv.FormatInt(*options.Limit, 10)
	}
	if options.Offset != nil {
		query["offset"] = strconv.FormatInt(*options.Offset, 10)
	}
	return query
}

// problem is the application/problem+json body the registry returns on error.
type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// apiErrorFrom builds an APIError from a non-2xx response, decoding the problem
// detail when the body carries one.
func apiErrorFrom(resp *resty.Response) *APIError {
	apiErr := &APIError{Status: resp.StatusCode()}

	var body problem
	if json.Unmarshal(resp.Body(), &body) == nil {
		apiErr.Type = body.Type
		apiErr.Title = body.Title
		apiErr.Detail = body.Detail
	}

	return apiErr
}
