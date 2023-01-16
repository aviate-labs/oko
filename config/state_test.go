package config_test

import (
	"fmt"
	"testing"

	"github.com/internet-computer/oko/config"
)

func ExamplePackageState() {
	json, _ := config.EmptyState().ToJSON()
	fmt.Println(string(json))
	// Output:
	// {
	//	"dependencies": []
	// }
}

func ExamplePackageState_AddPackage() {
	state := config.EmptyState()
	_ = state.AddPackage(config.PackageInfoRemote{
		Name:       "test",
		Repository: "url",
		Version:    "*",
	})
	json, _ := state.ToJSON()
	fmt.Println(string(json))
	// Output:
	// {
	// 	"dependencies": [
	// 		{
	// 			"name": "test",
	// 			"repository": "url",
	// 			"version": "*"
	// 		}
	// 	]
	// }
}

func ExamplePackageState_AddPackage_alreadyExits() {
	state := config.EmptyState()
	dep := config.PackageInfoRemote{
		Name:       "test",
		Repository: "url",
		Version:    "*",
	}
	_ = state.AddPackage(dep)
	_ = state.AddPackage(config.PackageInfoRemote{
		Dependencies: []string{"test"},
	}, dep)
	json, _ := state.ToJSON()
	fmt.Println(string(json))
	// Output:
	// {
	// 	"dependencies": [
	// 		{
	// 			"name": "",
	// 			"repository": "",
	// 			"version": "",
	// 			"dependencies": [
	// 				"test"
	// 			]
	// 		},
	// 		{
	// 			"name": "test",
	// 			"repository": "url",
	// 			"version": "*"
	// 		}
	// 	]
	// }
}

func ExamplePackageState_AddPackage_fromTransitive() {
	state := config.EmptyState()
	dep := config.PackageInfoRemote{
		Name:       "test",
		Repository: "url",
		Version:    "*",
	}
	_ = state.AddPackage(config.PackageInfoRemote{
		Name:         "test-v0.1.0",
		Repository:   "url",
		Version:      "v0.1.0",
		Dependencies: []string{"test"},
	}, dep)
	json, _ := state.ToJSON()
	fmt.Println(string(json))
	_ = state.AddPackage(dep)
	json, _ = state.ToJSON()
	fmt.Println(string(json))
	// Output:
	// {
	// 	"dependencies": [
	// 		{
	// 			"name": "test-v0.1.0",
	// 			"repository": "url",
	// 			"version": "v0.1.0",
	// 			"dependencies": [
	// 				"test"
	// 			]
	// 		}
	// 	],
	// 	"transitiveDependencies": [
	// 		{
	// 			"name": "test",
	// 			"repository": "url",
	// 			"version": "*"
	// 		}
	// 	]
	// }
	// {
	// 	"dependencies": [
	// 		{
	// 			"name": "test",
	// 			"repository": "url",
	// 			"version": "*"
	// 		},
	// 		{
	// 			"name": "test-v0.1.0",
	// 			"repository": "url",
	// 			"version": "v0.1.0",
	// 			"dependencies": [
	// 				"test"
	// 			]
	// 		}
	// 	]
	// }
}

func ExamplePackageState_RemovePackage() {
	state := config.EmptyState()
	_ = state.AddPackage(config.PackageInfoRemote{
		Name:       "test",
		Repository: "url",
		Version:    "*",
	})
	_ = state.RemovePackage("test")
	json, _ := state.ToJSON()
	fmt.Println(string(json))
	// Output:
	// {
	//	"dependencies": []
	// }
}

func TestPackageState_AddPackage(t *testing.T) {
	pkg := config.EmptyState()
	dep := config.PackageInfoRemote{
		Name:       "test",
		Repository: "url",
		Version:    "*",
	}
	if err := pkg.AddPackage(dep); err != nil {
		t.Fatal(err)
	}
	if err := pkg.AddPackage(dep); err == nil {
		t.Fatal()
	}
	if err := pkg.AddPackage(config.PackageInfoRemote{
		Dependencies: []string{"test"},
	}, dep); err != nil {
		t.Fatal(err)
	}
}

func TestPackageState_RemovePackage_otherDependency(t *testing.T) {
	state := config.EmptyState()
	dep := config.PackageInfoRemote{
		Name:       "test",
		Repository: "url",
		Version:    "*",
	}
	_ = state.AddPackage(dep)
	_ = state.AddPackage(config.PackageInfoRemote{
		Dependencies: []string{"test"},
	}, dep)
	// There is a dependency on "test".
	if err := state.RemovePackage("test"); err != nil {
		t.Error(err)
	}

	_ = state.AddPackage(config.PackageInfoRemote{
		Name:         "_",
		Version:      "_",
		Dependencies: []string{"test"},
	}, dep)
	if err := state.RemovePackage("_"); err != nil {
		t.Error(err)
	}
	if len(state.Dependencies) != 1 || len(state.TransitiveDependencies) != 1 {
		t.Error(state.Dependencies, state.TransitiveDependencies)
	}
}
