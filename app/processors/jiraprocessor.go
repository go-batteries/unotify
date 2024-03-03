package processors

import (
	"context"
	"errors"
	"strings"
	"unotify/app/pkg/config"
	"unotify/app/pkg/exmachine"
	"unotify/app/pkg/externals"
	"unotify/app/pkg/helpers"
	"unotify/app/pkg/workerpool"

	"github.com/sirupsen/logrus"
)

// type CommitParser interface {
// 	GetTokens(string) []string
// }

// type GithubJiraReleaseParser struct {
// 	rx *regexp.Regexp
// }

type JiraProcessor struct {
	client       *externals.JiraAPIClient
	statereactor exmachine.StateReactorEngine
	EventChannel chan string
}

const (
	DefaultJiraEventChanSize = 1000
)

func NewJiraProcessor(
	cfg *config.AppConfig,
	eventChanSize int,
	statereactor exmachine.StateReactorEngine,
) (*JiraProcessor, error) {
	logrus.Infoln("atlassian domain ", cfg.AtlassianURL)

	jclient, err := externals.NewJiraAPI(
		cfg.AtlassianEmail,
		cfg.AtlassianAPIKey,
		externals.WithHttpClient(
			externals.NewHTTPClient(),
		),
		externals.WithDomainURI(
			cfg.AtlassianURL,
		),
	)

	if err != nil {
		logrus.WithError(err).Errorf("failed to initialize jira client %+v\n", err)
		return nil, err
	}
	return &JiraProcessor{
		client:       jclient,
		EventChannel: make(chan string, eventChanSize),
		statereactor: statereactor,
	}, nil
}

func (jp *JiraProcessor) ProcessEach(ctx context.Context, issueID string) workerpool.Result {
	logrus.WithContext(ctx).Infoln("processing ticket with issue ID", issueID)

	transitionData, err := jp.client.GetTransitions(ctx, issueID)
	if err != nil {
		logrus.
			WithContext(ctx).
			WithError(err).
			Error("failed to get transition data")

		return workerpool.Result{Err: err}
	}

	logrus.Println("dlksfdslkfjdslkfjdsklfjsdlfk")

	status, err := jp.client.GetIssueStatus(ctx, issueID)
	if err != nil {
		logrus.
			WithContext(ctx).
			WithError(err).
			Error("failed to get issue status")

		return workerpool.Result{Err: err}
	}

	// debugtools.Logdeep(transitionData.Transitions)

	transitionMap := map[string]string{} // this string{} is int{}, but lord jira, decided "int"

	for _, transition := range transitionData.Transitions {
		transitionMap[transition.Name] = transition.ID
	}

	projectID := strings.ToLower(helpers.ToProjectID(issueID))

	logrus.WithContext(ctx).Infoln("status ", status, "state", status.State, "proj", projectID)
	// logrus.Infoln(jp.statereactor.MachineMap)

	statemachine, ok := jp.statereactor.MachineMap[projectID]
	if !ok {
		logrus.WithContext(ctx).Error("couldn't find state machine for ", projectID, issueID)
		return workerpool.Result{Err: err}
	}

	newState, final, err := statemachine.NextState(
		ctx,
		status.State,
		"next",
		exmachine.WithInvertedAlias,
	)
	if err != nil {
		logrus.
			WithContext(ctx).
			WithError(err).Error("failed to get next state")

		return workerpool.Result{Err: err}
	}
	logrus.Println("m,xznjshmfekwfodsijf", final)

	if final {
		logrus.WithContext(ctx).Infoln(issueID, " already in terminal state:", status)
		return workerpool.Result{Err: nil, Data: newState}
	}

	logrus.WithContext(ctx).Infoln("next state", newState.Alias)
	targetTransitionID, ok := transitionMap[newState.Alias]
	if !ok {
		err = errors.New("transition_state_not_defined")
		return workerpool.Result{Err: err}
	}

	err = jp.client.ResolveIssue(ctx, issueID, targetTransitionID)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to resolve issue")
	}

	return workerpool.Result{Err: err, Data: newState}
}
