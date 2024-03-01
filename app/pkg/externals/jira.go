package externals

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// TODO: Adds application/json at Transport layer in RoundTrip()
func NewHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     2 * time.Minute,
		},

		Timeout: 5 * time.Second,
	}

	return client
}

func BasicAuth(id, secret string) string {
	return base64.StdEncoding.EncodeToString([]byte(id + ":" + secret))
}

// ProjectID
// States
// Defining valid state transitions, for each project
type JiraProjectConfig struct {
	ProjectID string `yaml:"project_id" json:"project_id"`
}

type JiraAPIClient struct {
	Email     string `validate:"required"`
	APIKey    string `validate:"required"`
	DomainURI string `validate:"required"`

	client             *http.Client `validate:"required"`
	computedBasicToken string       `validate:"-"`
	allowInsecure      bool         `validate:"-"`
}

func NewJiraAPI(email string, apiKey string, opts ...JiraApiOpts) (client *JiraAPIClient, err error) {
	client = &JiraAPIClient{Email: email, APIKey: apiKey, allowInsecure: false}

	for _, opt := range opts {
		opt(client)
	}

	validate := validator.New()
	if err = validate.Struct(&client); err != nil {
		logrus.WithError(err).Error("failed to build jira client, required fields missing")
		return nil, err
	}

	uri, err := url.Parse(client.DomainURI)
	if err != nil {
		logrus.WithError(err).Error("invalid domain url provided. should be tenant.atlassian.net")
		return nil, err
	}

	client.computedBasicToken = base64.StdEncoding.EncodeToString([]byte(email + ":" + apiKey))

	if uri.Scheme != "" {
		return client, nil
	}

	uri.Scheme = "https"
	if client.allowInsecure {
		uri.Scheme = "http"
	}

	client.DomainURI = uri.String()

	return client, nil
}

type JiraApiOpts func(*JiraAPIClient)

func WithHttpClient(client *http.Client) JiraApiOpts {
	return func(ja *JiraAPIClient) {
		ja.client = client
	}
}

func WithDomainURI(domainURI string) JiraApiOpts {
	return func(ja *JiraAPIClient) {
		ja.DomainURI = domainURI
	}
}

func WithAllowInsecure(flag bool) JiraApiOpts {
	return func(ja *JiraAPIClient) {
		ja.allowInsecure = true
	}
}

// Get the present issue state
// Validate the state transition
// Call api to transition state

const (
	JSONContentType = "application/json"
)

const (
	TransitionsPath = "/rest/api/2/issue/%s/transitions"
	IssuesPath      = "/rest/api/2/issue/%s"
)

type TransitionState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TransitionResponseData struct {
	Expand      string            `json:"expand"`
	Transitions []TransitionState `json:"transitions"`
}

func (api *JiraAPIClient) GetTransitions(ctx context.Context, issueID string) (*TransitionResponseData, error) {
	destinationURL := filepath.Join(api.DomainURI, fmt.Sprintf(TransitionsPath, issueID))

	req, err := http.NewRequestWithContext(ctx, "GET", destinationURL, nil)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to build request")
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+api.computedBasicToken)
	req.Header.Add("Content-Type", JSONContentType)

	res, err := api.client.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get response from jira")
		return nil, err
	}

	defer res.Body.Close()

	data := &TransitionResponseData{}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(data); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to decode transitions")
		return nil, err
	}

	logrus.WithContext(ctx).Debugln("transitions data", *data)

	return data, nil
}

type IssueTransitionUpdateRequest struct {
	Transition map[string]string `json:"transition"`
}

func (api *JiraAPIClient) ResolveIssue(ctx context.Context, issueID string) error {
	destinationURL := filepath.Join(api.DomainURI, fmt.Sprintf(TransitionsPath, issueID))

	req, err := http.NewRequestWithContext(ctx, "POST", destinationURL, nil)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to build request")
		return err
	}

	_ = req
	return nil
}
