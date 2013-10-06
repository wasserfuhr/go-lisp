package lisp

import "fmt"

func EvalString(line string) (string, error) {
	tokenized := Tokenize(line)
	parsed, err := Parse(tokenized)
	if err != nil {
		return "", err
	}
	evaled, err := Eval(parsed)
	if err != nil {
		return "", err
	}
	scope.Create("_", evaled)
	return fmt.Sprintf("%v", evaled), nil
}

func Eval(expr Sexp) (val Value, err error) {
	for _, t := range expr {
		val, err = evalValue(t)
		if err != nil {
			break
		}
	}
	return
}

func evalValue(input Value) (val Value, err error) {
	switch input.Kind {
	case SexpKind:
		expr := input.Sexp()
		if len(expr) > 0 {
			switch expr[0].String() {
			case "quote":
				return quoteForm(expr)
			case "if":
				return ifForm(expr)
			case "set!":
				return setForm(expr)
			case "define":
				return defineForm(expr)
			case "lambda":
				return lambdaForm(expr)
			case "begin":
				return beginForm(expr)
			default:
				if isBuiltin(expr[0]) {
					return runBuiltin(expr)
				} else {
					return procForm(expr)
				}
			}
		}
	case NumberKind, StringKind:
		val = input
	case SymbolKind:
		sym := input.String()
		if v, ok := scope.Get(sym); ok {
			val = v
		} else if sym == "true" || sym == "false" {
			val = Value{SymbolKind, sym}
		} else {
			return Nil, fmt.Errorf("Unbound variable: %v", sym)
		}
	default:
		return Nil, fmt.Errorf("Unknown data type: %v", input)
	}
	return
}

func procForm(expr Sexp) (val Value, err error) {
	if val, err = evalValue(expr[0]); err == nil {
		if val.IsA(ProcKind) {
			var args []Value
			for _, v := range expr[1:] {
				if e, err := evalValue(v); err != nil {
					return Nil, err
				} else {
					args = append(args, e)
				}
			}
			val, err = val.Proc().Call(args)
		} else {
			err = fmt.Errorf("The object %v is not applicable", val)
		}
	}
	return
}

func beginForm(expr Sexp) (val Value, err error) {
	return Eval(expr[1:])
}

func setForm(expr Sexp) (val Value, err error) {
	if len(expr) == 3 {
		key := expr[1].String()
		if _, ok := scope.Get(key); ok {
			val, err = evalValue(expr[2])
			if err == nil {
				scope.Set(key, val)
			}
		} else {
			err = fmt.Errorf("Unbound variable: %v", key)
		}
	} else {
		err = fmt.Errorf("Ill-formed special form: %v", expr)
	}
	return
}

func ifForm(expr Sexp) (val Value, err error) {
	if len(expr) < 3 || len(expr) > 4 {
		err = fmt.Errorf("Ill-formed special form: %v", expr)
	} else {
		r, err := evalValue(expr[1])
		if err == nil {
			if !(r.IsA(SymbolKind) && r.String() == "false") && r != Nil && len(expr) > 2 {
				val, err = evalValue(expr[2])
			} else if len(expr) == 4 {
				val, err = evalValue(expr[3])
			}
		}
	}
	return
}

func lambdaForm(expr Sexp) (val Value, err error) {
	if len(expr) > 2 {
		params := expr[1].Sexp()
		val = Value{ProcKind, Proc{params, expr[2:], scope.Dup()}}
	} else {
		err = fmt.Errorf("Ill-formed special form: %v", expr)
	}
	return
}

func quoteForm(expr Sexp) (val Value, err error) {
	if len(expr) == 2 {
		val = expr[1]
	} else {
		err = fmt.Errorf("Ill-formed special form: %v", expr)
	}
	return
}

func defineForm(expr Sexp) (val Value, err error) {
	if len(expr) >= 2 && len(expr) <= 3 {
		if expr[1].IsA(SymbolKind) {
			key := expr[1].String()
			if len(expr) == 3 {
				var i Value
				if i, err = evalValue(expr[2]); err == nil {
					scope.Create(key, i)
				}
			} else {
				scope.Create(key, Nil)
			}
			return expr[1], err
		}
	}
	return Nil, fmt.Errorf("Ill-formed special form: %v", expr)
}
