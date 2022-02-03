package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	invitationList := make(map[string]struct{})

	invitations, _, err := client.Organizations.ListPendingOrgInvitations(ctx, "neotoolkit", &github.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	for _, invitation := range invitations {
		if invitation.Login != nil {
			invitationList[*invitation.Login] = struct{}{}
		}
	}

	repos, _, err := client.Repositories.ListByOrg(ctx, "neotoolkit", &github.RepositoryListByOrgOptions{
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	for _, repo := range repos {
		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()

		pulls, _, err := client.PullRequests.List(ctx, owner, repoName, &github.PullRequestListOptions{
			State:     "closed",
			Sort:      "updated",
			Direction: "desc",
		})
		if err != nil {
			log.Fatalln(err)
		}

		within := 5 * 24 * time.Hour

		for _, pull := range pulls {
			if pull.MergedAt == nil {
				continue
			}

			if closedAgo := time.Since(pull.GetClosedAt()); closedAgo > within {
				continue
			}

			if pull.GetUser().GetType() == "Bot" {
				continue
			}

			switch pull.GetAuthorAssociation() {
			case "OWNER", "MEMBER", "COLLABORATOR":
				continue
			}

			userName := pull.GetUser().GetLogin()

			if _, ok := invitationList[userName]; ok {
				continue
			}

			_, _, err := client.Teams.AddTeamMembershipBySlug(ctx, "neotoolkit", "team", userName, &github.TeamAddTeamMembershipOptions{})
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
