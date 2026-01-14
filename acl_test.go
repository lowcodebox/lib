package lib

import "fmt"

func ExampleCreateACLValue() {
	acl := map[string]uint16{}

	acl["1"], _ = CreateACLValue(ACLPriorityOthers, ACLPermissionAllow, ACLPermissionAllow, ACLPermissionAllow, ACLPermissionAllow)
	acl["2"], _ = CreateACLValue(ACLPriorityRole, ACLPermissionAllow, ACLPermissionNull, ACLPermissionDeny, ACLPermissionAllow)

	//fmt.Println(acl)

	token, err := GenTokenACL(acl, []byte("POIlhb123Y09olUi"), "1234", 0)
	if err != nil {
		fmt.Println(err)
	}

	r, w, x, a, err := ParseTokenACL(token, "1234", "1,2", []byte("POIlhb123Y09olUi"))
	fmt.Println(r, w, x, a, err)

	// Output:
	// true true false true <nil>
}
