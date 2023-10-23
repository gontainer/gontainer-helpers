/*
Package grouperror provides a toolset to join and split errors.

	err := grouperror.Prefix("my group: ", fmt.Errorf("error1), nil, fmt.Errorf("error2"))
	errs := grouperror.Collection(err) // []errors{error("my group: error1"), error("my group: error2")}
*/
package grouperror
