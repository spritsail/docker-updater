package dockerfile_updater

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"strings"
)

func trimV(release github.RepositoryRelease) (version string) {
	return strings.Trim(*release.TagName, "v")
}

func GetReleases(repo, authToken string, usePrereleases bool) (release string, err error) {

	var ctx = context.Background()

	release = ""

	repoSplit := strings.Split(repo, "/")

	if len(repoSplit) != 2 {
		err = errors.New(fmt.Sprintf("malformed repo string: %s", repo))
		return
	}

	var ghOwner = repoSplit[0]
	var ghRepo = repoSplit[1]

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: authToken})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	var latestRelease *github.RepositoryRelease

	if usePrereleases {
		var repositoryReleases []*github.RepositoryRelease
		repositoryReleases, _, err = client.Repositories.ListReleases(ctx, ghOwner, ghRepo, nil)

		if err != nil {
			return
		}

		if len(repositoryReleases) == 0 {
			err = errors.New(fmt.Sprintf("no releases found for %s", repo))
			return
		}

		latestRelease = repositoryReleases[0]

	} else {

		latestRelease, _, err = client.Repositories.GetLatestRelease(ctx, ghOwner, ghRepo)

		if err != nil {
			return
		}
	}

	return trimV(*latestRelease), nil

}
