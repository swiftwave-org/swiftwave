package task_queue

import (
	"context"
	"reflect"
	"strings"
)

func getTypeName(object interface{}) string {
	val := reflect.ValueOf(object)
	if val.Kind() == reflect.Ptr {
		return val.Elem().Type().Name()
	} else {
		return val.Type().Name()
	}
}

func invokeFunction(function interface{}, argument interface{}, argumentType ArgumentType) error {
	// create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create argument slice
	argumentSlice := []reflect.Value{
		reflect.ValueOf(argument),
		reflect.ValueOf(ctx),
		reflect.ValueOf(cancel),
	}
	// invoke function
	functionValue := reflect.ValueOf(function)
	returnValues := functionValue.Call(argumentSlice)
	// check for errors
	if len(returnValues) > 0 {
		if returnValues[0].Interface() != nil {
			return returnValues[0].Interface().(error)
		}
	}
	return nil
}

func isFunctionExported(functionName string) bool {
	name := strings.Split(functionName, ".")
	if len(name) == 0 {
		return false
	}
	firstCharOfName := name[len(name)-1][0]
	if firstCharOfName < 'A' || firstCharOfName > 'Z' {
		return false
	}
	return true
}

func isArgumentTypeExported(argumentType string) bool {
	name := strings.Split(argumentType, ".")
	if len(name) == 0 {
		return false
	}
	firstCharOfName := name[len(name)-1][0]
	if firstCharOfName < 'A' || firstCharOfName > 'Z' {
		return false
	}
	return true
}
