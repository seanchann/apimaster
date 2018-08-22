/*

Copyright 2018 This Project Authors.

Author:  seanchann <seanchann@foxmail.com>

See docs/ for more information about the  project.

*/

package swaggerdoc

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/golang/glog"
	"github.com/seanchann/apimaster/pkg/swaggerdoc/options"
)

//GenerateDoc generate swagger doc by options
// visit https://github.com/emicklei/go-restful-swagger12 for detail
func GenerateDoc(o *options.SwaggerDocOptions) {
	if o.TypeSrc == "" {
		glog.Fatalf("Please define -s flag as it is the source file")
	}

	var funcOut io.Writer
	if o.FunctionDest == "-" {
		funcOut = os.Stdout
	} else {
		file, err := os.Create(o.FunctionDest)
		if err != nil {
			glog.Fatalf("Couldn't open %v: %v", o.FunctionDest, err)
		}
		defer file.Close()
		funcOut = file
	}

	if o.HeaderFile != "" {

		b, err := ioutil.ReadFile(o.HeaderFile)
		if err != nil {
			glog.Fatalf("Error input header file %s\n", err)
		}

		io.WriteString(funcOut, string(b))
	}

	docsForTypes := runtime.ParseDocumentationFrom(o.TypeSrc)

	if o.Verify == true {
		rc, err := runtime.VerifySwaggerDocsExist(docsForTypes, funcOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in verification process: %s\n", err)
		}
		os.Exit(rc)
	}

	if docsForTypes != nil && len(docsForTypes) > 0 {
		if err := runtime.WriteSwaggerDocFunc(docsForTypes, funcOut); err != nil {
			fmt.Fprintf(os.Stderr, "Error when writing swagger documentation functions: %s\n", err)
			os.Exit(-1)
		}
	}
}
