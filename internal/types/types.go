package types

type Auth struct {
	Password   string `yaml:"password"`
	KeyPath    string `yaml:"keypath"`
	Passphrase string `yaml:"passphrase"`
	Proxy      *Node  `yaml:"proxy"`
}

type User struct {
	Name string `yaml:"name"`
	Auth `yaml:"auth,omitempty"`
}

type Node struct {
	Name     string   `yaml:"name"`
	Alias    []string `yaml:"alias"`
	Hostname string   `yaml:"hostname"`
	IPv4     string   `yaml:"ipv4"`
	Port     string   `yaml:"port"`
	Users    []*User  `yaml:"users"`
}
