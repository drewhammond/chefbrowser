package chef

import "github.com/go-chef/chef"

// this package is the interface between the chef api and our application

type Repository struct {
	ServiceInterface
}

type Service struct {
	Repository
}

func (s Service) GetNodes() (interface{}, error) {
	nodes := chef.Node{Name: "foo"}
	return nodes, nil
}

type ServiceInterface interface {
	GetNodes() ([]chef.Node, error)
	GetRoles() ([]chef.Role, error)
	GetCookbooks() (interface{}, error)
	GetEnvironments() error
	GetNode(id string) (interface{}, error)
	CheckHealth() error
}

type ConcreteImplementation struct {
	ServiceInterface
}

type Client struct {
	*Repository
}

func New() *Service {
	s := &Service{}
	return s
}
