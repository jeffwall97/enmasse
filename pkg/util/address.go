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
	addressNameExpression *regexp.Regexp
	replaceExpression     *regexp.Regexp
	operatorUuidNamespace uuid.UUID = uuid.MustParse("1516b246-23aa-11e9-b615-c85b762e5a2c")
)

func init() {

	var err error
	addressNameExpression, err = regexp.Compile("^[a-zA-Z]+$")
	replaceExpression, err = regexp.Compile("[^a-zA-Z]")

	if err != nil {
		panic(err)
	}
}

// Get an address name from an IoTProject
func AddressName(object metav1.Object, prefix string) string {
	return prefix + "/" + object.GetNamespace() + "." + object.GetName()
}

// Encode an address name so that it can be put inside the .metadata.name field of an Address object
func EncodeAsMetaName(addressName string) string {

	if addressNameExpression.MatchString(addressName) {
		return addressName
	}

	newPrefix := replaceExpression.ReplaceAllString(addressName, "")
	if len(newPrefix) > 0 {
		newPrefix = newPrefix + "-"
	}

	return newPrefix + uuid.NewMD5(operatorUuidNamespace, []byte(addressName)).String()
}
