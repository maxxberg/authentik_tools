package main

import (
	"authentik_group_manager/authentik"
	"authentik_group_manager/membership"
	"log/slog"
	"os"
	"slices"

	"goauthentik.io/api/v3"
)

func diffUsers(a, b []membership.User) (pks []int32, names []string) {
	for _, user := range a {
		if user.GetEmail() == "" {
			continue
		}
		if !slices.ContainsFunc(b, func(other membership.User) bool {
			return user.GetUid() == other.GetUid()
		}) {
			pks = append(pks, user.GetPk())
			names = append(names, user.GetName())
		}
	}
	return
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Debug("Logger initiated")
	apiToken := os.Getenv("AGM_API_TOKEN")
	if len(apiToken) == 0 {
		logger.Error("No API key provided")
		os.Exit(1)
	}
	authentikHost := os.Getenv("AGM_API_HOST")
	if len(authentikHost) == 0 {
		logger.Error("No Authentik Host provided")
		os.Exit(1)
	}
	appConfig := authentik.NewConfig(apiToken)
	clientCfg := api.NewConfiguration()
	clientCfg.Host = authentikHost
	clientCfg.Scheme = "https"
	aClient := api.NewAPIClient(clientCfg)

	client := authentik.NewAuthentikAPI(aClient, appConfig)

	depR := membership.NewDependencyResolver(client, logger)

	for _, user := range client.ListUsers() {
		depR.ConsumeUser(user)
	}

	for gName, gMembers := range depR.GetGroupMembers() {
		if gName == "authentik Admins" {
			continue
		}

		actual := make([]membership.User, len(gMembers.Group.UsersObj))
		for i := range gMembers.Group.UsersObj {
			actual[i] = &gMembers.Group.UsersObj[i]
		}

		removePKs, removeNames := diffUsers(actual, gMembers.Users)
		if err := client.RemoveUsersFromGroup(gMembers.Group, removePKs); err != nil {
			logger.Error("Error removing users from group", "error", err)
		}

		addPKs, addNames := diffUsers(gMembers.Users, actual)
		if err := client.AddUsersToGroup(gMembers.Group, addPKs); err != nil {
			logger.Error("Error adding users to group", "error", err)
		}

		var memberNames []string
		for _, m := range gMembers.Users {
			memberNames = append(memberNames, m.GetName())
		}
		logger.Info("Group dependency", "group", gName, "members", memberNames, "negativeDiff", removeNames, "positiveDiff", addNames)
	}
}
