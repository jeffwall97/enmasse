/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package util

import (
	"regexp"

	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	addressNameExpression  *regexp.Regexp = regexp.MustCompile("^[a-zA-Z]+$")
	replaceExpression      *regexp.Regexp = regexp.MustCompile("[^a-zA-Z]")
	replaceStartExpression *regexp.Regexp = regexp.MustCompile("^[^a-zA-Z]+")
	operatorUuidNamespace  uuid.UUID      = uuid.MustParse("1516b246-23aa-11e9-b615-c85b762e5a2c")
)

// Get an address name from an IoTProject
func AddressName(object metav1.Object, prefix string) string {
	return prefix + "/" + object.GetNamespace() + "." + object.GetName()
}

// Encode an address name so that it can be put inside the .metadata.name field of an Address object
func EncodeAsMetaName(addressSpaceName string, addressName string) string {

	if addressNameExpression.MatchString(addressName) {
		return addressSpaceName + "." + addressName
	}

	newPrefix := replaceExpression.ReplaceAllString(addressName, "")
	if len(newPrefix) > 0 {
		newPrefix = newPrefix + "-"
	}

	name := newPrefix + uuid.NewMD5(operatorUuidNamespace, []byte(addressName)).String()

	rname := []rune(name)
	l := len(rname)
	if l > 60 {
		s := l - 60
		rname = rname[s:l]
		name = string(rname)
		name = replaceStartExpression.ReplaceAllString(name, "")
	}

	return addressSpaceName + "." + name
}
