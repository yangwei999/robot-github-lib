package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/google/go-github/v36/github"
	"golang.org/x/oauth2"
)

func NewClient(getToken func() []byte) Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: string(getToken()),
	})
	tc := oauth2.NewClient(context.Background(), ts)

	return client{sdk.NewClient(tc)}
}

type client struct {
	c *sdk.Client
}

func (cl client) AddPRLabel(pr PRInfo, label string) error {
	_, _, err := cl.c.Issues.AddLabelsToIssue(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, []string{label},
	)

	return err
}

func (cl client) RemovePRLabel(pr PRInfo, label string) error {
	r, err := cl.c.Issues.RemoveLabelForIssue(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, label,
	)
	if err != nil && r != nil && r.StatusCode == 404 {
		return nil
	}

	return err
}

func (cl client) CreatePRComment(pr PRInfo, comment string) error {
	ic := sdk.IssueComment{
		Body: sdk.String(comment),
	}
	_, _, err := cl.c.Issues.CreateComment(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, &ic,
	)

	return err
}

func (cl client) DeletePRComment(org, repo string, commentId int64) error {
	_, err := cl.c.Issues.DeleteComment(context.Background(), org, repo, commentId)

	return err
}

func (cl client) GetPRComments(pr PRInfo) ([]*sdk.IssueComment, error) {
	comments := []*sdk.IssueComment{}

	opt := &sdk.IssueListCommentsOptions{}
	opt.Page = 1

	for {
		v, resp, err := cl.c.Issues.ListComments(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
		if err != nil {
			return comments, err
		}

		comments = append(comments, v...)

		link := parseLinks(resp.Header.Get("Link"))["next"]
		if link == "" {
			break
		}

		pagePath, err := url.Parse(link)
		if err != nil {
			break
		}

		p := pagePath.Query().Get("page")
		if p == "" {
			break
		}

		page, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		opt.Page = page
	}

	return comments, nil
}

func (cl client) GetPRCommits(pr PRInfo) ([]*sdk.RepositoryCommit, error) {
	commits := []*sdk.RepositoryCommit{}

	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.ListCommits(context.Background(), pr.Org, pr.Repo, pr.Number, nil)
			if err != nil {
				return err
			}

			commits = append(commits, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return commits, err
}

func (cl client) UpdatePR(pr PRInfo, request *sdk.PullRequest) (*sdk.PullRequest, error) {

	pull, _, err := cl.c.PullRequests.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, request)
	if err != nil {
		return nil, err
	}

	return pull, nil
}

func (cl client) GetPullRequests(pr PRInfo) ([]*sdk.PullRequest, error) {
	var prs []*sdk.PullRequest
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.List(context.Background(), pr.Org, pr.Repo,
				&sdk.PullRequestListOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			prs = append(prs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return prs, err
}

func (cl client) ListCollaborator(pr PRInfo) ([]*sdk.User, error) {
	var collaborator []*sdk.User

	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Repositories.ListCollaborators(context.Background(), pr.Org, pr.Repo,
				&sdk.ListCollaboratorsOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			collaborator = append(collaborator, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return collaborator, err
}

func (cl client) IsCollaborator(pr PRInfo, login string) (bool, error) {
	b, _, err := cl.c.Repositories.IsCollaborator(context.Background(), pr.Org, pr.Repo, login)
	if err != nil {
		return false, err
	}
	return b, nil
}

func (cl client) RemoveRepoMember(pr PRInfo, login string) error {
	_, err := cl.c.Repositories.RemoveCollaborator(context.Background(), pr.Org, pr.Repo, login)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) AddRepoMember(pr PRInfo, login, permission string) error {
	_, _, err := cl.c.Repositories.AddCollaborator(context.Background(), pr.Org, pr.Repo, login,
		&sdk.RepositoryAddCollaboratorOptions{Permission: permission})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetPullRequestChanges(pr PRInfo) ([]*sdk.CommitFile, error) {
	var files []*sdk.CommitFile

	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.ListFiles(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
			if err != nil {
				return err
			}

			files = append(files, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return files, err
}

func (cl client) GetPRLabels(pr PRInfo) ([]string, error) {
	pull, _, err := cl.c.PullRequests.Get(context.Background(), pr.Org, pr.Repo, pr.Number)
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(pull.Labels))
	for _, p := range pull.Labels {
		labels = append(labels, *p.Name)
	}

	return labels, nil
}

func (cl client) GetRepositoryLabels(pr PRInfo) ([]string, error) {
	var rLabels []*sdk.Label
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListLabels(context.Background(), pr.Org, pr.Repo, opt)
			if err != nil {
				return err
			}

			rLabels = append(rLabels, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}
	labels := make([]string, len(rLabels))
	for _, r := range rLabels {
		labels = append(labels, *r.Name)
	}

	return labels, nil
}

func (cl client) UpdatePRComment(pr PRInfo, commentID int64, ic *sdk.IssueComment) error {
	_, _, err := cl.c.Issues.EditComment(context.Background(), pr.Org, pr.Repo, commentID, ic)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ClosePR(pr PRInfo) error {
	action := ActionClosed
	_, _, err := cl.c.PullRequests.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &sdk.PullRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ReopenPR(pr PRInfo) error {
	action := "open"
	_, _, err := cl.c.PullRequests.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &sdk.PullRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) AssignPR(pr PRInfo, logins []string) error {
	_, _, err := cl.c.Issues.AddAssignees(context.Background(), pr.Org, pr.Repo, pr.Number, logins)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UnAssignPR(pr PRInfo, logins []string) error {
	_, _, err := cl.c.Issues.RemoveAssignees(context.Background(), pr.Org, pr.Repo, pr.Number, logins)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) CloseIssue(pr PRInfo) error {
	action := ActionClosed
	_, _, err := cl.c.Issues.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &sdk.IssueRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ReopenIssue(pr PRInfo) error {
	action := "open"
	_, _, err := cl.c.Issues.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &sdk.IssueRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) MergePR(pr PRInfo, commitMessage string, opt *sdk.PullRequestOptions) error {
	_, _, err := cl.c.PullRequests.Merge(context.Background(), pr.Org, pr.Repo, pr.Number, commitMessage, opt)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetRepos(org string) ([]*sdk.Repository, error) {
	var rps []*sdk.Repository
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1
		opt.PerPage = 100

		for {
			v, resp, err := cl.c.Repositories.ListByOrg(context.Background(), org, &sdk.RepositoryListByOrgOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			rps = append(rps, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return rps, err
}

func (cl client) GetRepo(org, repo string) (*sdk.Repository, error) {
	r, _, err := cl.c.Repositories.Get(context.Background(), org, repo)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (cl client) CreateRepo(org string, r *sdk.Repository) error {
	_, _, err := cl.c.Repositories.Create(context.Background(), org, r)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UpdateRepo(org, repo string, r *sdk.Repository) error {
	_, _, err := cl.c.Repositories.Edit(context.Background(), org, repo, r)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) CreateRepoLabel(org, repo, label string) error {
	_, _, err := cl.c.Issues.CreateLabel(context.Background(), org, repo, &sdk.Label{Name: &label})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetRepoLabels(org, repo string) ([]string, error) {
	var lbs []*sdk.Label
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListLabels(context.Background(), org, repo, opt)
			if err != nil {
				return err
			}

			lbs = append(lbs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(lbs))
	for _, l := range lbs {
		labels = append(labels, *l.Name)
	}

	return labels, nil
}

func (cl client) AssignSingleIssue(is PRInfo, login string) error {
	_, _, err := cl.c.Issues.AddAssignees(context.Background(), is.Org, is.Repo, is.Number, []string{login})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UnAssignSingleIssue(is PRInfo, login string) error {
	_, _, err := cl.c.Issues.RemoveAssignees(context.Background(), is.Org, is.Repo, is.Number, []string{login})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) CreateIssueComment(is PRInfo, comment string) error {
	ic := sdk.IssueComment{
		Body: sdk.String(comment),
	}
	_, _, err := cl.c.Issues.CreateComment(context.Background(), is.Org, is.Repo, is.Number, &ic)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UpdateIssueComment(is PRInfo, commentID int64, c *sdk.IssueComment) error {
	_, _, err := cl.c.Issues.EditComment(context.Background(), is.Org, is.Repo, commentID, c)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ListIssueComments(is PRInfo) ([]*sdk.IssueComment, error) {
	var comments []*sdk.IssueComment

	opt := &sdk.IssueListCommentsOptions{}
	opt.Page = 1

	for {
		v, resp, err := cl.c.Issues.ListComments(context.Background(), is.Org, is.Repo, is.Number, opt)
		if err != nil {
			return comments, err
		}

		comments = append(comments, v...)

		link := parseLinks(resp.Header.Get("Link"))["next"]
		if link == "" {
			break
		}

		pagePath, err := url.Parse(link)
		if err != nil {
			break
		}

		p := pagePath.Query().Get("page")
		if p == "" {
			break
		}

		page, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		opt.Page = page
	}

	return comments, nil
}

func (cl client) RemoveIssueLabel(is PRInfo, label string) error {
	_, err := cl.c.Issues.RemoveLabelForIssue(context.Background(), is.Org, is.Repo, is.Number, label)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) AddIssueLabel(is PRInfo, label []string) error {
	_, _, err := cl.c.Issues.AddLabelsToIssue(context.Background(), is.Org, is.Repo, is.Number, label)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetIssueLabels(is PRInfo) ([]string, error) {
	var lbs []*sdk.Label
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListLabelsByIssue(context.Background(), is.Org, is.Repo, is.Number, opt)
			if err != nil {
				return err
			}

			lbs = append(lbs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(lbs))
	for _, l := range lbs {
		labels = append(labels, *l.Name)
	}

	return labels, nil
}

func (cl client) UpdateIssue(is PRInfo, iss *sdk.IssueRequest) error {
	_, _, err := cl.c.Issues.Edit(context.Background(), is.Org, is.Repo, is.Number, iss)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetSingleIssue(is PRInfo) (*sdk.Issue, error) {
	issue, _, err := cl.c.Issues.Get(context.Background(), is.Org, is.Repo, is.Number)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

func (cl client) ListBranches(org, repo string) ([]*sdk.Branch, error) {
	var brs []*sdk.Branch
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Repositories.ListBranches(context.Background(), org, repo,
				&sdk.BranchListOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			brs = append(brs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(brs))
	for _, b := range brs {
		labels = append(labels, *b.Name)
	}

	return brs, nil
}

func (cl client) SetProtectionBranch(org, repo, branch string, pre *sdk.ProtectionRequest) error {
	_, _, err := cl.c.Repositories.UpdateBranchProtection(context.Background(), org, repo, branch, pre)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) RemoveProtectionBranch(org, repo, branch string) error {
	_, err := cl.c.Repositories.RemoveBranchProtection(context.Background(), org, repo, branch)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetDirectoryTree(org, repo, branch string, recursive bool) ([]*sdk.TreeEntry, error) {
	trees, _, err := cl.c.Git.GetTree(context.Background(), org, repo, branch, recursive)
	if err != nil {
		return nil, err
	}

	return trees.Entries, nil
}

func (cl client) GetPathContent(org, repo, path, branch string) (*sdk.RepositoryContent, error) {
	fc, _, _, err := cl.c.Repositories.GetContents(context.Background(), org, repo, path,
		&sdk.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return nil, err
	}

	return fc, nil
}

func (cl client) CreateFile(org, repo, path, branch, commitMSG, sha string, content []byte) error {
	_, _, err := cl.c.Repositories.CreateFile(context.Background(), org, repo, path,
		&sdk.RepositoryContentFileOptions{Content: content, Message: &commitMSG, Branch: &branch, SHA: &sha})

	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetUserPermissionOfRepo(org, repo, user string) (*sdk.RepositoryPermissionLevel, error) {
	permission, _, err := cl.c.Repositories.GetPermissionLevel(context.Background(), org, repo, user)
	if err != nil {
		return nil, err
	}

	return permission, nil
}

func (cl client) CreateIssue(org, repo string, request *sdk.IssueRequest) (*sdk.Issue, error) {
	is, _, err := cl.c.Issues.Create(context.Background(), org, repo, request)
	if err != nil {
		return nil, err
	}

	return is, nil
}

func (cl client) GetRef(org, repo, ref string) (*sdk.Reference, error) {
	r, _, err := cl.c.Git.GetRef(context.Background(), org, repo, ref)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (cl client) CreateBranch(org, repo string, reference *sdk.Reference) error {
	_, _, err := cl.c.Git.CreateRef(context.Background(), org, repo, reference)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ListOperationLogs(pr PRInfo) ([]*sdk.Timeline, error) {
	var t []*sdk.Timeline
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListIssueTimeline(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
			if err != nil {
				return err
			}

			t = append(t, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (cl client) GetEnterprisesMember(org string) ([]*sdk.User, error) {
	var t []*sdk.User
	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Organizations.ListMembers(context.Background(), org,
				&sdk.ListMembersOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			t = append(t, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (cl client) GetSinglePR(org, repo string, number int) (*sdk.PullRequest, error) {
	p, _, err := cl.c.PullRequests.Get(context.Background(), org, repo, number)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (cl client) GetBot() (string, error) {
	u, _, err := cl.c.Users.Get(context.Background(), "")
	if err != nil {
		return "", err
	}

	return u.GetLogin(), err
}

func (cl client) ListOrg() ([]string, error) {
	var r []string

	opt := sdk.ListOptions{PerPage: 99, Page: 1}
	for {
		ls, _, err := cl.c.Organizations.List(context.Background(), "", &opt)
		if err != nil {
			return nil, err
		}

		if len(ls) == 0 {
			break
		}

		for _, v := range ls {
			r = append(r, v.GetLogin())
		}

		opt.Page += 1
	}

	return r, nil
}
