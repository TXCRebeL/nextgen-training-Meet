package main

import (
	"Day8/stack"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type MismatchError struct {
	Position int
	Message  string
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf("Mismatch Error at position %d: %s", e.Position, e.Message)
}

type SyntaxError struct {
	Position int
	Message  string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("Syntax Error at position %d: %s", e.Position, e.Message)
}

var DivisionByZeroError = errors.New("division by zero")

// This function will return the error if brackets are not balanced.
func BracketMatcher(expression string) error {
	stack := stack.NewSliceStack[rune]()
	for i, char := range expression {
		switch char {
		case '(', '[', '{', '<':
			stack.Push(char)
		case ')', '}', ']', '>':
			if stack.IsEmpty() {
				return &MismatchError{
					Position: i,
					Message:  fmt.Sprintf("Unexpected closing bracket: %c", char),
				}
			}
			expected, _ := stack.Pop()
			if !isMatching(expected, char) {
				return &MismatchError{
					Position: i,
					Message:  fmt.Sprintf("Mismatched brackets: expected matching for %c, found %c", expected, char),
				}
			}
		}
	}
	if !stack.IsEmpty() {
		return &MismatchError{
			Position: len(expression),
			Message:  "Unclosed opening bracket",
		}
	}
	return nil
}

func isMatching(open, close rune) bool {
	return (open == '(' && close == ')') ||
		(open == '{' && close == '}') ||
		(open == '[' && close == ']') ||
		(open == '<' && close == '>')
}

// InfixToPostfix converts an infix expression to a space-separated postfix string.
func InfixToPostfix(expression string) (string, error) {
	stack := stack.NewSliceStack[string]()
	var postfix string
	i := 0
	lastTokenWasOperator := true // Start with true to catch unary minus at beginning

	for i < len(expression) {
		char := rune(expression[i])
		switch {
		case char == ' ':
			i++
		case char == '(':
			stack.Push("(")
			i++
			lastTokenWasOperator = true
		case char == ')':
			peek, _ := stack.Peek()
			for !stack.IsEmpty() && peek != "(" {
				op, _ := stack.Pop()
				postfix += op + " "
				peek, _ = stack.Peek()
			}
			if stack.IsEmpty() {
				return "", &MismatchError{Position: i, Message: "Missing opening bracket ("}
			}
			stack.Pop() // Pop "("
			i++
			lastTokenWasOperator = false
		case isOperator(char):
			op := string(char)
			// Check for unary minus
			if char == '-' && lastTokenWasOperator {
				op = "u-"
			}

			peek, _ := stack.Peek()
			for !stack.IsEmpty() && peek != "(" && shouldPop(peek, op) {
				pOp, _ := stack.Pop()
				postfix += pOp + " "
				peek, _ = stack.Peek()
			}
			stack.Push(op)
			i++
			lastTokenWasOperator = true
		case isDigit(char) || char == '.':
			numStr := ""
			for i < len(expression) && (isDigit(rune(expression[i])) || expression[i] == '.') {
				numStr += string(expression[i])
				i++
			}
			postfix += numStr + " "
			lastTokenWasOperator = false
		default:
			return "", &SyntaxError{Position: i, Message: fmt.Sprintf("Invalid character: %c", char)}
		}
	}

	for !stack.IsEmpty() {
		op, _ := stack.Pop()
		if op == "(" {
			return "", &MismatchError{Position: len(expression), Message: "Missing closing bracket )"}
		}
		postfix += op + " "
	}
	return strings.TrimSpace(postfix), nil
}

func GetPrecedenceForOp(op string) int {
	switch op {
	case "u-":
		return 4 // Highest precedence
	case "^":
		return 3
	case "*", "/":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}

func shouldPop(stackOp, currentOp string) bool {
	pStack := GetPrecedenceForOp(stackOp)
	pCurr := GetPrecedenceForOp(currentOp)
	if isRightAssociative(currentOp) {
		return pStack > pCurr
	}
	return pStack >= pCurr
}

func isRightAssociative(op string) bool {
	return op == "u-" || op == "^"
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func isOperator(char rune) bool {
	return char == '+' || char == '-' || char == '*' || char == '/' || char == '^'
}

// PostfixEvaluation evaluates a space-separated postfix expression.
func PostfixEvaluation(expression string) (float64, error) {
	if strings.TrimSpace(expression) == "" {
		return 0, nil
	}
	stack := stack.NewSliceStack[float64]()
	tokens := strings.Fields(expression)
	for _, token := range tokens {
		if token == "u-" {
			if stack.Size() < 1 {
				return 0, &SyntaxError{Message: "Invalid expression: missing operand for unary minus"}
			}
			val, _ := stack.Pop()
			stack.Push(-val)
		} else if isOperatorStr(token) {
			if stack.Size() < 2 {
				return 0, &SyntaxError{Message: fmt.Sprintf("Invalid expression: not enough operands for %s", token)}
			}
			op2, _ := stack.Pop()
			op1, _ := stack.Pop()
			switch token {
			case "+":
				stack.Push(op1 + op2)
			case "-":
				stack.Push(op1 - op2)
			case "*":
				stack.Push(op1 * op2)
			case "/":
				if op2 == 0 {
					return 0, DivisionByZeroError
				}
				stack.Push(op1 / op2)
			case "^":
				stack.Push(math.Pow(op1, op2))
			}
		} else {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, &SyntaxError{Message: fmt.Sprintf("Invalid operand: %s", token)}
			}
			stack.Push(num)
		}
	}
	if stack.Size() != 1 {
		if stack.Size() == 0 {
			return 0, nil
		}
		return 0, &SyntaxError{Message: "Invalid expression: malformed postfix (too many operands)"}
	}
	result, _ := stack.Pop()
	return result, nil
}

func isOperatorStr(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/" || s == "^"
}

// This function will calculate the value of the given expression and return all the errors
func Calculate(expression string) (float64, error) {

	// First check if the expression is balanced
	err := BracketMatcher(expression)
	if err != nil {
		return 0, fmt.Errorf("evaluating '%s': %w", expression, err)
	}
	// Then convert the expression to postfix notation
	postfix, err := InfixToPostfix(expression)
	if err != nil {
		return 0, fmt.Errorf("evaluating '%s': %w", expression, err)
	}
	// Then evaluate the postfix expression
	result, err := PostfixEvaluation(postfix)
	if err != nil {
		return 0, fmt.Errorf("evaluating '%s': %w", expression, err)
	}
	return result, nil
}

func main() {
	fmt.Println(Calculate("2 + 3 * 4"))
}
