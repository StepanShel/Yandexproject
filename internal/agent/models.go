package agent

type AgTask struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Response struct {
	Id  string  `json:"id"`
	Res float64 `json:"res"`
}

type Agent struct {
	compPower int
	port      int
}
