/*
Package caller provides functions that allow calling other functions with unknown arguments.

# Example

	sum := func(a, b int) int {
		return a + b
	}

	returns, _ := caller.Call(sum, 2, 3)
	fmt.Println(returns) // [5]

# Provider

It is a function that returns 1 or 2 values. The first value is the desired output of the provider.
The optional second value may contain information about a potential error.

Provider that does not return any error:

	func NewHttpClient(timeout time.Duration) *http.Client {
		return &http.Client{
			Timeout: timeout,
		}
	}

	// httpClient, _ := caller.CallProvider(NewHttpClient, time.Minute)

Provider that may return an error:

	func NewDB(username string, password string) (any, error) {
	    db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/test", username, password))
	    if err != nil {
	         return nil, err
	    }

	    db.SetConnMaxLifetime(time.Minute * 3)
	    db.SetMaxOpenConns(10)
	    db.SetMaxIdleConns(10)

	    return db, nil
	}

	// db, err := caller.CallProvider(NewDB, "root", "root")
	// if err != nil {
	// 	panic(err)
	// }

# Wither

It is a method that returns one value always:

	type Person struct {
		Name string
	}

	func (p Person) WithName(n string) Person {
		p.Name = n
		return p
	}

	// p, _ := caller.CallWitherByName(caller.Person{}, "WithName", "Jane")
	// fmt.Println(p) // {Jane}
*/
package caller
