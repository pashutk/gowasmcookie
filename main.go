package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"syscall/js"
)

func parseGetArgs(args []js.Value) (string, js.Value, error) {
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

type cookieOptions struct {
	expires string
	path    string
	domain  string
	secure  bool
}

func parseCookieOptions(optionsArg js.Value) (cookieOptions, error) {
	options := cookieOptions{}

	if optionsArg.Type() != js.TypeUndefined {
		if optionsArg.Type() != js.TypeObject {
			return options, errors.New("Third parameter should be undefined or Object type")
		}

		expiresValue := optionsArg.Get("expires")
		pathValue := optionsArg.Get("path")
		domainValue := optionsArg.Get("domain")
		secureValue := optionsArg.Get("secure")

		if expiresValue.Type() != js.TypeUndefined {
			dateClass := js.Global().Get("Date")
			if !expiresValue.InstanceOf(dateClass) {
				return options, errors.New("expires option should be instance of Date")
			}

			options.expires = expiresValue.Call("toUTCString").String()
		}

		if pathValue.Type() != js.TypeUndefined {
			if pathValue.Type() != js.TypeString {
				return options, errors.New("path option should be String type")
			}
			options.path = pathValue.String()
		}

		if domainValue.Type() != js.TypeUndefined {
			if domainValue.Type() != js.TypeString {
				return options, errors.New("domain option should be String type")
			}
			options.domain = domainValue.String()
		}

		if secureValue.Type() != js.TypeUndefined {
			if secureValue.Type() != js.TypeBoolean {
				return options, errors.New("secure option should be Boolean type")
			}
			options.secure = secureValue.Bool()
		}
	}

	return options, nil
}

func parseSetArgs(args []js.Value) (string, string, cookieOptions, js.Value, error) {
	var options cookieOptions

	if len(args) < 3 {
		return "", "", options, js.Null(), errors.New("Too few parameters")
	}

	nameArg := args[0]
	valueArg := args[1]
	optionsArg := args[2]
	callbackArg := args[3]

	if nameArg.Type() != js.TypeString {
		return "", "", options, js.Null(), errors.New("First parameter should be String type")
	}

	if valueArg.Type() != js.TypeString {
		return "", "", options, js.Null(), errors.New("Second parameter should be String type")
	}

	if callbackArg.Type() != js.TypeFunction {
		return "", "", options, js.Null(), errors.New("Fourth parameter should be Function type")
	}

	options, err := parseCookieOptions(optionsArg)
	if err != nil {
		return "", "", options, js.Null(), err
	}

	return nameArg.String(), valueArg.String(), options, callbackArg, nil
}

func getDocument() (js.Value, error) {
	documentProperty := js.Global().Get("document")
	if documentProperty.Type() != js.TypeObject {
		return js.Null(), errors.New("No document property of global object")
	}

	return documentProperty, nil
}

func getCookieFromDocument(document js.Value) (string, error) {
	cookieProperty := document.Get("cookie")
	if cookieProperty.Type() != js.TypeString {
		return "", errors.New("document.cookie is not a string")
	}

	return cookieProperty.String(), nil
}

func getDocumentCookie() (string, error) {
	document, err := getDocument()
	if err != nil {
		return "", err
	}
	return getCookieFromDocument(document)
}

func getCookie(key string) (string, bool, error) {
	cookieString, err := getDocumentCookie()
	if err != nil {
		return "", false, err
	}

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

func prependCookieOption(optionName string) string {
	return "; " + optionName
}

func buildCookieOption(optionName, optionValue string) string {
	return prependCookieOption(optionName) + "=" + optionValue
}

func setCookie(key, value string, options cookieOptions) (string, error) {
	document, err := getDocument()
	if err != nil {
		return "", err
	}

	escapedValue := url.PathEscape(value)

	cookie := key + "=" + escapedValue

	if options.expires != "" {
		cookie += buildCookieOption("expires", options.expires)
	}

	if options.path != "" {
		cookie += buildCookieOption("path", options.path)
	}

	if options.domain != "" {
		cookie += buildCookieOption("domain", options.domain)
	}

	if options.secure {
		cookie += prependCookieOption("secure")
	}

	document.Set("cookie", cookie)

	return cookie, nil
}

func jsGetCookie(args []js.Value) {
	key, cb, err := parseGetArgs(args)
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

func jsSetCookie(args []js.Value) {
	key, value, options, cb, err := parseSetArgs(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	cookie, err := setCookie(key, value, options)
	if err != nil {
		fmt.Println(err)
		return
	}

	cb.Invoke(cookie)
}

func main() {
	libcookie := js.NewCallback(func(args []js.Value) {}).New()
	libcookie.Set("get", js.NewCallback(jsGetCookie))
	libcookie.Set("set", js.NewCallback(jsSetCookie))

	js.Global().Set("libcookie", libcookie)

	done := make(chan bool)
	<-done
}
