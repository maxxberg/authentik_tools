package main

import (
	"authentik_group_manager/authentik"
	"fmt"
	"goauthentik.io/api/v3"
	"os"
	"slices"
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

func main() {
	apiToken := os.Getenv("AGM_API_TOKEN")
	appConfig := authentik.NewConfig(apiToken)
	clientCfg := api.NewConfiguration()
	clientCfg.Host = os.Getenv("AGM_API_HOST")
	clientCfg.Scheme = "https"
	aClient := api.NewAPIClient(clientCfg)

	client := authentik.NewAuthentikAPI(aClient, appConfig)

	for _, user := range client.ListUsers() {
		err := updateUser(user, client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	groups, err := client.ListGroups()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, group := range groups {
		fmt.Println(group.Name)
		for _, user := range group.UsersObj {
			fmt.Printf("\t%v\n", user.Name)
		}
	}
}
