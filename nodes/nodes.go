package nodes

import "net/url"

type Nodes struct {
	Addresses map[string]string
}

func New() *Nodes {
	return &Nodes{make(map[string]string)}
}

func (n *Nodes) Register(nodeUrl, address string) error {
	u, err := url.Parse(nodeUrl)
	if err != nil {
		return err
	}
	n.Addresses[address] = u.Host
	return nil
}
