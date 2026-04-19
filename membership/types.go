package membership

import (
	"authentik_group_manager/authentik"
	"log/slog"

	"goauthentik.io/api/v3"
)

type GroupUserDependency struct {
	Group *api.Group
	Users []User
}

type DependencyResolver struct {
	client       *authentik.AuthentikAPI
	groupMembers map[string]*GroupUserDependency
	logger       *slog.Logger
}

func NewDependencyResolver(client *authentik.AuthentikAPI, logger *slog.Logger) DependencyResolver {
	dR := DependencyResolver{client: client, groupMembers: make(map[string]*GroupUserDependency), logger: logger}
	groups, err := client.ListGroups()
	if err != nil {
		logger.Error("Error retrieving all groups", "error", err)
		panic(err)
	}
	for _, g := range groups {
		dR.groupMembers[g.Name] = &GroupUserDependency{Group: &g}
	}
	return dR
}

func (dr *DependencyResolver) GetGroupMembers() map[string]*GroupUserDependency {
	return dr.groupMembers
}

func (dr *DependencyResolver) addUserToGroup(group *api.Group, user User) {
	if _, ok := dr.groupMembers[group.Name]; !ok {
		dr.groupMembers[group.Name] = &GroupUserDependency{Group: group}
	}
	dr.groupMembers[group.Name].Users = append(dr.groupMembers[group.Name].Users, user)
}

func (dr *DependencyResolver) ConsumeUser(user api.User) {
	for _, group := range dr.getExplicitGroups(&user) {
		dr.addUserToGroup(group, &user)
		for _, sub := range dr.getSubGroups(group) {
			dr.addUserToGroup(sub, &user)
		}
	}
}

func (dr *DependencyResolver) getSubGroups(group *api.Group) []*api.Group {
	var subGroups []*api.Group
	if val, ok := group.Attributes["subGroups"]; ok {
		sgs, ok2 := val.([]any)
		if ok2 {
			for _, groupString := range sgs {
				if sub, err := dr.client.GetGroupByName(groupString.(string)); err == nil {
					subGroups = append(subGroups, sub)
					subGroups = append(subGroups, dr.getSubGroups(sub)...)
				}
			}
		} else {
			dr.logger.Error("%v: subGroups is not a string array: %v\n", group.Name, val)
		}
	}
	return subGroups
}

func (dr *DependencyResolver) getExplicitGroups(user *api.User) []*api.Group {
	var groups []*api.Group
	groupsStr, ok := user.Attributes["explicitGroups"]
	if !ok {
		dr.logger.Info("User does not have explicitGroups",
			"user", user.Name)
		return groups
	}
	result, ok := groupsStr.([]any)
	if !ok {
		dr.logger.Error("explicitGroups is no List!",
			"user", user.Name)
		return groups
	}
	for _, group := range result {
		groupString, ok := group.(string)
		if !ok {
			dr.logger.Error("explicitGroups contains a non string member", "user", user.Name)
			return groups
		}
		gbn, err := dr.client.GetGroupByName(groupString)
		if err != nil {
			dr.logger.Error("Group does not exist", "group", group)
			continue
		}
		groups = append(groups, gbn)
	}
	return groups
}
