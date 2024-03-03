package externals

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unotify/app/pkg/debugtools"

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

const (
	JSONContentType = "application/json"
)

const (
	TransitionsURL = "%s/rest/api/2/issue/%s/transitions"
	IssuesURL      = "%s/rest/api/2/issue/%s"
)

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

	client             *http.Client
	computedBasicToken string
	allowInsecure      bool

	cacher ResponseCacher[TransitionResponseData]
}

func NewJiraAPI(email string, apiKey string, opts ...JiraApiOpts) (client *JiraAPIClient, err error) {
	cacher := NewAppCacher[TransitionResponseData]("jira::transitions", DefaultResponseCacheDuration)

	client = &JiraAPIClient{
		Email:         email,
		APIKey:        apiKey,
		allowInsecure: false,
		cacher:        cacher,
	}

	for _, opt := range opts {
		opt(client)
	}

	validate := validator.New()
	if err = validate.Struct(client); err != nil {
		logrus.WithError(err).Error("failed to build jira client, required fields missing")
		return nil, err
	}

	uri, err := url.Parse(client.DomainURI)
	if err != nil {
		logrus.WithError(err).Error("invalid domain url provided. should be tenant.atlassian.net")
		return nil, err
	}

	if uri.Scheme == "" {
		uri.Scheme = "https"
	}
	// if client.allowInsecure {
	// 	uri.Scheme = "http"
	// }

	client.DomainURI = strings.TrimSuffix(uri.String(), "/")

	// logrus.Println("mnopns domain url", client.DomainURI, uri.Scheme, uri.String())

	client.computedBasicToken = base64.
		StdEncoding.
		EncodeToString([]byte(email + ":" + apiKey))

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

type TransitionState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TransitionResponseData struct {
	Expand      string            `json:"expand"`
	Transitions []TransitionState `json:"transitions"`
}

func (api *JiraAPIClient) GetTransitions(
	ctx context.Context,
	issueID string,
) (
	*TransitionResponseData,
	error,
) {
	destinationURL := fmt.Sprintf(TransitionsURL, api.DomainURI, issueID)

	// Get projectID from issueID

	projectID := issueID

	re := regexp.MustCompile(`^([\w\d]+)-\d{1,}`)
	matches := re.FindStringSubmatch(issueID)
	if len(matches) == 2 {
		projectID = matches[1]
	}

	logrus.WithContext(ctx).Infoln("getting project transitions for", projectID)

	cachedData, ok := api.cacher.Get(ctx, projectID)
	if ok {
		logrus.
			WithContext(ctx).
			Infoln("returning project transitions from cache", projectID)

		return &cachedData, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", destinationURL, nil)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to build request")
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+api.computedBasicToken)
	req.Header.Add("Content-Type", JSONContentType)

	// debugtools.HttpRequestLog(req)

	res, err := api.client.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get response from jira")
		return nil, err
	}

	defer res.Body.Close()

	data := &TransitionResponseData{}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to read api response")
		return nil, err
	}

	logrus.WithContext(ctx).Debugln("response data", string(b))

	// decoder := json.NewDecoder(res.Body)
	// if err := decoder.Decode(data); err != nil {
	// 	logrus.WithContext(ctx).WithError(err).Error("failed to decode transitions")
	// 	return nil, err
	// }
	if err := json.Unmarshal(b, data); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to decode transitions")
		return nil, err
	}

	logrus.WithContext(ctx).Debugln("transitions data", *data)

	api.cacher.Set(
		ctx,
		projectID,
		*data,
		DefaultResponseCacheDuration,
	)

	return data, nil
}

type IssueTransitionUpdateRequest struct {
	Transition map[string]string `json:"transition"`
}

type IssuesResponse struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Status IssueStatus `json:"status"`
}

type IssueStatus struct {
	State   string `json:"name"`
	IssueID string `json:"-"`
}

func (api *JiraAPIClient) GetIssueStatus(
	ctx context.Context,
	issueID string,
) (
	*IssueStatus,
	error,
) {
	destinationURL := fmt.Sprintf(IssuesURL, api.DomainURI, issueID)

	req, err := http.NewRequestWithContext(ctx, "GET", destinationURL, nil)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to build request ", destinationURL)
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+api.computedBasicToken)
	req.Header.Add("Content-Type", JSONContentType)

	res, err := api.client.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get response from jira")
		return nil, err
	}

	if res.StatusCode > http.StatusBadRequest {
		return nil, errors.New("jira_api_failed")
	}

	defer res.Body.Close()

	data := &IssuesResponse{
		Fields: IssueFields{
			Status: IssueStatus{IssueID: issueID},
		},
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to read api response")
		return nil, err
	}

	logrus.WithContext(ctx).Debugln("response data", string(b))

	// decoder := json.NewDecoder(res.Body)
	// if err := decoder.Decode(data); err != nil {
	// 	logrus.WithContext(ctx).WithError(err).Error("failed to decode transitions")
	// 	return nil, err
	// }

	if err := json.Unmarshal(b, data); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to decode transitions")
		return nil, err
	}

	logrus.WithContext(ctx).Debugln("issues data data", *data)

	return &data.Fields.Status, nil
}

func (api *JiraAPIClient) ResolveIssue(
	ctx context.Context,
	issueID string,
	targetTransitionID string,
) error {
	destinationURL := fmt.Sprintf(TransitionsURL, api.DomainURI, issueID)

	transitionRequest := fmt.Sprintf(`{"transition": { "id": "%s" }}`, targetTransitionID)
	reader := strings.NewReader(transitionRequest)

	req, err := http.NewRequestWithContext(ctx, "POST", destinationURL, reader)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to build request")
		return err
	}

	req.Header.Add("Authorization", "Basic "+api.computedBasicToken)
	req.Header.Add("Content-Type", JSONContentType)

	debugtools.HttpRequestLog(req)

	res, err := api.client.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get response from jira")
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		return errors.New("fuck_your_request")
	}

	logrus.WithContext(ctx).Infoln("transition success")

	return nil
}
