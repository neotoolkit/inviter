package main

import (
	"context"
	"log"
	"os"

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

	_, _, err = client.Teams.AddTeamMembershipBySlug(ctx, "neotoolkit", "team", "neotoolkit-bot", &github.TeamAddTeamMembershipOptions{})
	if err != nil {
		log.Fatalln(err)
	}
}
