package task_queue

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
)

func inspectFunction(function WorkerFunctionType) (functionMetadata, error) {
	metadata := functionMetadata{}

	// ensure provided function is not nil
	if function == nil {
		return metadata, errors.New("function is nil")
	}
	fnValue := reflect.ValueOf(function)
	fnType := fnValue.Type()

	// verify that function is type of func
	if fnType.Kind() != reflect.Func {
		foundType := fnType.Kind().String()
		return metadata, errors.New("function is not type of func. Found " + foundType + " instead")
	}

	// function name
	functionName := runtime.FuncForPC(fnValue.Pointer()).Name()

	// verify function is exported
	if !isFunctionExported(functionName) {
		return metadata, errors.New("function `" + functionName + "` is not exported. Your function name should start with a capital letter to be exported")
	}
	// set function name
	metadata.functionName = functionName
	metadata.function = function

	// ensure provided function has only one return type and it's an error
	if fnType.NumOut() != 1 || fnType.Out(0).Kind() != reflect.Interface || fnType.Out(0).Name() != "error" {
		// build tuple of return types
		returnTypes := "("
		for i := 0; i < fnType.NumOut(); i++ {
			returnTypes += fnType.Out(i).Name()
			if i != fnType.NumOut()-1 {
				returnTypes += ", "
			}
		}
		returnTypes += ")"
		return metadata, errors.New("function must return only one error. Found " + returnTypes)
	}

	// ensure provided function has only one argument
	if fnType.NumIn() != 3 {
		return metadata, errors.New("function must have only one custom argument and a context and cancel context function")
	}

	// ensure values are provided as value and not as pointer
	if fnType.In(0).Kind() == reflect.Ptr {
		return metadata, errors.New("function argument must be a value and not a pointer")
	}

	// ensure 2nd argument is context
	if fnType.In(1).Kind() != reflect.Interface || fnType.In(1).Name() != "Context" {
		return metadata, errors.New("function 2nd argument must be a context")
	}

	// ensure 3rd argument is cancel context function
	if fnType.In(2).Kind() != reflect.Func || fnType.In(2).Name() != "CancelFunc" {
		return metadata, errors.New("function 3rd argument must be a cancel function")
	}

	// ensure values are need to be some kind of struct
	if fnType.In(0).Kind() != reflect.Struct {
		foundType := fnType.In(0).Kind().String()
		return metadata, errors.New("function argument must be a struct, found " + foundType + " instead")
	}

	// ensure argument type is exported
	if !isArgumentTypeExported(fnType.In(0).Name()) {
		return metadata, errors.New("argument type `" + fnType.In(0).Name() + "` is not exported. Your argument type should start with a capital letter to be exported")
	}

	// validate that each field of the struct has a json tag
	for i := 0; i < fnType.In(0).NumField(); i++ {
		if strings.Compare(fnType.In(0).Field(i).Tag.Get("json"), "") == 0 {
			return metadata, errors.New("field " + fnType.In(0).Field(i).Name + " of struct " + fnType.In(0).Name() + " is missing json tag")
		}
	}

	// set the metadata
	metadata.argumentTypeName = fnType.In(0).Name()
	metadata.argumentType = fnType.In(0)

	return metadata, nil
}
