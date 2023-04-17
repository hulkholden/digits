package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
)

var (
	digitsStr = flag.String("digits", "", "A comma-separated list of digits")

	targetRange = flag.String("target_range", "", "The target range to produce solutions for (inclusive)")
	target      = flag.Int("target", 0, "The exact target value to solve for")
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

func parseTargetRange(s string) (int, int, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("want 2 comma-separated values, got %d", len(s))
	}

	min, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing %q: %v", parts[0], err)
	}
	max, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing %q: %v", parts[1], err)
	}

	if min <= 0 {
		return 0, 0, fmt.Errorf("range lower bound must be positive, got %d", min)
	}
	if max <= 0 {
		return 0, 0, fmt.Errorf("range upper bound must be positive, got %d", max)
	}

	// Just flip inverted ranges.
	if min > max {
		return max, min, nil
	}
	return min, max, nil
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

func makeAdd(a, b expression) expression {
	return expression{Val: a.Val + b.Val, Op: opAdd, AExp: &a, BExp: &b}
}

func makeSubtract(a, b expression) expression {
	// TODO: check positive result?
	return expression{Val: a.Val - b.Val, Op: opSubtract, AExp: &a, BExp: &b}
}

func makeMultiply(a, b expression) expression {
	return expression{Val: a.Val * b.Val, Op: opMultiply, AExp: &a, BExp: &b}
}

func makeDivide(a, b expression) expression {
	// TODO: check exact?
	if b.Val == 0 {
		panic("denominator is zero")
	}
	return expression{Val: a.Val / b.Val, Op: opDivide, AExp: &a, BExp: &b}
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
				solutions = append(solutions, makeAdd(aExp, soln))
			}
		}

		// Subtraction.
		if a > target {
			for _, soln := range solve(a-target, other) {
				solutions = append(solutions, makeSubtract(aExp, soln))
			}
		}
		for _, soln := range solve(target+a, other) {
			solutions = append(solutions, makeSubtract(soln, aExp))
		}

		// Multiplication.
		if (target % a) == 0 {
			for _, soln := range solve(target/a, other) {
				solutions = append(solutions, makeMultiply(aExp, soln))
			}
		}

		// Division.
		if (a % target) == 0 {
			for _, soln := range solve(a/target, other) {
				solutions = append(solutions, makeDivide(aExp, soln))
			}
		}
		for _, soln := range solve(target*a, other) {
			solutions = append(solutions, makeDivide(soln, aExp))
		}
	}

	// TODO: divide digits into two sets. For each solution in set A, see if there is a solution in set B which will form the target.

	for _, soln := range solutions {
		if soln.Val != target {
			panic(fmt.Sprintf("generated invalid solution: %s = %d, want %d", soln, soln.Val, target))
		}
	}

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

func shortest(solns []expression) (expression, error) {
	shortest := ""
	var shortestSoln expression

	for _, soln := range solns {
		str := soln.String()
		if shortest == "" || len(str) < len(shortest) {
			shortest = str
			shortestSoln = soln
		}
	}
	return shortestSoln, nil
}

func main() {
	flag.Parse()

	if *digitsStr == "" {
		log.Fatalf("--digits must be provided")
	}
	digits, err := parseDigits(*digitsStr)
	if err != nil {
		log.Fatalf("--digits invalid: %v", err)
	}

	switch {
	case *targetRange != "":
		min, max, err := parseTargetRange(*targetRange)
		if err != nil {
			log.Fatalf("--target_range invalid: %v", err)
		}
		for i := min; i <= max; i++ {
			solns := solve(i, digits)
			fmt.Printf("%d: %d solutions found\n", i, len(solns))
		}
	case *target != 0:
		solns := solve(*target, digits)
		if len(solns) == 0 {
			fmt.Printf("no solution found :(\n")
			return
		}

		for i, soln := range solns {
			result, ok := soln.eval()
			if !ok {
				log.Fatalf("result is invalid")
			}
			if result != *target {
				log.Fatalf("generated incorrect solution: %s = %d, != %d!", soln, result, *target)
			}
			fmt.Printf("%d: %d = %s\n", i, result, soln)
		}

		shortest, err := shortest(solns)
		if err != nil {
			log.Fatalf("Failed to get shortest solution: %v", err)
		}
		fmt.Printf("Shortest solution: %s\n", shortest)
	default:
		log.Fatalf("--target or --solve_all must be provided")
	}
}
