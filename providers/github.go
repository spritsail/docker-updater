package github

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"strings"
)

func GetReleases(repo, authToken string, usePrereleases bool) (release *github.RepositoryRelease, err error) {
	repoSplit := strings.Split(repo, "/")
	if len(repoSplit) != 2 {
		err = errors.New(fmt.Sprintf("malformed repo string: %s", repo))
		return
	}

	var ghOwner, ghRepo = repoSplit[0], repoSplit[1]

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: authToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	if usePrereleases {
		var releases []*github.RepositoryRelease
		releases, _, err := client.Repositories.ListReleases(ctx, ghOwner, ghRepo, nil)

		if err != nil || len(releases) == 0 {
			err = errors.New(fmt.Sprintf("no releases found for %s: %s", repo, err))
			return nil, err
		}
		release = releases[0]

	} else {
		release, _, err = client.Repositories.GetLatestRelease(ctx, ghOwner, ghRepo)
		if err != nil {
			err = errors.New(fmt.Sprintf("no releases found for %s: %s", repo, err))
			return
		}
	}
	return
}

func GetReleaseTag(repo, authToken string, usePrereleases, stripV bool) (tag string, err error) {
	release, err := GetReleases(repo, authToken, usePrereleases)
	if err != nil {
		return
	}

	if stripV {
		tag = strings.Trim(*release.TagName, "v")
	} else {
		tag = *release.TagName
	}
	return
}
