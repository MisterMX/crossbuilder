package apis

//go:generate go run -tags generate ../../../cmd/xrd-gen paths=./... xrd:allowDangerousTypes=true,crdVersions=v1 output:artifacts:config=../package/xrds

// This example relies on the XRD generator's default API version, v2.