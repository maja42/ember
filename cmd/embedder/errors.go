package main

//
//import (
//	"fmt"
//	"os"
//)
//
//type FatalErr struct {
//	ExitCode int
//	Message  string
//}
//
//// Exitf quits the application with the given ExitCode after executing all deferred function calls.
//func Exitf(exitCode int, format string, args ...interface{}) {
//	// panic ensures that all defers() are called before exiting the application
//	panic(FatalErr{
//		ExitCode: exitCode,
//		Message:  fmt.Sprintf(format, args...),
//	})
//}
//
//// CheckExit check the given error. If non-nil, Exitf is called to quit the application.
//func CheckExit(err error, exitCode int, format string, args ...interface{}) {
//	if err != nil {
//		Exitf(exitCode, format, args...)
//	}
//}
//
//// HandleExit recovers panics triggered via Exitf and quits the application with the given exit code.
//// Must be defer-called before using Exitf.
//func HandleExit(logger Logger) {
//	if err := recover(); err != nil {
//		if fatalErr, ok := err.(FatalErr); ok {
//			logger.Printf("%s\n", fatalErr.Message)
//			os.Exit(fatalErr.ExitCode)
//		} else {
//			logger.Printf("%v\n", err)
//			os.Exit(-1)
//		}
//	}
//}
