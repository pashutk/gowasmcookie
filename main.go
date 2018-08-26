package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"syscall/js"
)

func parseArgs(args []js.Value) (string, js.Value, error) {
	if len(args) < 2 {
		return "", js.Null(), errors.New("Too few parameters")
	}

	if args[0].Type() != js.TypeString {
		return "", js.Null(), errors.New("First parameter should be String type")
	}

	if args[1].Type() != js.TypeFunction {
		return "", js.Null(), errors.New("Second parameter should be Function type")
	}

	return args[0].String(), args[1], nil
}

func getCookie(key string) (string, bool, error) {
	documentProperty := js.Global().Get("document")
	if documentProperty.Type() != js.TypeObject {
		return "", false, errors.New("No document property of global object")
	}

	cookieProperty := documentProperty.Get("cookie")
	if cookieProperty.Type() != js.TypeString {
		return "", false, errors.New("document.cookie is not a string")
	}

	cookieString := cookieProperty.String()
	cookies := strings.Split(cookieString, "; ")
	for _, cookie := range cookies {
		parts := strings.Split(cookie, "=")

		if len(parts) < 2 {
			return "", false, nil
		}

		unescaped, err := url.PathUnescape(parts[1])
		if err != nil {
			return unescaped, false, err
		}

		if parts[0] == key {
			return unescaped, true, nil
		}
	}

	return "", false, nil
}

func jsGetCookie(args []js.Value) {
	key, cb, err := parseArgs(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	result, ok, err := getCookie(key)
	if err != nil {
		fmt.Println(err)
	}

	if ok {
		cb.Invoke(result)
	} else {
		cb.Invoke(js.Null())
	}
}

func main() {
	libcookie := js.NewCallback(func(args []js.Value) {}).New()
	libcookie.Set("get", js.NewCallback(jsGetCookie))

	js.Global().Set("libcookie", libcookie)

	done := make(chan bool)
	<-done
}
