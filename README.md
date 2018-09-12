# Cookie library

## Prerequisites
```
golang 1.11
last Chrome browser
```

## Build
```
mkdir build
cp resources/index.html build
cp $(go env GOROOT)/misc/wasm/wasm_exec.js build
GOARCH=wasm GOOS=js go build -o build/gowasmcookie.wasm main.go && go run resources/server.go
```
Then open localhost:8080/build in Chrome and use library in Console or in your js code

## Usage
```
window.libcookie.get('cookiename', value => console.log(value));
window.libcookie.set('cookiename', 'cookievalue', {expires: new Date(), path: '/', domain: 'localhost', secure: true}, cookie => console.log(cookie));
window.libcookie.remove('cookiename', ok => console.log(ok));

```
