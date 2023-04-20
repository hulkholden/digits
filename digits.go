package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
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
	opNegate
)

var opStrings = map[operation]string{
	opAdd:      "+",
	opSubtract: "-",
	opMultiply: "*",
	opDivide:   "/",
	opNegate:   "-",
}

func (op operation) commutative() bool {
	return op == opAdd || op == opMultiply
}

func (op operation) evalBinary(a, b int) (int, bool) {
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
	Op       operation
	Children []*expression
}

func makeConstant(v int) expression {
	return expression{Val: v}
}

func makeNegate(a expression) expression {
	return expression{Val: -a.Val, Op: opNegate, Children: []*expression{&a}}
}

func makeAdd(a, b expression) expression {
	return expression{Val: a.Val + b.Val, Op: opAdd, Children: []*expression{&a, &b}}
}

func makeSubtract(a, b expression) expression {
	// TODO: check positive result?
	return expression{Val: a.Val - b.Val, Op: opSubtract, Children: []*expression{&a, &b}}
}

func makeMultiply(a, b expression) expression {
	return expression{Val: a.Val * b.Val, Op: opMultiply, Children: []*expression{&a, &b}}
}

func makeDivide(a, b expression) expression {
	// TODO: check exact?
	if b.Val == 0 {
		panic("denominator is zero")
	}
	return expression{Val: a.Val / b.Val, Op: opDivide, Children: []*expression{&a, &b}}
}

func (e expression) String() string {
	if e.Op == opNone {
		return fmt.Sprintf("%d", e.Val)
	}

	if e.Op == opNegate {
		if len(e.Children) != 1 {
			panic(fmt.Sprintf("want 1 operand for negate, got %d", len(e.Children)))
		}
		return fmt.Sprintf("-%s", e.Children[0].String())
	}

	children := make([]string, len(e.Children))
	for i, c := range e.Children {
		children[i] = c.String()
	}

	return fmt.Sprintf("(%s)", strings.Join(children, fmt.Sprintf(" %s ", e.Op.String())))
}

func (e expression) eval() (int, bool) {
	if e.Op == opNone {
		return e.Val, true
	}
	if e.Op == opNegate {
		if len(e.Children) != 1 {
			panic(fmt.Sprintf("want 1 operand for negate, got %d", len(e.Children)))
		}
		operand, ok := e.Children[0].eval()
		if !ok {
			return 0, false
		}
		return -operand, true
	}

	var val int
	for i, c := range e.Children {
		operand, ok := c.eval()
		if !ok {
			return 0, false
		}

		if i == 0 {
			val = operand
		} else {
			val, ok = e.Op.evalBinary(val, operand)
			if !ok {
				return 0, false
			}
		}
	}
	return val, true
}

// fuse merges nested expressions like (a + (b + c)) into (a + b + c)
func (e expression) fuse() expression {
	// TODO: we can do this for opSubtract and opDiv too, but we need to make sure first element stays the same.
	// Or, we could represent subtraction as addition over negated values?
	if !e.Op.commutative() {
		return e
	}

	newChildren := make([]*expression, 0, len(e.Children))
	for _, c := range e.Children {
		if c.Op != e.Op {
			newChildren = append(newChildren, c)
		} else {
			newChildren = append(newChildren, c.Children...)
		}
	}

	e.Children = newChildren
	return e
}

// canonicalize ensures commutative operations are always expressed consistently (lowest operand first).
func (e expression) canonicalize() expression {
	// Sort operands by magnitude (largest to smallest).
	if e.Op.commutative() {
		slices.SortFunc(e.Children, func(a, b *expression) bool { return abs(a.Val) > abs(b.Val) })
	}
	return e
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
				solutions = append(solutions, makeAdd(aExp, makeNegate(soln)))
			}
		}
		for _, soln := range solve(target+a, other) {
			solutions = append(solutions, makeAdd(soln, makeNegate(aExp)))
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

	// Normalize and remove duplicates.
	seen := make(map[string]bool)
	var solsOut []expression
	for _, s := range solutions {
		s = s.fuse()
		s = s.canonicalize()
		key := s.String()
		if !seen[key] {
			seen[key] = true
			solsOut = append(solsOut, s)
		}
	}
	return solsOut
}

func shortest(solns []expression) (expression, error) {
	if len(solns) == 0 {
		return expression{}, fmt.Errorf("no solutions")
	}

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

	// TODO: When printing out the solution we want to show binary operations (i.e. unfuse the n-ary operations).

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
