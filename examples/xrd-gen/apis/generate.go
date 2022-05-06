package apis

//go:generate rm -rf ../package/crds

//go:generate go run -tags generate ../../../cmd/xrd-gen xrd paths=./... xrd:allowDangerousTypes=true,crdVersions=v1 object:headerFile=../../../hack/boilerplate.go.txt,year=2022 output:artifacts:config=../package/xrds
