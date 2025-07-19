package membership

import (
	"authentik_group_manager/authentik"
	"log/slog"
	"slices"

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
	for _, group := range groups {
		dR.groupMembers[group.Name] = &GroupUserDependency{Group: &group}
	}
	return dR
}

func (dr *DependencyResolver) GetGroupMembers() map[string]*GroupUserDependency {
	return dr.groupMembers
}

func (dr *DependencyResolver) ConsumeUser(user api.User) {
	var userGroups []*api.Group
	for _, group := range dr.getExplicitGroups(&user) {
		subGroups := dr.getSubGroups(group)
		if _, ok := dr.groupMembers[group.Name]; !ok {
			dr.groupMembers[group.Name] = &GroupUserDependency{Group: group}
		}

		dr.groupMembers[group.Name].Users = slices.Concat(dr.groupMembers[group.Name].Users, []User{&user})
		if len(subGroups) > 0 {
			userGroups = slices.Concat(userGroups, subGroups, []*api.Group{group})
			for _, subSubGroup := range subGroups {
				if _, ok := dr.groupMembers[subSubGroup.Name]; !ok {
					dr.groupMembers[subSubGroup.Name] = &GroupUserDependency{Group: group}
				}
				dr.groupMembers[subSubGroup.Name].Users = slices.Concat(dr.groupMembers[subSubGroup.Name].Users, []User{&user})
			}
		}
	}
}

func (dr *DependencyResolver) getSubGroups(group *api.Group) []*api.Group {
	var subGroups []*api.Group
	if val, ok := group.Attributes["subGroups"]; ok {
		sgs, ok2 := val.([]any)
		if ok2 {
			for _, groupString := range sgs {
				if group, err := dr.client.GetGroupByName(groupString.(string)); err == nil {
					subGroups = slices.Concat(subGroups, []*api.Group{group})
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
		groups = slices.Concat(groups, []*api.Group{gbn})
	}
	return groups
}
