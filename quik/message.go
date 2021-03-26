package quik

type Message struct {
	Cmd   string      `json:"cmd"`
	Data  interface{} `json:"data"`
	Time  string      `json:"t"`
	Error string      `json:"lua_error"`
}
