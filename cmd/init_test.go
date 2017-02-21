package cmd

import (
	"testing"
	"github.com/continuouspipe/remote-environment-client/test"
	"fmt"
)

func TestInitHandler_Complete(t *testing.T) {
	fmt.Println("Running TestInitHandler_Complete")
	defer fmt.Println("TestInitHandler_Complete Done")

	handler := &InitHandler{}

	//missing token
	expected := "Invalid token. Please go to continouspipe.io to obtain a valid token."
	err := handler.Complete([]string{""})
	test.AssertError(t, expected, err)

	//token present
	err = handler.Complete([]string{"some-token"})
	test.AssertNotError(t, err)
}

func TestInitHandler_Validate(t *testing.T) {
	fmt.Println("Running TestInitHandler_Validate")
	defer fmt.Println("TestInitHandler_Validate Done")

	handler := &InitHandler{}

	//Malformed token (not base64 encoded)
	malformedErrorText := "Malformed token. Please go to continouspipe.io to obtain a valid token."

	handler.Complete([]string{"some-token"})
	err := handler.Validate()
	test.AssertError(t, malformedErrorText, err)

	//Malformed token (base64 encoded but not with the expected decoded values)
	handler.Complete([]string{"dGhpcyxpcyxtYWxmb3JtZWQ="})
	err = handler.Validate()
	test.AssertError(t, malformedErrorText, err)

	//Valid token
	handler.Complete([]string{"dGhpcyxpcyxhLHZhbGlk"})
	err = handler.Validate()
	test.AssertNotError(t, err)
}
