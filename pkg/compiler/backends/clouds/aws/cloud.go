// Copyright 2016 Marapongo, Inc. All rights reserved.

package aws

import (
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/marapongo/mu/pkg/ast"
	"github.com/marapongo/mu/pkg/compiler/backends/clouds"
	"github.com/marapongo/mu/pkg/compiler/core"
)

// New returns a fresh instance of an AWS Cloud implementation.  This targets "native AWS" for the code-gen outputs.
// This primarily means CloudFormation as the stack templating output, and idiomatic AWS services like S3, DynamoDB,
// Lambda, and so on, for the actual services in those stack templates.
//
// For more details, see https://github.com/marapongo/mu/blob/master/docs/targets.md#amazon-web-services-aws
func New() clouds.Cloud {
	return &awsCloud{}
}

type awsCloud struct {
	clouds.Cloud
	// TODO: support cloud provider options (e.g., ranging from simple like YAML vs. JSON to complex like IAM).
}

func (c *awsCloud) CodeGen(comp core.Compiland) {
	// For now, this routine simply generates the equivalent CloudFormation stack for the input.  Eventually this needs
	// to do a whole lot more, which the following running list of TODOs will serve as a reminder about:
	// TODO: perform delta analysis so that we can emit changesets:
	//     http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-changesets.html
	// TODO: allow for a "dry-run" mode that queries the target, checks things like limits, shows what will be done.
	// TODO: prepare full deployment packages (e.g., tarballs of code, Docker images, etc).
	nm := c.genStackName(comp)
	cf := c.genStackTemplate(comp)
	// TODO: actually save this (and any other outputs) to disk, rather than spewing to STDOUT.
	y, err := yaml.Marshal(cf)
	if err != nil {
		comp.Diag.Errorf(ErrorMarshalingCloudFormationTemplate.WithDocument(comp.Doc), err)
		return
	}
	fmt.Printf("%v:\n", nm)
	fmt.Println(string(y))
}

// genClusterTemplate creates a CloudFormation template for a standard overall cluster.
func (c *awsCloud) genClusterTemplate(comp core.Compiland) *cfTemplate {
	// TODO: this.
	return nil
}

// genStackName creates a name for the stack, which must be globally unique within an account.
func (c *awsCloud) genStackName(comp core.Compiland) string {
	return fmt.Sprintf("MuStack-%v-%v", comp.Target.Name, comp.Stack.Name)
}

// genStackTemplate creates a CloudFormation template for an entire stack and all of its services.
func (c *awsCloud) genStackTemplate(comp core.Compiland) *cfTemplate {
	// Allocate a new template object that we will populate and return.
	cf := &cfTemplate{
		AWSTemplateFormatVersion: cfVersion,
		Description:              comp.Stack.Description,
	}

	// TODO: add parameters.
	// TODO: due to the way we expand Mu templates, we don't leverage AWS CloudFormation parameters.  That's generally
	//     simpler, however, sometimes a customer may want the parameterization to persist (e.g., so they end up with
	//     a single CloudFormation template across multiple environments, say).  This extends to other templatization
	//     that would normally use CloudFormation's own conditionals.  It's possible we can just have a
	//     --skip-template-expansion mode that keeps the Mu templates and/or transforms them into AWS ones.

	// Emit the services.  Although services can depend on one another, the order in which we emit them here doesn't
	// matter.  The reason is that those dependencies are "runtime"-based and will get resolved elsewhere.
	for _, svc := range ast.StableServices(comp.Stack.Services.Private) {
		private := comp.Stack.Services.Private[svc]
		cf.Resources[string(private.Name)] = *c.genServiceTemplate(comp, &private)
	}
	for _, svc := range ast.StableServices(comp.Stack.Services.Public) {
		public := comp.Stack.Services.Public[svc]
		cf.Resources[string(public.Name)] = *c.genServiceTemplate(comp, &public)
	}

	// TODO: emit output exports (public services) that can be consumed by other stacks.

	return cf
}

// genServiceTemplate creates a CloudFormation resource for a single service.
func (c *awsCloud) genServiceTemplate(comp core.Compiland, svc *ast.Service) *cfResource {
	return nil
}