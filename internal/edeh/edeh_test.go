package edeh

import (
	"fmt"
)

func Example_splitHeader() {
	sample := func(line string) {
		name, eno, body, err := splitHeader([]byte(line))
		if err == nil {
			fmt.Printf("%s:%d[%s]\n", name, eno, string(body))
		} else {
			fmt.Println(err)
		}
	}

	sample("-payload 1-")
	sample("foo\t-payload 2-")
	sample("bar:\t-payload 3-")
	sample(":1\t-payload 4-")
	sample("baz:1\t-payload 5-")
	// Output:
	// no event prefix in `line:-payload 1-`
	// foo:-1[-payload 2-]
	// strconv.Atoi: parsing "": invalid syntax
	// no filename in journal prefix of `line::1	-payload 4-`
	// baz:1[-payload 5-]
}
