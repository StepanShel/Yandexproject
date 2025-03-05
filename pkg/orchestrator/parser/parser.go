package parser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/StepanShel/YandexProject/pkg/orchestrator/config"
	uuid "github.com/google/uuid"
)

func toPostfix(tokens []string) ([]string, error) {
	var result []string
	var stack []string
	var precedence = map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	for _, token := range tokens {
		if isNumber(token) {
			result = append(result, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				result = append(result, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, fmt.Errorf("Mismatshed parathes")
			}
			stack = stack[:len(stack)-1]
		} else if isOperator(token) {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] && isOperator(stack[len(stack)-1]) {
				result = append(result, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			return nil, fmt.Errorf("unsudnsud")
		}
	}
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" || stack[len(stack)-1] == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		result = append(result, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return result, nil
}

func Ast(tokens []string) (*Node, error) {
	postfix, err := toPostfix(tokens)
	if err != nil {
		return nil, err
	}
	stack := []*Node{}

	for _, token := range postfix {
		if isOperator(token) {
			if len(stack) < 2 {
				return nil, fmt.Errorf("ошибка в постфикснрй записи хд")
			}
			left := stack[len(stack)-2]
			right := stack[len(stack)-1]
			stack = stack[:len(stack)-2]

			stack = append(stack, &Node{
				left:  left,
				right: right,
				value: token,
			})
		} else {
			stack = append(stack, &Node{
				left:  nil,
				right: nil,
				value: token,
			})
		}
	}
	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid exp")
	}
	return stack[0], nil
}

func ParsingAST(node *Node, cfg *config.Config, tasksch chan Task, resultchan chan Result) (float64, error) {
	var task = Task{}
	var divByZeroeErr error
	operationTime := map[string]int{
		"+": cfg.AddTime,
		"-": cfg.Subtime,
		"/": cfg.Divtime,
		"*": cfg.MultiplicTime,
	}

	if isNumber(node.value) {
		res, err := strconv.ParseFloat(node.value, 64)
		if err != nil {
			return 0, err
		}
		return res, nil
	}

	if node.left == nil || node.right == nil {
		return 0, errors.New("invalid AST")
	}

	leftresult, err := ParsingAST(node.left, cfg, tasksch, resultchan)
	if err != nil {
		return 0, err
	}
	rightresult, err := ParsingAST(node.right, cfg, tasksch, resultchan)
	if err != nil {
		return 0, err
	}
	if rightresult == 0 && node.value == "/" {
		divByZeroeErr = errors.New("DIV BY ZERO")
	} else {
		divByZeroeErr = nil
	}
	task = Task{
		ID:            uuid.NewString(),
		Arg1:          leftresult,
		Arg2:          rightresult,
		Operation:     node.value,
		OperationTime: operationTime[node.value],
		Err:           divByZeroeErr,
	}

	tasksch <- task

	var result = Result{}
	for result = range resultchan {
		if result.Id == task.ID {
			node.value = fmt.Sprintf("%v", result.Res)
			return result.Res, nil
		}
	}
	return 0, errors.New("how")

}
