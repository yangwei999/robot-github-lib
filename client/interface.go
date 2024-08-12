package client

import (
	"fmt"

	sdk "github.com/google/go-github/v36/github"
)

type PRInfo struct {
	Org    string
	Repo   string
	Number int
}

func (p PRInfo) String() string {
	return fmt.Sprintf("%s/%s:%d", p.Org, p.Repo, p.Number)
}

// Client interface for GitHub API
type Client interface {
	AddPRLabel(pr PRInfo, label string) error
	RemovePRLabel(pr PRInfo, label string) error
	CreatePRComment(pr PRInfo, comment string) error
	DeletePRComment(org, repo string, ID int64) error
	GetPRCommits(pr PRInfo) ([]*sdk.RepositoryCommit, error)
	GetPRComments(pr PRInfo) ([]*sdk.IssueComment, error)
	UpdatePR(pr PRInfo, request *sdk.PullRequest) (*sdk.PullRequest, error)
	GetPullRequests(pr PRInfo) ([]*sdk.PullRequest, error)
	ListCollaborator(pr PRInfo) ([]*sdk.User, error)
	IsCollaborator(pr PRInfo, login string) (bool, error)
	RemoveRepoMember(pr PRInfo, login string) error
	AddRepoMember(pr PRInfo, login, permission string) error
	GetPullRequestChanges(pr PRInfo) ([]*sdk.CommitFile, error)
	GetPRLabels(pr PRInfo) ([]string, error)
	GetRepositoryLabels(pr PRInfo) ([]string, error)
	UpdatePRComment(pr PRInfo, commentID int64, ic *sdk.IssueComment) error
	ClosePR(pr PRInfo) error
	ReopenPR(pr PRInfo) error
	AssignPR(pr PRInfo, logins []string) error
	UnAssignPR(pr PRInfo, logins []string) error
	CloseIssue(pr PRInfo) error
	ReopenIssue(pr PRInfo) error
	MergePR(pr PRInfo, commitMessage string, opt *sdk.PullRequestOptions) error
	GetRepos(org string) ([]*sdk.Repository, error)
	GetRepo(org, repo string) (*sdk.Repository, error)
	CreateRepo(org string, r *sdk.Repository) error
	UpdateRepo(org, repo string, r *sdk.Repository) error
	CreateRepoLabel(org, repo, label string) error
	GetRepoLabels(org, repo string) ([]string, error)
	AssignSingleIssue(is PRInfo, login string) error
	UnAssignSingleIssue(is PRInfo, login string) error
	CreateIssueComment(is PRInfo, comment string) error
	UpdateIssueComment(is PRInfo, commentID int64, c *sdk.IssueComment) error
	ListIssueComments(is PRInfo) ([]*sdk.IssueComment, error)
	RemoveIssueLabel(is PRInfo, label string) error
	AddIssueLabel(is PRInfo, label []string) error
	GetIssueLabels(is PRInfo) ([]string, error)
	UpdateIssue(is PRInfo, iss *sdk.IssueRequest) error
	GetSingleIssue(is PRInfo) (*sdk.Issue, error)
	ListBranches(org, repo string) ([]*sdk.Branch, error)
	SetProtectionBranch(org, repo, branch string, pre *sdk.ProtectionRequest) error
	RemoveProtectionBranch(org, repo, branch string) error
	GetDirectoryTree(org, repo, branch string, recursive bool) ([]*sdk.TreeEntry, error)
	GetPathContent(org, repo, path, branch string) (*sdk.RepositoryContent, error)
	CreateFile(org, repo, path, branch, commitMSG, sha string, content []byte) error
	GetUserPermissionOfRepo(org, repo, user string) (*sdk.RepositoryPermissionLevel, error)
	CreateIssue(org, repo string, request *sdk.IssueRequest) (*sdk.Issue, error)
	GetRef(org, repo, ref string) (*sdk.Reference, error)
	CreateBranch(org, repo string, reference *sdk.Reference) error
	ListOperationLogs(pr PRInfo) ([]*sdk.Timeline, error)
	GetEnterprisesMember(org string) ([]*sdk.User, error)
	GetSinglePR(org, repo string, number int) (*sdk.PullRequest, error)
	GetBot() (string, error)
	ListOrg() ([]string, error)
	ListCommits(org, repo, path string) ([]*sdk.RepositoryCommit, error)
}
