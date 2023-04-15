package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	digitsStr = flag.String("digits", "", "A comma-separated list of digits")
	target    = flag.Int("target", 0, "The target value to solve for")
)

func parseDigits(s string) ([]int, error) {
	parts := strings.Split(s, ",")

	r := make([]int, len(parts))
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		r[i] = v
	}
	return r, nil
}

type operation int

const (
	opNone operation = iota
	opAdd
	opSubtract
	opMultiply
	opDivide
)

var opStrings = map[operation]string{
	opAdd:      "+",
	opSubtract: "-",
	opMultiply: "*",
	opDivide:   "/",
}

func (op operation) eval(a, b int) (int, bool) {
	switch op {
	case opAdd:
		return a + b, true
	case opSubtract:
		// Subtract is only valid for positive results.
		return a - b, a > b
	case opMultiply:
		return a * b, true
	case opDivide:
		if b == 0 {
			return 0, false
		}
		// Divide is only valid for exact results.
		return a / b, (a % b) == 0
	}

	return 0, false
}

func (op operation) String() string {
	if s, ok := opStrings[op]; ok {
		return s
	}
	return "?"
}

type expression struct {
	Op   operation
	AExp *expression
	BExp *expression
	// If Op==opNone this is a constant value.
	Val int
}

func makeConstant(v int) expression {
	return expression{Val: v}
}

func makeAdd(a, b *expression) expression {
	return expression{Op: opAdd, AExp: a, BExp: b}
}

func makeSubtract(a, b *expression) expression {
	return expression{Op: opSubtract, AExp: a, BExp: b}
}

func makeMultiply(a, b *expression) expression {
	return expression{Op: opMultiply, AExp: a, BExp: b}
}

func makeDivide(a, b *expression) expression {
	return expression{Op: opDivide, AExp: a, BExp: b}
}

func (e expression) String() string {
	if e.Op == opNone {
		return fmt.Sprintf("%d", e.Val)
	}
	return fmt.Sprintf("(%s %s %s)", e.AExp.String(), e.Op.String(), e.BExp.String())
}

func (e expression) value() (int, bool) {
	if e.Op == opNone {
		return e.Val, true
	}
	a, aOk := e.AExp.value()
	b, bOk := e.BExp.value()
	if aOk && bOk {
		return e.Op.eval(a, b)
	}
	return 0, false
}

func solve(target int, digits []int) (expression, bool) {
	if len(digits) == 0 {
		return expression{}, false
	}

	other := make([]int, 0, len(digits))

	// See if there is a valid solution of the form 'a op otherDigits' or 'otherDigits op a'.
	for aIdx := 0; aIdx < len(digits); aIdx++ {
		a := digits[aIdx]
		aExp := makeConstant(a)
		if a == target {
			return aExp, true
		}

		other = other[:0]
		other = append(other, digits[:aIdx]...)
		other = append(other, digits[aIdx+1:]...)

		// opAdd
		if target > a {
			if e, ok := solve(target-a, other); ok {
				return makeAdd(&aExp, &e), true
			}
		}

		// opSubtract
		if a > target {
			if e, ok := solve(a-target, other); ok {
				return makeSubtract(&aExp, &e), true
			}
		}
		if e, ok := solve(target+a, other); ok {
			return makeSubtract(&e, &aExp), true
		}

		// opMultiply
		if (target % a) == 0 {
			if e, ok := solve(target/a, other); ok {
				return makeMultiply(&aExp, &e), true
			}
		}

		// opDivide
		if (a % target) == 0 {
			if e, ok := solve(a/target, other); ok {
				return makeDivide(&aExp, &e), true
			}
		}
		if e, ok := solve(target*a, other); ok {
			return makeDivide(&e, &aExp), true
		}
	}

	// TODO: divide digits into two sets, recurse.

	return expression{}, false
}

func main() {
	flag.Parse()

	if *digitsStr == "" {
		fmt.Fprintf(os.Stderr, "--digits must be provided")
		os.Exit(1)
	}
	digits, err := parseDigits(*digitsStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "--digits invalid: %v", err)
		os.Exit(1)
	}

	if *target == 0 {
		fmt.Fprintf(os.Stderr, "--target must be provided")
		os.Exit(1)
	}

	e, ok := solve(*target, digits)
	if !ok {
		fmt.Printf("no solution found :(\n")
		return
	}
	result, ok := e.value()
	if !ok {
		fmt.Fprintf(os.Stderr, "result is invalid\n")
		os.Exit(1)
	}
	fmt.Printf("%s = %d\n", e.String(), result)
}