package server

import (
	"errors"
	"net"
	"slices"
)

type Group struct {
	Id      uint16
	Name    string
	Members []*net.UDPAddr
}

func (g Group) GetMemberIndex(member *net.UDPAddr) int {
	idx := slices.IndexFunc(g.Members, func(e *net.UDPAddr) bool {
		if e.String() == member.String() {
			return true
		}
		return false
	})

	return idx
}

func (g *Group) AddMember(newMember *net.UDPAddr) error {
	idx := g.GetMemberIndex(newMember)
	if idx != -1 {
		return errors.New("membro já está no grupo")
	}

	g.Members = append(g.Members, newMember)
	return nil
}

func (g *Group) RemoveMember(newMember *net.UDPAddr) error {
	idx := g.GetMemberIndex(newMember)
	if idx == -1 {
		return errors.New("membro não está no grupo")
	}

	g.Members = slices.Delete(g.Members, idx, idx+1)
	return nil
}

func NewGroup(id uint16, name string) *Group {
	return &Group{
		Name:    name,
		Members: []*net.UDPAddr{},
		Id:      id,
	}
}
