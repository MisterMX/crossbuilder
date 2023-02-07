package xrd

import (
	xapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// XRDMarker defines a marker for XRD types.
type XRDMarker interface { // nolint:golint
	ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error
}

// PackageOverride overrides the loading of some package
// (potentially setting custom schemata, etc).  It must
// call AddPackage if it wants to continue with the default
// loading behavior.
type PackageOverride func(p *Parser, pkg *loader.Package)

// Parser is used to apply XRD markers.
type Parser struct {
	Collector *markers.Collector

	// Types contains the known TypeInfo for this parser.
	Types map[crd.TypeIdent]*markers.TypeInfo
	// GroupVersions contains the known group-versions of each package in this parser.
	GroupVersions map[*loader.Package]schema.GroupVersion

	// PackageOverrides indicates that the loading of any package with
	// the given path should be handled by the given overrider.
	PackageOverrides map[string]PackageOverride

	// checker stores persistent partial type-checking/reference-traversal information.
	Checker *loader.TypeChecker

	// packages marks packages as loaded, to avoid re-loading them.
	packages map[*loader.Package]struct{}
}

func (p *Parser) init() {
	if p.packages == nil {
		p.packages = make(map[*loader.Package]struct{})
	}
	if p.Types == nil {
		p.Types = make(map[crd.TypeIdent]*markers.TypeInfo)
	}
	if p.PackageOverrides == nil {
		p.PackageOverrides = make(map[string]PackageOverride)
	}
	if p.GroupVersions == nil {
		p.GroupVersions = make(map[*loader.Package]schema.GroupVersion)
	}
}

// indexTypes loads all types in the package into Types.
func (p *Parser) indexTypes(pkg *loader.Package) {
	// autodetect
	pkgMarkers, err := markers.PackageMarkers(p.Collector, pkg)
	if err != nil {
		pkg.AddError(err)
	} else {
		if skipPkg := pkgMarkers.Get("kubebuilder:skip"); skipPkg != nil {
			return
		}
		if nameVal := pkgMarkers.Get("groupName"); nameVal != nil {
			versionVal := pkg.Name // a reasonable guess
			if versionMarker := pkgMarkers.Get("versionName"); versionMarker != nil {
				versionVal = versionMarker.(string)
			}

			p.GroupVersions[pkg] = schema.GroupVersion{
				Version: versionVal,
				Group:   nameVal.(string),
			}
		}
	}

	if err := markers.EachType(p.Collector, pkg, func(info *markers.TypeInfo) {
		ident := crd.TypeIdent{
			Package: pkg,
			Name:    info.Name,
		}

		p.Types[ident] = info
	}); err != nil {
		pkg.AddError(err)
	}
}

// NeedCRDFor lives off in spec.go

// AddPackage indicates that types and type-checking information is needed
// for the the given package, *ignoring* overrides.
// Generally, consumers should call NeedPackage, while PackageOverrides should
// call AddPackage to continue with the normal loading procedure.
func (p *Parser) AddPackage(pkg *loader.Package) {
	p.init()
	if _, checked := p.packages[pkg]; checked {
		return
	}
	p.indexTypes(pkg)
	p.Checker.Check(pkg)
	p.packages[pkg] = struct{}{}
}

// NeedPackage indicates that types and type-checking information
// is needed for the given package.
func (p *Parser) NeedPackage(pkg *loader.Package) {
	p.init()
	if _, checked := p.packages[pkg]; checked {
		return
	}
	// overrides are going to be written without vendor.  This is why we index by the actual
	// object when we can.
	if override, overridden := p.PackageOverrides[loader.NonVendorPath(pkg.PkgPath)]; overridden {
		override(p, pkg)
		p.packages[pkg] = struct{}{}
		return
	}
	p.AddPackage(pkg)
}

// ApplyForXRD applies all markers to the generated XRD.
func (p *Parser) ApplyForXRD(xrd *xapiext.CompositeResourceDefinition) {
	packages := []*loader.Package{}
	for pkg, gv := range p.GroupVersions {
		if gv.Group != xrd.Spec.Group {
			continue
		}
		packages = append(packages, pkg)
	}

	// apply markers
	for _, pkg := range packages {
		typeIdent := crd.TypeIdent{Package: pkg, Name: xrd.Spec.Names.Kind}
		typeInfo := p.Types[typeIdent]
		if typeInfo == nil {
			continue
		}
		ver := p.GroupVersions[pkg].Version

		for _, markerVals := range typeInfo.Markers {
			for _, val := range markerVals {
				xrdMarker, isXRDMarker := val.(XRDMarker)
				if !isXRDMarker {
					continue
				}
				if err := xrdMarker.ApplyToXRD(xrd, ver); err != nil {
					pkg.AddError(loader.ErrFromNode(err /* an okay guess */, typeInfo.RawSpec))
				}
			}
		}
	}
}
