package calculator

import "golang.org/x/exp/constraints"

func Add[T interface {
	constraints.Integer | constraints.Float | constraints.Complex
}](nums ...T) T {
	var out T
	for _, n := range nums {
		out += n
	}
	return out
}

func Sub[T interface {
	constraints.Integer | constraints.Float | constraints.Complex
}](nums ...T) T {
	if len(nums) == 0 {
		return 0
	}
	out := nums[0]
	for _, n := range nums[1:] {
		out -= n
	}
	return out
}

func Mul[T interface {
	constraints.Integer | constraints.Float | constraints.Complex
}](nums ...T) T {
	if len(nums) == 0 {
		return 0
	}
	out := nums[0]
	for _, n := range nums[1:] {
		out *= n
	}
	return out
}

func Div[T interface {
	constraints.Integer | constraints.Float | constraints.Complex
}](n T, d T) T {
	return n / d
}
