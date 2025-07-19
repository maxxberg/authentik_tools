package main

import (
	"authentik_group_manager/authentik"
	"authentik_group_manager/membership"
	"log/slog"
	"os"
	"slices"

	"goauthentik.io/api/v3"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func updateUser(user api.User, client *authentik.AuthentikAPI) error {
	for _, group := range user.GroupsObj {
		if val, ok := group.Attributes["subGroups"]; ok {
			for _, subGroup := range val.([]interface{}) {
				if !slices.Contains(user.Groups, subGroup.(string)) {
					err := client.AddUsersToGroupByGroupString(subGroup.(string), []int32{user.Pk})
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type UserComparable interface {
	GetUid() string
}

func compareUserFunc[T membership.User](a T) func(b T) bool {
	return func(b T) bool {
		return a.GetUid() == b.GetUid()
	}
}

func diffUserLists[T membership.User](a, b []T) []T {
	var diff []T
	for _, user := range a {
		if user.GetEmail() == "" {
			continue
		}
		if !slices.ContainsFunc(b, compareUserFunc(user)) {
			diff = append(diff, user)
		}
	}
	return diff
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
		var members []string
		for _, member := range gMembers.Users {
			members = append(members, member.GetName())
		}
		users := make([]membership.User, len(gMembers.Group.UsersObj))
		for i, user := range gMembers.Group.UsersObj {
			users[i] = &user
		}
		negativeDiff := diffUserLists(users, gMembers.Users)
		negativePKs := make([]int32, len(negativeDiff))
		for i, user := range negativeDiff {
			negativePKs[i] = user.GetPk()
		}
		err := client.RemoveUsersFromGroupByGroupString(gName, negativePKs)
		if err != nil {
			logger.Error("Error removing users from group", "error", err)
		}
		negativeDiffNames := make([]string, len(negativeDiff))
		for i, user := range negativeDiff {
			negativeDiffNames[i] = user.GetName()
		}
		positiveDiff := diffUserLists(gMembers.Users, users)
		positivePKs := make([]int32, len(positiveDiff))
		for i, user := range positiveDiff {
			positivePKs[i] = user.GetPk()
		}
		err = client.AddUsersToGroupByGroupString(gName, positivePKs)
		if err != nil {
			logger.Error("Error adding users to group", "error", err)
		}
		positiveDiffNames := make([]string, len(positiveDiff))
		for i, user := range positiveDiff {
			positiveDiffNames[i] = user.GetName()
		}
		logger.Info("Group dependency", "group", gName, "members", members, "negativeDiff", negativeDiffNames, "positiveDiff", positiveDiffNames)
	}

	// for _, user := range client.ListUsers() {
	// 	err := updateUser(user, client)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}
	// }

	// groups, err := client.ListGroups()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// for _, group := range groups {
	// 	fmt.Println(group.Name)
	// 	for _, user := range group.UsersObj {
	// 		fmt.Printf("\t%v\n", user.Name)
	// 	}
	// }
}
