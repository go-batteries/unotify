package processors

import (
	"context"
	"errors"
	"unotify/app/pkg/config"
	"unotify/app/pkg/debugtools"
	"unotify/app/pkg/exmachine"
	"unotify/app/pkg/externals"
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
	statemachine exmachine.StateMachine
	EventChannel chan string
}

const (
	DefaultJiraEventChanSize = 1000
)

func NewJiraProcessor(cfg *config.AppConfig, eventChanSize int) (*JiraProcessor, error) {
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
	status, err := jp.client.GetIssueStatus(ctx, issueID)
	if err != nil {
		logrus.
			WithContext(ctx).
			WithError(err).
			Error("failed to get issue status")

		return workerpool.Result{Err: err}
	}

	debugtools.Logdeep(transitionData.Transitions)

	transitionMap := map[string]string{} // this string{} is int{}, but lord jira, decided "int"

	for _, transition := range transitionData.Transitions {
		transitionMap[transition.Name] = transition.ID
	}

	logrus.WithContext(ctx).Infoln("status ", status, "state", status.State)

	newState, final, err := jp.statemachine.NextState(
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

	targetTransitionID, ok := transitionMap[newState.Alias]
	if !ok {
		err = errors.New("transition_state_not_defined")
		return workerpool.Result{Err: err}
	}

	logrus.WithContext(ctx).Infoln("next state", newState.Alias)
	if final {
		logrus.WithContext(ctx).Infoln(issueID, " already in terminal state:", status)
		return workerpool.Result{Err: nil, Data: newState}
	}

	err = jp.client.ResolveIssue(ctx, issueID, targetTransitionID)
	return workerpool.Result{Err: err}
}
