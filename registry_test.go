package lib

import (
	"fmt"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

func ExampleOptimizePathMeshDC() {
	var pongs []models.Pong
	fmt.Println(OptimizePathMeshDC(pongs, ""))

	pongs = []models.Pong{{Host: "http://127.0.0.1", PortHTTP: 2080, DC: "el"}}
	fmt.Println(OptimizePathMeshDC(pongs, ""))
	fmt.Println(OptimizePathMeshDC(pongs, "el"))

	pongs = []models.Pong{{Host: "http://127.0.0.1", PortHTTP: 2080, DC: "el"}, {Host: "http://127.0.0.1", PortHTTP: 2090, DC: "dp"}}
	fmt.Println(OptimizePathMeshDC(pongs, "el"))

	// Output:
	// [] []
	// [] [http://127.0.0.1:2080]
	// [http://127.0.0.1:2080] []
	// [http://127.0.0.1:2080] [http://127.0.0.1:2090]
}
