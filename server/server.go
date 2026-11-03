package server

import (
	c "chat/consts"
	m "chat/message"
	"fmt"
	"iter"
	"maps"
	"net"
	"slices"
	"strings"
)

type Server struct {
	Ip         string
	Port       string
	Protocol   string
	Addr       *net.UDPAddr
	Clients    map[string]*net.UDPAddr
	Groups     []*Group
	GroupSeq   uint16
	Connection *net.UDPConn
}

func NewServer(ip string, port string) *Server {
	server := Server{
		Ip:       ip,
		Port:     port,
		Protocol: "udp",
		Clients:  make(map[string]*net.UDPAddr),
		Groups:   []*Group{},
	}

	server.Addr, _ = net.ResolveUDPAddr(server.Protocol, server.address())

	return &server
}

func (s Server) address() string {
	return s.Ip + ":" + s.Port
}

func (s *Server) GetGroup(id uint16) (*Group, bool) {
	if idx := slices.IndexFunc(s.Groups, func(g *Group) bool {
		return g.Id == id
	}); idx != -1 {
		return s.Groups[idx], true
	}

	return nil, false
}

func (s *Server) GetGroupIndex(name string) (int, bool) {
	idx := slices.IndexFunc(s.Groups, func(g *Group) bool {
		return strings.EqualFold(g.Name, name)
	})
	if idx != -1 {
		return idx, true
	}

	return idx, false
}

func (s *Server) EnterGroup(id uint16, user *net.UDPAddr) (*Group, error) {
	group, found := s.GetGroup(id)
	if !found {
		return nil, fmt.Errorf("grupo de id %d não existe", id)
	}

	err := group.AddMember(user)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Server) ExitGroup(id uint16, user *net.UDPAddr) (*Group, error) {
	group, found := s.GetGroup(id)
	if !found {
		return nil, fmt.Errorf("grupo de nome %d não existe", id)
	}

	err := group.RemoveMember(user)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Server) NewGroup(name string) (*Group, error) {
	found := slices.ContainsFunc(s.Groups, func(g *Group) bool {
		return strings.EqualFold(g.Name, name)
	})
	if found {
		return nil, fmt.Errorf("grupo com nome %s já existe", name)
	}

	s.GroupSeq = s.GroupSeq + 1
	group := NewGroup(s.GroupSeq, name)

	s.Groups = append(s.Groups, group)
	return group, nil
}

func handleError(payload string, reply *m.Message) {
	reply.Type = m.Error
	// payload :=
	reply.Payload = []byte(payload)
}

func (s *Server) CreateConnection() error {
	con, err := net.ListenUDP(s.Protocol, s.Addr)
	if err != nil {
		return err
	}

	s.Connection = con
	return nil
}

// criar struct request para encapsular  message e addr
func (s *Server) Read() (*m.Message, *net.UDPAddr, error) {
	bytes := make([]byte, c.MAX_MESSAGE_SIZE)
	n, remoteAddr, err := s.Connection.ReadFromUDP(bytes)
	if err != nil {
		return nil, remoteAddr, err
	}

	fmt.Printf("Requisição de: %s\n", remoteAddr)

	msg, err := m.Decode(n, bytes)
	if err != nil {
		return nil, remoteAddr, fmt.Errorf("[ERRO] Requisição mal-formatada: %s", string(bytes[:n]))
	}

	return msg, remoteAddr, nil
}

func (s *Server) Write(msg *m.Message, addr *net.UDPAddr) error {
	_, err := s.Connection.WriteToUDP(msg.Encode(), addr)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Listen() {
	err := s.CreateConnection()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Connection.Close()

	for {
		msg, remoteAddr, err := s.Read()
		if err != nil {
			fmt.Println(err)
			continue
		}

		s.handleMessage(msg, remoteAddr)
	}
}

func (s *Server) handleMessage(msg *m.Message, addr *net.UDPAddr) {
	reply, _ := m.NewMessage(m.Reply, 0, []byte("mensagem recebida"))
	socket := addr.String()

	if _, exists := s.Clients[socket]; !exists {
		s.Clients[socket] = addr

		old := reply.Payload

		reply.Payload = []byte("bem vindo ao chat")
		s.Write(reply, addr)

		reply.Payload = old
	}

	switch msg.Type {
	case m.Send:
		err := s.handleSend(msg, reply, addr)
		if err != nil {
			handleError(err.Error(), reply)
		}

	case m.CreateGroup:
		err := s.handleCreateGroup(msg, reply)
		if err != nil {
			handleError(err.Error(), reply)
		}

	case m.EnterGroup:
		err := s.handleEnterGroup(msg, reply, addr)
		if err != nil {
			handleError(err.Error(), reply)
		}

	case m.ExitGroup:
		err := s.handleExitGroup(msg, reply, addr)
		if err != nil {
			handleError(err.Error(), reply)
		}

	default:
		handleError(fmt.Sprintf("usuário enviou mensagem de tipo %q", msg.Type), reply)
	}

	s.Connection.WriteToUDP(reply.Encode(), addr)
}

func (s *Server) handleSend(msg *m.Message, reply *m.Message, addr *net.UDPAddr) error {
	if msg.Group == 0 {
		s.handleForwarding(msg, maps.All(s.Clients), addr)

		return nil
	}

	group, found := s.GetGroup(msg.Group)
	if !found {
		return fmt.Errorf("grupo de id %d não existe", msg.Group)
	}

	s.handleForwarding(msg, slices.All(group.Members), addr)

	reply.Payload = []byte("mensagem enviada")
	return nil
}

func (s *Server) handleForwarding[K any](msg *m.Message, clients iter.Seq2[K, *net.UDPAddr], sender *net.UDPAddr) {
	for _, client := range clients {
		if client.String() != sender.String() {
			s.Write(msg, client)
		}
	}
}

func (s *Server) handleCreateGroup(msg *m.Message, reply *m.Message) error {
	group, err := s.NewGroup(string(msg.Payload))
	if err != nil {
		return err
	}

	payload := fmt.Sprintf("grupo de nome %s criado com sucesso!", msg.Payload)
	reply.Payload = []byte(payload)
	reply.Group = group.Id
	return nil
}

func (s *Server) handleEnterGroup(msg *m.Message, reply *m.Message, addr *net.UDPAddr) error {
	group, err := s.EnterGroup(msg.Group, addr)
	if err != nil {
		return err
	}

	payload := fmt.Sprintf("entrou no grupo de id %d e nome %s", group.Id, group.Name)
	reply.Payload = []byte(payload)
	reply.Group = group.Id
	return nil
}

func (s *Server) handleExitGroup(msg *m.Message, reply *m.Message, addr *net.UDPAddr) error {
	group, err := s.ExitGroup(msg.Group, addr)
	if err != nil {
		return err
	}

	payload := fmt.Sprintf("saiu do grupo de id %d e nome %s", group.Id, group.Name)
	reply.Payload = []byte(payload)
	reply.Group = group.Id
	return nil
}
