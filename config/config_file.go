package config

type ServerConfig struct {
	Version int      `json."version"`
	Port    int      `json:"port"`
	Address string   `json:"address"`
	Servers []Server `json:"servers"`
}

type Server struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}
