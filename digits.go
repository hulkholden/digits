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
	// Val is the value of the expression.
	Val int
	// Op is the expression operation.
	// If it's opNone the expression represents a constant with value Val.
	Op   operation
	AExp *expression
	BExp *expression
}

func makeConstant(v int) expression {
	return expression{Val: v}
}

func makeAdd(v int, a, b expression) expression {
	if r := a.Val + b.Val; r != v {
		panic(fmt.Sprintf("%s + %s = %d, want %d", a, b, r, v))
	}
	return expression{Val: v, Op: opAdd, AExp: &a, BExp: &b}
}

func makeSubtract(v int, a, b expression) expression {
	if r := a.Val - b.Val; r != v {
		panic(fmt.Sprintf("%s - %s = %d, want %d", a, b, r, v))
	}
	return expression{Val: v, Op: opSubtract, AExp: &a, BExp: &b}
}

func makeMultiply(v int, a, b expression) expression {
	if r := a.Val * b.Val; r != v {
		panic(fmt.Sprintf("%s * %s = %d, want %d", a, b, r, v))
	}
	return expression{Val: v, Op: opMultiply, AExp: &a, BExp: &b}
}

func makeDivide(v int, a, b expression) expression {
	if b.Val == 0 {
		panic("denominator is zero")
	}
	if r := a.Val / b.Val; r != v {
		panic(fmt.Sprintf("%s / %s = %d, want %d", a, b, r, v))
	}
	return expression{Val: v, Op: opDivide, AExp: &a, BExp: &b}
}

func (e expression) String() string {
	if e.Op == opNone {
		return fmt.Sprintf("%d", e.Val)
	}
	return fmt.Sprintf("(%s %s %s)", e.AExp.String(), e.Op.String(), e.BExp.String())
}

func (e expression) eval() (int, bool) {
	if e.Op == opNone {
		return e.Val, true
	}
	a, aOk := e.AExp.eval()
	b, bOk := e.BExp.eval()
	if aOk && bOk {
		return e.Op.eval(a, b)
	}
	return 0, false
}

// canonicalize ensures commutative operations are always expressed consistently (lowest operand first).
func (e expression) canonicalize() expression {
	if e.Op == opAdd || e.Op == opMultiply {
		if e.BExp.Val < e.AExp.Val {
			e.AExp, e.BExp = e.BExp, e.AExp
		}
	}

	// TODO: this should also consider associativity (e.g. (a + (b+c)) == ((a+b) + c))

	return e
}

func solve(target int, digits []int) []expression {
	var solutions []expression

	// Cache this outside the loop to reduce thrashing.
	var other []int

	// See if there is a valid solution of the form 'a op otherDigits' or 'otherDigits op a'.
	for aIdx := 0; aIdx < len(digits); aIdx++ {
		// Identity.
		a := digits[aIdx]
		aExp := makeConstant(a)
		if a == target {
			solutions = append(solutions, aExp)
		}

		other = other[:0]
		other = append(other, digits[:aIdx]...)
		other = append(other, digits[aIdx+1:]...)

		// Addition.
		if target > a {
			for _, soln := range solve(target-a, other) {
				solutions = append(solutions, makeAdd(target, aExp, soln))
			}
		}

		// Subtraction.
		if a > target {
			for _, soln := range solve(a-target, other) {
				solutions = append(solutions, makeSubtract(target, aExp, soln))
			}
		}
		for _, soln := range solve(target+a, other) {
			solutions = append(solutions, makeSubtract(target, soln, aExp))
		}

		// Multiplication.
		if (target % a) == 0 {
			for _, soln := range solve(target/a, other) {
				solutions = append(solutions, makeMultiply(target, aExp, soln))
			}
		}

		// Division.
		if (a % target) == 0 {
			for _, soln := range solve(a/target, other) {
				solutions = append(solutions, makeDivide(target, aExp, soln))
			}
		}
		for _, soln := range solve(target*a, other) {
			solutions = append(solutions, makeDivide(target, soln, aExp))
		}
	}

	// TODO: divide digits into two sets. For each solution in set A, see if there is a solution in set B which will form the target.

	// Normalize remove duplicates.
	seen := make(map[string]bool)
	var solsOut []expression
	for i := range solutions {
		s := solutions[i].canonicalize()
		key := s.String()
		if !seen[key] {
			seen[key] = true
			solsOut = append(solsOut, s)
		}
	}
	return solsOut
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

	solns := solve(*target, digits)
	if len(solns) == 0 {
		fmt.Printf("no solution found :(\n")
		return
	}

	shortest := ""

	for i, soln := range solns {
		result, ok := soln.eval()
		if !ok {
			fmt.Fprintf(os.Stderr, "result is invalid\n")
			os.Exit(1)
		}
		if result != *target {
			fmt.Fprintf(os.Stderr, "generated incorrect solution: %s = %d, != %d!\n", soln.String(), result, *target)
			os.Exit(2)
		}
		str := soln.String()
		if shortest == "" || len(str) < len(shortest) {
			shortest = str
		}
		fmt.Printf("%d: %d = %s\n", i, result, str)
	}
	fmt.Printf("Shortest solution: %s\n", shortest)
}
