package parser

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	Err           error
}

type Result struct {
	Id  string  `json:"id"`
	Res float64 `json:"res"`
}

type Node struct {
	left  *Node
	right *Node
	value string
}
