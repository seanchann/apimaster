/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package options

import "github.com/spf13/pflag"

//SwaggerDocOptions swagger doc options
type SwaggerDocOptions struct {
	FunctionDest string
	TypeSrc      string
	Verify       bool
	HeaderFile   string
	Package      string
}

//NewSwaggerDocOptions new swagger options
func NewSwaggerDocOptions() *SwaggerDocOptions {
	return &SwaggerDocOptions{}
}

//AddFlags add flags
func (s *SwaggerDocOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.FunctionDest, "func-dest", "-", "Output for swagger functions; '-' means stdout (default)")
	fs.StringVar(&s.TypeSrc, "type-src", "", "From where we are going to read the types")
	fs.BoolVar(&s.Verify, "verify", false, "Verifies if the given type-src file has documentation for every type")
	fs.StringVar(&s.HeaderFile, "header-file", "", "append your header file")
	fs.StringVar(&s.Package, "package-name", "", "output file in what package")
}
