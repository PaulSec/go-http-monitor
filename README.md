Go-mon
=====

This small utility is dead simple and will allow you to monitor HTTP endpoints, and easily pluggable in the CI.
Everything is configurable through a YAML file (`monitor.yml`) and looks like this: 

```yaml
insecure: false
timeout_seconds: 5
plugins:
  - url: "https://www.cfptime.org"
    status_code: 200
    match: "Loading"
  - url: "https://shodan.io/"
    status_code: 200
    match: "the Internet of Things"
```

Options
=======

| Option  | What does it mean |
| ------------- | ------------- |
| insecure  | Skip SSL/TLS verification (in case of self-signed certificates, ..)  |
| timeout_seconds  | How long you'd like to wait for the HTTP request |
| checks  | A list of all the checks you'd like to perform |
| checks > url  | The URL you want to monitor |
| checks > status_code  | The status code you expect |
| checks > match  | One string you're looking for on the webpage |


Usage
=======

In order to launch it, just build it and run it this way: 

```bash
$ go build main.go && ./main
[OK] https://www.cfptime.org
[OK] https://shodan.io/
$ echo $?
0
```

If you add another check that will obviously fail: 

```yaml
  - url: "https://urlthatdoesnotexist.foobar"
    status_code: 500
    match: "the Internet of Things"
```

Resulting output will be: 

```bash
$ ./main
[OK] https://www.cfptime.org
[OK] https://shodan.io/
[NOK] https://urlthatdoesnotexist.foobar
$ echo $?
1
```

Contributing
=======

Feel free to fork the project and do whatever you want with it. 